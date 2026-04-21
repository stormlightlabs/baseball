package seed

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

const (
	chadwickRegisterShardURL = "https://raw.githubusercontent.com/chadwickbureau/register/master/data/people-%s.csv"
	chadwickRegisterRepoURL  = "https://github.com/chadwickbureau/register"
	chadwickManifestFilename = "manifest.json"
	mlbStatsAPIBaseURL       = "https://statsapi.mlb.com/api"
)

var chadwickShardKeys = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"a", "b", "c", "d", "e", "f",
}

var chadwickRequiredColumns = []string{
	"key_mlbam",
	"key_retro",
	"key_bbref",
	"name_first",
	"name_last",
}

// FetchChadwickRegisterData ensures Chadwick register people.csv is available.
func FetchChadwickRegisterData(ctx context.Context, dataDir string, force bool) error {
	if dataDir == "" {
		dataDir = ChadwickDir("")
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", dataDir, err)
	}

	singlePath := filepath.Join(dataDir, "people.csv")
	if !force {
		info, err := os.Stat(singlePath)
		if err == nil && !info.IsDir() {
			return nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to check existing Chadwick people.csv: %w", err)
		}
	}

	tmpShardDir, err := os.MkdirTemp(dataDir, ".chadwick-shards-*")
	if err != nil {
		return fmt.Errorf("failed creating Chadwick shard temp directory: %w", err)
	}
	defer os.RemoveAll(tmpShardDir)

	shardPaths := make([]string, 0, len(chadwickShardKeys))
	for _, shard := range chadwickShardKeys {
		target := filepath.Join(tmpShardDir, fmt.Sprintf("people-%s.csv", shard))
		url := fmt.Sprintf(chadwickRegisterShardURL, shard)
		if _, err := downloadIfNeeded(ctx, url, target, force); err != nil {
			return fmt.Errorf("failed downloading Chadwick shard %q: %w", shard, err)
		}
		shardPaths = append(shardPaths, target)
	}

	if err := mergeCSVShards(singlePath, shardPaths); err != nil {
		return fmt.Errorf("failed assembling Chadwick people.csv from shards: %w", err)
	}
	if err := writeChadwickManifest(dataDir, singlePath, nil); err != nil {
		return fmt.Errorf("failed writing Chadwick manifest: %w", err)
	}

	return nil
}

func mergeCSVShards(outputPath string, shardPaths []string) error {
	if len(shardPaths) == 0 {
		return fmt.Errorf("no shard paths provided")
	}

	tmpPath := outputPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed creating temporary merged CSV: %w", err)
	}

	wroteHeader := false
	var firstErr error
	writer := csv.NewWriter(out)

	for _, shardPath := range shardPaths {
		file, openErr := os.Open(shardPath)
		if openErr != nil {
			firstErr = fmt.Errorf("failed opening shard %s: %w", shardPath, openErr)
			break
		}

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1

		rowNum := 0
		for {
			record, readErr := reader.Read()
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					break
				}
				firstErr = fmt.Errorf("failed reading shard %s: %w", shardPath, readErr)
				break
			}

			rowNum++
			if rowNum == 1 {
				if wroteHeader {
					continue
				}
				wroteHeader = true
			}

			if writeErr := writer.Write(record); writeErr != nil {
				firstErr = fmt.Errorf("failed writing merged CSV: %w", writeErr)
				break
			}
		}

		closeErr := file.Close()
		if firstErr != nil {
			_ = closeErr
			break
		}
		if closeErr != nil {
			firstErr = fmt.Errorf("failed closing shard %s: %w", shardPath, closeErr)
			break
		}
	}

	writer.Flush()
	if flushErr := writer.Error(); flushErr != nil && firstErr == nil {
		firstErr = fmt.Errorf("failed flushing merged CSV writer: %w", flushErr)
	}

	if closeErr := out.Close(); closeErr != nil && firstErr == nil {
		firstErr = fmt.Errorf("failed closing merged CSV: %w", closeErr)
	}

	if firstErr != nil {
		_ = os.Remove(tmpPath)
		return firstErr
	}

	if err := os.Rename(tmpPath, outputPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed promoting merged CSV to %s: %w", outputPath, err)
	}

	return nil
}

