package seed

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/core"
	"stormlightlabs.org/baseball/internal/db"
	"stormlightlabs.org/baseball/internal/echo"
)

const (
	chadwickRegisterShardURL = "https://raw.githubusercontent.com/chadwickbureau/register/master/data/people-%s.csv"
	mlbStatsAPIBaseURL       = "https://statsapi.mlb.com/api"
)

var chadwickShardKeys = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"a", "b", "c", "d", "e", "f",
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
	shardPaths := make([]string, 0, len(chadwickShardKeys))
	for _, shard := range chadwickShardKeys {
		target := filepath.Join(dataDir, fmt.Sprintf("people-%s.csv", shard))
		url := fmt.Sprintf(chadwickRegisterShardURL, shard)
		if _, err := downloadIfNeeded(ctx, url, target, force); err != nil {
			return fmt.Errorf("failed downloading Chadwick shard %q: %w", shard, err)
		}
		shardPaths = append(shardPaths, target)
	}

	if err := mergeCSVShards(singlePath, shardPaths); err != nil {
		return fmt.Errorf("failed assembling Chadwick people.csv from shards: %w", err)
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

// LoadPlayerMLBAMMappings ingests Chadwick register IDs into player_mlbam_map.
func LoadPlayerMLBAMMappings(ctx context.Context, database *db.DB, dataDir string) (int64, error) {
	const logProgressEvery = int64(100000)
	startedAt := time.Now()

	csvPath, err := ensureChadwickRegisterCSV(ctx, dataDir)
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

	required := []string{"key_mlbam", "key_retro", "key_bbref", "name_first", "name_last"}
	for _, k := range required {
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

	if _, err := tx.ExecContext(ctx, `TRUNCATE TABLE player_mlbam_map`); err != nil {
		return 0, fmt.Errorf("failed to clear player_mlbam_map: %w", err)
	}

	buildStartedAt := time.Now()
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
		LEFT JOIN LATERAL (
			SELECT ARRAY_AGG(DISTINCT p."playerID") FILTER (WHERE p."playerID" IS NOT NULL) AS retro_matches
			FROM "People" p
			WHERE s.retro_id IS NOT NULL AND p."retroID" = s.retro_id
		) r ON TRUE
		LEFT JOIN LATERAL (
			SELECT ARRAY_AGG(DISTINCT p."playerID") FILTER (WHERE p."playerID" IS NOT NULL) AS bbref_matches
			FROM "People" p
			WHERE s.bbref_id IS NOT NULL AND p."bbrefID" = s.bbref_id
		) b ON TRUE
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to build player_mlbam_map: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit player_mlbam_map load: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		rows = staged
	}
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
func LoadTeamMLBAMMappings(ctx context.Context, database *db.DB, years []int) (int64, error) {
	if len(years) == 0 {
		years = []int{time.Now().Year()}
	}
	unique := map[int]struct{}{}
	for _, y := range years {
		if y > 0 {
			unique[y] = struct{}{}
		}
	}

	totalRows := int64(0)
	for season := range unique {
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
		totalRows += seasonRows
	}

	if err := database.RecordDatasetRefresh(ctx, "mlbam_teams_map", totalRows); err != nil {
		return totalRows, fmt.Errorf("failed to record mlbam_teams_map refresh: %w", err)
	}
	return totalRows, nil
}