type chadwickManifest struct {
	SchemaVersion   int                    `json:"schema_version"`
	GeneratedAtUTC  string                 `json:"generated_at_utc"`
	SourceRepoURL   string                 `json:"source_repo_url"`
	SourceShardURL  string                 `json:"source_shard_url_template"`
	ShardKeys       []string               `json:"shard_keys"`
	RequiredColumns []string               `json:"required_columns"`
	Files           []chadwickManifestFile `json:"files"`
}

type chadwickManifestFile struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

func writeChadwickManifest(dataDir, mergedPath string, shardPaths []string) error {
	paths := append([]string{}, shardPaths...)
	paths = append(paths, mergedPath)
	slices.Sort(paths)

	files := make([]chadwickManifestFile, 0, len(paths))
	for _, path := range paths {
		record, err := buildChadwickManifestFile(dataDir, path)
		if err != nil {
			return err
		}
		files = append(files, record)
	}

	manifest := chadwickManifest{
		SchemaVersion:   1,
		GeneratedAtUTC:  time.Now().UTC().Format(time.RFC3339),
		SourceRepoURL:   chadwickRegisterRepoURL,
		SourceShardURL:  chadwickRegisterShardURL,
		ShardKeys:       slices.Clone(chadwickShardKeys),
		RequiredColumns: slices.Clone(chadwickRequiredColumns),
		Files:           files,
	}

	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest JSON: %w", err)
	}
	content = append(content, '\n')

	manifestPath := filepath.Join(dataDir, chadwickManifestFilename)
	tmpPath := manifestPath + ".tmp"
	if err := os.WriteFile(tmpPath, content, 0644); err != nil {
		return fmt.Errorf("failed writing manifest temp file: %w", err)
	}
	if err := os.Rename(tmpPath, manifestPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed promoting manifest file: %w", err)
	}
	return nil
}

func buildChadwickManifestFile(rootDir, path string) (chadwickManifestFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return chadwickManifestFile{}, fmt.Errorf("failed to stat %s: %w", path, err)
	}
	hash, err := hashFileSHA256(path)
	if err != nil {
		return chadwickManifestFile{}, err
	}

	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return chadwickManifestFile{}, fmt.Errorf("failed to build relative path for %s: %w", path, err)
	}

	return chadwickManifestFile{
		Path:      filepath.ToSlash(rel),
		SizeBytes: info.Size(),
		SHA256:    hash,
	}, nil
}

func hashFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open %s for checksum: %w", path, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash %s: %w", path, err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func ensureChadwickRegisterCSV(ctx context.Context, dataDir string) (string, error) {
	if dataDir == "" {
		dataDir = ChadwickDir("")
	}
	csvPath := filepath.Join(dataDir, "people.csv")
	if _, err := os.Stat(csvPath); err == nil {
		return csvPath, nil
	}
	if err := FetchChadwickRegisterData(ctx, dataDir, false); err != nil {
		return "", err
	}
	if _, err := os.Stat(csvPath); err != nil {
		return "", fmt.Errorf("chadwick register CSV not available at %s", csvPath)
	}
	return csvPath, nil
}

// PlayerMLBAMMappingOptions controls player crosswalk ingestion behavior.
type PlayerMLBAMMappingOptions struct {
	DataDir string
	Skip    bool
}

// LoadPlayerMLBAMMappings ingests Chadwick register IDs into player_mlbam_map.
func LoadPlayerMLBAMMappings(ctx context.Context, database *db.DB, opts PlayerMLBAMMappingOptions) (int64, error) {
	const logProgressEvery = int64(100000)
	startedAt := time.Now()

	if opts.Skip {
		loaded, err := database.IsDatasetLoaded(ctx, "mlbam_players_map")
		if err != nil {
			return 0, fmt.Errorf("failed to check if mlbam_players_map is loaded: %w", err)
		}
		if loaded {
			echo.Info("Player MLBAM crosswalk already loaded, skipping")
			return 0, nil
		}
	}

	csvPath, err := ensureChadwickRegisterCSV(ctx, opts.DataDir)
	if err != nil {
		return 0, err
	}
	fileInfo, err := os.Stat(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to stat chadwick register CSV: %w", err)
	}
	echo.Infof(
		"Loading player MLBAM crosswalk from %s (%s); this can take a while for full-history registers.",
		csvPath,
		formatByteSize(fileInfo.Size()),
	)

	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open chadwick register CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read chadwick headers: %w", err)
	}
	index := map[string]int{}
	for i, h := range headers {
		index[strings.TrimSpace(strings.ToLower(h))] = i
	}

	for _, k := range chadwickRequiredColumns {
		if _, ok := index[k]; !ok {
			return 0, fmt.Errorf("chadwick CSV missing required column %q", k)
		}
	}

	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		CREATE TEMP TABLE chadwick_player_stage (
			mlbam_id INTEGER PRIMARY KEY,
			retro_id VARCHAR(8),
			bbref_id VARCHAR(16),
			full_name TEXT
		) ON COMMIT DROP
	`); err != nil {
		return 0, fmt.Errorf("failed to create chadwick staging table: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO chadwick_player_stage (mlbam_id, retro_id, bbref_id, full_name)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''))
		ON CONFLICT (mlbam_id) DO UPDATE SET
			retro_id = EXCLUDED.retro_id,
			bbref_id = EXCLUDED.bbref_id,
			full_name = EXCLUDED.full_name
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare chadwick stage insert: %w", err)
	}
	defer stmt.Close()

	parseStartedAt := time.Now()
	var processed int64
	var staged int64
	var skippedMissingMLBAM int64
	var skippedInvalidMLBAM int64
	for {
		rec, readErr := reader.Read()
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			return 0, fmt.Errorf("failed reading chadwick register CSV: %w", readErr)
		}
		if rec == nil {
			break
		}
		processed++

		mlbamRaw := strings.TrimSpace(rec[index["key_mlbam"]])
		if mlbamRaw == "" {
			skippedMissingMLBAM++
			continue
		}
		mlbamID, parseErr := strconv.Atoi(mlbamRaw)
		if parseErr != nil || mlbamID <= 0 {
			skippedInvalidMLBAM++
			continue
		}

		retroID := strings.TrimSpace(rec[index["key_retro"]])
		bbrefID := strings.TrimSpace(rec[index["key_bbref"]])
		fullName := strings.TrimSpace(strings.Join([]string{rec[index["name_first"]], rec[index["name_last"]]}, " "))

		if _, err := stmt.ExecContext(ctx, mlbamID, retroID, bbrefID, fullName); err != nil {
			return 0, fmt.Errorf("failed to stage chadwick row: %w", err)
		}
		staged++

		if processed%logProgressEvery == 0 {
			echo.Infof(
				"  Chadwick parse progress: processed=%s staged=%s skipped=%s elapsed=%s",
				formatNumber(processed),
				formatNumber(staged),
				formatNumber(skippedMissingMLBAM+skippedInvalidMLBAM),
				time.Since(parseStartedAt).Round(time.Second),
			)
		}
	}
	echo.Infof(
		"  Chadwick parse complete: processed=%s staged=%s skipped_missing_mlbam=%s skipped_invalid_mlbam=%s (%s)",
		formatNumber(processed),
		formatNumber(staged),
		formatNumber(skippedMissingMLBAM),
		formatNumber(skippedInvalidMLBAM),
		time.Since(parseStartedAt).Round(time.Second),
	)
	recordETLPhaseEvent(
		ctx,
		database,
		"load.crosswalk.players_mlbam",
		"parse_stage",
		"completed",
		staged,
		parseStartedAt,
		map[string]any{
			"processed":              processed,
			"staged":                 staged,
			"skipped_missing_mlbam":  skippedMissingMLBAM,
			"skipped_invalid_mlbam":  skippedInvalidMLBAM,
			"source_file":            csvPath,
			"source_file_size_bytes": fileInfo.Size(),
		},
		nil,
	)

	clearStartedAt := time.Now()
	echo.Info("  Crosswalk phase: clear target table (player_mlbam_map)")
	clearResult, err := tx.ExecContext(ctx, `DELETE FROM player_mlbam_map`)
	if err != nil {
		recordETLPhaseEvent(
			ctx,
			database,
			"load.crosswalk.players_mlbam",
			"clear_target",
			"failed",
			0,
			clearStartedAt,
			map[string]any{"target_table": "player_mlbam_map"},
			err,
		)
		return 0, fmt.Errorf("failed to clear player_mlbam_map: %w", err)
	}
	clearedRows, _ := clearResult.RowsAffected()
	recordETLPhaseEvent(
		ctx,
		database,
		"load.crosswalk.players_mlbam",
		"clear_target",
		"completed",
		clearedRows,
		clearStartedAt,
		map[string]any{"target_table": "player_mlbam_map"},
		nil,
	)
	echo.Infof("  Cleared player_mlbam_map rows=%s (%s)", formatNumber(clearedRows), time.Since(clearStartedAt).Round(time.Millisecond))

	buildStartedAt := time.Now()
	echo.Info("  Crosswalk phase: build People lookup tables")
	var peopleRows int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM "People"`).Scan(&peopleRows); err == nil {
		echo.Infof("  Building player_mlbam_map from staged=%s against People=%s...", formatNumber(staged), formatNumber(peopleRows))
	}
	if _, err := tx.ExecContext(ctx, `
		CREATE TEMP TABLE chadwick_retro_lookup ON COMMIT DROP AS
		SELECT
			p."retroID" AS retro_id,
			ARRAY_AGG(DISTINCT p."playerID") FILTER (WHERE p."playerID" IS NOT NULL) AS retro_matches
		FROM "People" p
		WHERE p."retroID" IS NOT NULL
		GROUP BY p."retroID";
		CREATE INDEX chadwick_retro_lookup_retro_id_idx ON chadwick_retro_lookup(retro_id);

		CREATE TEMP TABLE chadwick_bbref_lookup ON COMMIT DROP AS
		SELECT
			p."bbrefID" AS bbref_id,
			ARRAY_AGG(DISTINCT p."playerID") FILTER (WHERE p."playerID" IS NOT NULL) AS bbref_matches
		FROM "People" p
		WHERE p."bbrefID" IS NOT NULL
		GROUP BY p."bbrefID";
		CREATE INDEX chadwick_bbref_lookup_bbref_id_idx ON chadwick_bbref_lookup(bbref_id);
	`); err != nil {
		recordETLPhaseEvent(
			ctx,
			database,
			"load.crosswalk.players_mlbam",
			"build_lookup",
			"failed",
			0,
			buildStartedAt,
			map[string]any{
				"staged_rows": staged,
				"people_rows": peopleRows,
			},
			err,
		)
		return 0, fmt.Errorf("failed to build player crosswalk lookups: %w", err)
	}
	recordETLPhaseEvent(
		ctx,
		database,
		"load.crosswalk.players_mlbam",
		"build_lookup",
		"completed",
		staged,
		buildStartedAt,
		map[string]any{
			"staged_rows": staged,
			"people_rows": peopleRows,
		},
		nil,
	)
	echo.Infof("  Built lookup tables (%s)", time.Since(buildStartedAt).Round(time.Millisecond))

	insertStartedAt := time.Now()
	echo.Info("  Crosswalk phase: insert merged mappings")
	result, err := tx.ExecContext(ctx, `
		INSERT INTO player_mlbam_map (
			mlbam_id,
			lahman_id,
			retro_id,
			bbref_id,
			full_name,
			source,
			confidence,
			updated_at
		)
		SELECT
			s.mlbam_id,
			CASE
				WHEN COALESCE(cardinality(r.retro_matches), 0) > 1 THEN NULL
				WHEN COALESCE(cardinality(b.bbref_matches), 0) > 1 THEN NULL
				WHEN COALESCE(cardinality(r.retro_matches), 0) = 1
					AND COALESCE(cardinality(b.bbref_matches), 0) = 1
					AND r.retro_matches[1] <> b.bbref_matches[1] THEN NULL
				ELSE COALESCE(r.retro_matches[1], b.bbref_matches[1])
			END AS lahman_id,
			s.retro_id,
			s.bbref_id,
			s.full_name,
				'chadwick',
				CASE
					WHEN COALESCE(cardinality(r.retro_matches), 0) > 1 THEN 'none'
					WHEN COALESCE(cardinality(b.bbref_matches), 0) > 1 THEN 'none'
				WHEN COALESCE(cardinality(r.retro_matches), 0) = 1
					AND COALESCE(cardinality(b.bbref_matches), 0) = 1
					AND r.retro_matches[1] <> b.bbref_matches[1] THEN 'none'
				WHEN COALESCE(r.retro_matches[1], b.bbref_matches[1]) IS NOT NULL THEN 'high'
				ELSE 'none'
				END,
				NOW()
			FROM chadwick_player_stage s
			LEFT JOIN chadwick_retro_lookup r ON r.retro_id = s.retro_id
			LEFT JOIN chadwick_bbref_lookup b ON b.bbref_id = s.bbref_id
	`)
	if err != nil {
		recordETLPhaseEvent(
			ctx,
			database,
			"load.crosswalk.players_mlbam",
			"insert",
			"failed",
			0,
			insertStartedAt,
			map[string]any{"staged_rows": staged},
			err,
		)
		return 0, fmt.Errorf("failed to build player_mlbam_map: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		rows = staged
	}
	recordETLPhaseEvent(
		ctx,
		database,
		"load.crosswalk.players_mlbam",
		"insert",
		"completed",
		rows,
		insertStartedAt,
		map[string]any{"staged_rows": staged},
		nil,
	)
	echo.Infof("  Inserted player_mlbam_map rows=%s (%s)", formatNumber(rows), time.Since(insertStartedAt).Round(time.Millisecond))

	commitStartedAt := time.Now()
	if err := tx.Commit(); err != nil {
		recordETLPhaseEvent(
			ctx,
			database,
			"load.crosswalk.players_mlbam",
			"commit",
			"failed",
			0,
			commitStartedAt,
			map[string]any{
				"inserted_rows": rows,
			},
			err,
		)
		return 0, fmt.Errorf("failed to commit player_mlbam_map load: %w", err)
	}
	recordETLPhaseEvent(
		ctx,
		database,
		"load.crosswalk.players_mlbam",
		"commit",
		"completed",
		rows,
		commitStartedAt,
		map[string]any{
			"inserted_rows": rows,
		},
		nil,
	)
	echo.Infof("  Crosswalk commit completed in %s", time.Since(commitStartedAt).Round(time.Millisecond))

	echo.Infof(
		"  Built player_mlbam_map rows=%s from staged=%s (%s)",
		formatNumber(rows),
		formatNumber(staged),
		time.Since(buildStartedAt).Round(time.Second),
	)
	echo.Infof("  Player MLBAM crosswalk step completed in %s", time.Since(startedAt).Round(time.Second))

	if err := database.RecordDatasetRefresh(ctx, "mlbam_players_map", rows); err != nil {
		return rows, fmt.Errorf("failed to record mlbam_players_map refresh: %w", err)
	}
	return rows, nil
}

// TeamMLBAMMappingOptions controls team crosswalk ingestion behavior.
type TeamMLBAMMappingOptions struct {
	Years []int
	Skip  bool
}

func formatByteSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(size)
	unitIdx := -1
	for value >= 1024 && unitIdx < len(units)-1 {
		value /= 1024
		unitIdx++
	}
	return fmt.Sprintf("%.1f %s", value, units[unitIdx])
}

type teamMatchCandidate struct {
	TeamID      core.TeamID
	FranchiseID core.FranchiseID
	Name        string
	League      core.LeagueID
}

type localTeamIndexes struct {
	byTeamID      map[string]teamMatchCandidate
	byFranchiseID map[string]teamMatchCandidate
	byName        map[string]teamMatchCandidate
}

var mlbCodeCandidates = map[string][]string{
	"ARI": {"ARI"},
	"ATL": {"ATL"},
	"BAL": {"BAL"},
	"BOS": {"BOS"},
	"CHC": {"CHN", "CHC"},
	"CWS": {"CHA", "CHW"},
	"CIN": {"CIN"},
	"CLE": {"CLE"},
	"COL": {"COL"},
	"DET": {"DET"},
	"HOU": {"HOU"},
	"KC":  {"KCA", "KCR"},
	"LAA": {"LAA", "ANA"},
	"LAD": {"LAN", "LAD"},
	"MIA": {"MIA", "FLO"},
	"MIL": {"MIL"},
	"MIN": {"MIN"},
	"NYM": {"NYN", "NYM"},
	"NYY": {"NYA", "NYY"},
	"ATH": {"OAK", "ATH"},
	"PHI": {"PHI"},
	"PIT": {"PIT"},
	"SD":  {"SDN", "SDP"},
	"SEA": {"SEA"},
	"SF":  {"SFN", "SFG"},
	"STL": {"SLN", "STL"},
	"TB":  {"TBA", "TBD", "TBR"},
	"TEX": {"TEX"},
	"TOR": {"TOR"},
	"WSH": {"WAS", "WSN"},
}

func normalizeLookupKey(value string) string {
	upper := strings.ToUpper(strings.TrimSpace(value))
	if upper == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range upper {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func buildLocalTeamIndexes(rows []teamMatchCandidate) localTeamIndexes {
	indexes := localTeamIndexes{
		byTeamID:      make(map[string]teamMatchCandidate, len(rows)),
		byFranchiseID: make(map[string]teamMatchCandidate, len(rows)),
		byName:        make(map[string]teamMatchCandidate, len(rows)),
	}
	for _, row := range rows {
		if k := normalizeLookupKey(string(row.TeamID)); k != "" {
			indexes.byTeamID[k] = row
		}
		if k := normalizeLookupKey(string(row.FranchiseID)); k != "" {
			indexes.byFranchiseID[k] = row
		}
		if k := normalizeLookupKey(row.Name); k != "" {
			indexes.byName[k] = row
		}
	}
	return indexes
}

func localCandidatesForMLBTeam(team core.MLBTeam) (codes []string, names []string) {
	seenCode := map[string]struct{}{}
	addCode := func(v string) {
		k := normalizeLookupKey(v)
		if k == "" {
			return
		}
		if _, ok := seenCode[k]; ok {
			return
		}
		seenCode[k] = struct{}{}
		codes = append(codes, k)
	}

	seenName := map[string]struct{}{}
	addName := func(v string) {
		k := normalizeLookupKey(v)
		if k == "" {
			return
		}
		if _, ok := seenName[k]; ok {
			return
		}
		seenName[k] = struct{}{}
		names = append(names, k)
	}

	addCode(team.Abbreviation)
	addCode(team.TeamCode)
	addCode(team.FileCode)
	for _, c := range mlbCodeCandidates[normalizeLookupKey(team.Abbreviation)] {
		addCode(c)
	}

	addName(team.Name)
	addName(team.FranchiseName)
	addName(team.ClubName)
	addName(team.TeamName)
	addName(team.ShortName)
	if team.LocationName != "" && team.TeamName != "" {
		addName(team.LocationName + " " + team.TeamName)
	}
	return codes, names
}

func listLocalTeamsForSeason(ctx context.Context, database *db.DB, season int) ([]teamMatchCandidate, error) {
	rows, err := database.QueryContext(ctx, `
		SELECT DISTINCT "teamID", "franchID", name, "lgID"
		FROM "Teams"
		WHERE "yearID" = $1
	`, season)
	if err != nil {
		return nil, fmt.Errorf("failed to list local teams for season %d: %w", season, err)
	}
	defer rows.Close()

	out := []teamMatchCandidate{}
	for rows.Next() {
		var row teamMatchCandidate
		if err := rows.Scan(&row.TeamID, &row.FranchiseID, &row.Name, &row.League); err != nil {
			return nil, fmt.Errorf("failed to scan local team row: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func fetchMLBTeamsForSeason(ctx context.Context, season int) (core.MLBTeamsResponse, error) {
	target, err := url.JoinPath(mlbStatsAPIBaseURL, "v1", "teams")
	if err != nil {
		return core.MLBTeamsResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return core.MLBTeamsResponse{}, err
	}
	q := req.URL.Query()
	q.Set("sportId", "1")
	q.Set("season", strconv.Itoa(season))
	req.URL.RawQuery = q.Encode()
	req.Header.Set("User-Agent", "Stormlight-Baseball-ETL/1.0")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return core.MLBTeamsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return core.MLBTeamsResponse{}, fmt.Errorf("mlb teams request failed with %d", resp.StatusCode)
	}

	var payload core.MLBTeamsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return core.MLBTeamsResponse{}, fmt.Errorf("failed to decode MLB teams payload: %w", err)
	}
	return payload, nil
}

// LoadTeamMLBAMMappings builds season-scoped team MLBAM mappings for the selected ETL years.
func LoadTeamMLBAMMappings(ctx context.Context, database *db.DB, opts TeamMLBAMMappingOptions) (int64, error) {
	years := dedupeSortedYears(opts.Years)
	if len(years) == 0 {
		years = []int{time.Now().Year()}
	}

	totalRows := int64(0)
	loadedSeasons := 0
	for _, season := range years {
		if opts.Skip {
			seasonKey := fmt.Sprintf("mlbam_teams_map_%d", season)
			loaded, err := database.IsDatasetLoaded(ctx, seasonKey)
			if err != nil {
				return totalRows, fmt.Errorf("failed to check if %s is loaded: %w", seasonKey, err)
			}
			if loaded {
				echo.Infof("Team MLBAM crosswalk already loaded for season %d, skipping", season)
				continue
			}
		}

		mlbTeams, err := fetchMLBTeamsForSeason(ctx, season)
		if err != nil {
			return totalRows, fmt.Errorf("failed to fetch MLB teams for %d: %w", season, err)
		}
		localTeams, err := listLocalTeamsForSeason(ctx, database, season)
		if err != nil {
			return totalRows, err
		}
		index := buildLocalTeamIndexes(localTeams)

		tx, err := database.BeginTx(ctx, nil)
		if err != nil {
			return totalRows, fmt.Errorf("failed to begin tx for season %d: %w", season, err)
		}

		if _, err := tx.ExecContext(ctx, `DELETE FROM team_mlbam_map WHERE season = $1`, season); err != nil {
			tx.Rollback()
			return totalRows, fmt.Errorf("failed clearing team_mlbam_map for season %d: %w", season, err)
		}

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO team_mlbam_map (
				season,
				mlbam_team_id,
				mlb_abbreviation,
				mlb_team_code,
				mlb_file_code,
				mlb_team_name,
				mlb_franchise_name,
				mlb_club_name,
				local_team_id,
				local_franchise_id,
				local_team_name,
				local_league,
				match_method,
				confidence,
				source,
				updated_at
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'mlb-api',NOW()
			)
		`)
		if err != nil {
			tx.Rollback()
			return totalRows, fmt.Errorf("failed to prepare team_mlbam_map insert: %w", err)
		}

		seasonRows := int64(0)
		for _, mlbTeam := range mlbTeams.Teams {
			var matched *teamMatchCandidate
			matchMethod := ""
			confidence := "none"

			codes, names := localCandidatesForMLBTeam(mlbTeam)
			for _, code := range codes {
				if row, ok := index.byTeamID[code]; ok {
					r := row
					matched = &r
					matchMethod = "team_id"
					confidence = "high"
					break
				}
				if row, ok := index.byFranchiseID[code]; ok {
					r := row
					matched = &r
					matchMethod = "franchise_id"
					confidence = "high"
					break
				}
			}
			if matched == nil {
				for _, name := range names {
					if row, ok := index.byName[name]; ok {
						r := row
						matched = &r
						matchMethod = "name"
						confidence = "medium"
						break
					}
				}
			}

			var localTeamID any
			var localFranchiseID any
			var localTeamName any
			var localLeague any
			if matched != nil {
				localTeamID = string(matched.TeamID)
				localFranchiseID = string(matched.FranchiseID)
				localTeamName = matched.Name
				localLeague = string(matched.League)
			}

			if _, err := stmt.ExecContext(
				ctx,
				season,
				mlbTeam.ID,
				mlbTeam.Abbreviation,
				mlbTeam.TeamCode,
				mlbTeam.FileCode,
				mlbTeam.Name,
				mlbTeam.FranchiseName,
				mlbTeam.ClubName,
				localTeamID,
				localFranchiseID,
				localTeamName,
				localLeague,
				matchMethod,
				confidence,
			); err != nil {
				stmt.Close()
				tx.Rollback()
				return totalRows, fmt.Errorf("failed inserting team_mlbam_map row: %w", err)
			}
			seasonRows++
		}

		stmt.Close()
		if err := tx.Commit(); err != nil {
			return totalRows, fmt.Errorf("failed committing team_mlbam_map season %d: %w", season, err)
		}
		seasonKey := fmt.Sprintf("mlbam_teams_map_%d", season)
		if err := database.RecordDatasetRefresh(ctx, seasonKey, seasonRows); err != nil {
			return totalRows, fmt.Errorf("failed to record %s refresh: %w", seasonKey, err)
		}

		totalRows += seasonRows
		loadedSeasons++
	}

	if loadedSeasons > 0 {
		if err := database.RecordDatasetRefresh(ctx, "mlbam_teams_map", totalRows); err != nil {
			return totalRows, fmt.Errorf("failed to record mlbam_teams_map refresh: %w", err)
		}
	}
	return totalRows, nil
}
