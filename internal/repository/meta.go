package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

// MetaRepository implements core.MetaRepository backed by PostgreSQL.
type MetaRepository struct {
	db *sql.DB
}

func NewMetaRepository(db *sql.DB) *MetaRepository {
	return &MetaRepository{db: db}
}

func (r *MetaRepository) SeasonCoverage(ctx context.Context) (core.SeasonYear, core.SeasonYear, core.SeasonYear, core.SeasonYear, error) {
	var minLahman, maxLahman sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT MIN("yearID"), MAX("yearID") FROM "Teams"`).Scan(&minLahman, &maxLahman); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to query Lahman season coverage: %w", err)
	}

	var minRetro, maxRetro sql.NullString
	if err := r.db.QueryRowContext(ctx, `SELECT MIN(date), MAX(date) FROM games`).Scan(&minRetro, &maxRetro); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to query Retrosheet coverage: %w", err)
	}

	return toSeasonYear(minLahman), toSeasonYear(maxLahman), seasonFromDate(minRetro), seasonFromDate(maxRetro), nil
}

func (r *MetaRepository) LastUpdated(ctx context.Context) (time.Time, time.Time, error) {
	refreshes, err := r.loadRefreshes(ctx)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	var lahman time.Time
	if entry, ok := refreshes["lahman"]; ok {
		lahman = entry.lastLoaded
	}

	var retro time.Time
	for _, key := range []string{"retrosheet_games", "retrosheet_plays"} {
		if entry, ok := refreshes[key]; ok {
			if entry.lastLoaded.After(retro) {
				retro = entry.lastLoaded
			}
		}
	}

	return lahman, retro, nil
}

func (r *MetaRepository) DatasetStatuses(ctx context.Context) ([]core.DatasetStatus, error) {
	refreshes, err := r.loadRefreshes(ctx)
	if err != nil {
		return nil, err
	}

	return r.buildDatasetStatuses(ctx, refreshes)
}

func (r *MetaRepository) Readiness(ctx context.Context) (core.ReadinessStatus, error) {
	refreshes, err := r.loadRefreshes(ctx)
	if err != nil {
		return core.ReadinessStatus{}, err
	}

	statuses, err := r.buildDatasetStatuses(ctx, refreshes)
	if err != nil {
		return core.ReadinessStatus{}, err
	}

	required := make([]core.DatasetStatus, 0, len(statuses))
	missing := make([]string, 0)
	ready := true

	for _, status := range statuses {
		if !status.Required {
			continue
		}
		required = append(required, status)
		if !status.Healthy {
			ready = false
			missing = append(missing, status.ID)
		}
	}

	state := "ready"
	if !ready {
		state = "degraded"
	}

	return core.ReadinessStatus{
		Status:    state,
		Ready:     ready,
		CheckedAt: time.Now().UTC(),
		Datasets:  required,
		Missing:   missing,
	}, nil
}

func (r *MetaRepository) buildDatasetStatuses(ctx context.Context, refreshes map[string]refreshRecord) ([]core.DatasetStatus, error) {
	minLahman, maxLahman, minRetro, maxRetro, err := r.SeasonCoverage(ctx)
	if err != nil {
		return nil, err
	}

	var minSalary, maxSalary sql.NullInt64
	_ = r.db.QueryRowContext(ctx, `SELECT MIN("yearID"), MAX("yearID") FROM "Salaries"`).Scan(&minSalary, &maxSalary)

	return []core.DatasetStatus{
		r.datasetStatus(
			"lahman",
			"Lahman Baseball Database",
			"https://sabr.org/lahman-database/",
			true,
			map[string]int64{
				"people":      r.tableCount(ctx, "People", `SELECT COUNT(*) FROM "People"`),
				"teams":       r.tableCount(ctx, "Teams", `SELECT COUNT(*) FROM "Teams"`),
				"batting":     r.tableCount(ctx, "Batting", `SELECT COUNT(*) FROM "Batting"`),
				"pitching":    r.tableCount(ctx, "Pitching", `SELECT COUNT(*) FROM "Pitching"`),
				"appearances": r.tableCount(ctx, "Appearances", `SELECT COUNT(*) FROM "Appearances"`),
				"salaries":    r.tableCount(ctx, "Salaries", `SELECT COUNT(*) FROM "Salaries"`),
			},
			seasonPtr(minLahman),
			seasonPtr(maxLahman),
			refreshes,
			[]string{"lahman"},
		),
		r.datasetStatus(
			"retrosheet",
			"Retrosheet Game Logs & Plays",
			"https://www.retrosheet.org/",
			true,
			map[string]int64{
				"games": r.tableCount(ctx, "games", `SELECT COUNT(*) FROM games`),
				"plays": r.tableCount(ctx, "plays", `SELECT COUNT(*) FROM plays`),
			},
			seasonPtr(minRetro),
			seasonPtr(maxRetro),
			refreshes,
			[]string{"retrosheet_games", "retrosheet_plays"},
		),
		r.datasetStatus(
			"fangraphs_constants",
			"FanGraphs constants",
			"https://www.fangraphs.com/tools/guts",
			true,
			map[string]int64{
				"woba_constants":   r.tableCount(ctx, "woba_constants", `SELECT COUNT(*) FROM woba_constants`),
				"league_constants": r.tableCount(ctx, "league_constants", `SELECT COUNT(*) FROM league_constants`),
				"park_factors":     r.tableCount(ctx, "park_factors", `SELECT COUNT(*) FROM park_factors`),
			},
			nil,
			nil,
			refreshes,
			[]string{"fangraphs_constants"},
		),
		r.datasetStatus(
			"retrosheet_players",
			"Retrosheet player crosswalk",
			"https://www.retrosheet.org/downloads/csvdownloads.html",
			false,
			map[string]int64{
				"retrosheet_players": r.tableCount(ctx, "retrosheet_players", `SELECT COUNT(*) FROM retrosheet_players`),
			},
			nil,
			nil,
			refreshes,
			[]string{"retrosheet_players"},
		),
		r.datasetStatus(
			"biodata",
			"Retrosheet biodata",
			"https://www.retrosheet.org/downloads/csvdownloads.html",
			false,
			map[string]int64{
				"player_bio_extended": r.tableCount(ctx, "player_bio_extended", `SELECT COUNT(*) FROM player_bio_extended`),
				"player_relatives":    r.tableCount(ctx, "player_relatives", `SELECT COUNT(*) FROM player_relatives`),
				"coaches":             r.tableCount(ctx, "coaches", `SELECT COUNT(*) FROM coaches`),
				"umpires":             r.tableCount(ctx, "umpires", `SELECT COUNT(*) FROM umpires`),
			},
			nil,
			nil,
			refreshes,
			[]string{"biodata"},
		),
		r.datasetStatus(
			"salary_summary",
			"Salary summary",
			"https://legacy.baseballprospectus.com/compensation/cots/",
			false,
			map[string]int64{
				"salary_summary": r.tableCount(ctx, "salary_summary", `SELECT COUNT(*) FROM salary_summary`),
			},
			seasonPtr(toSeasonYear(minSalary)),
			seasonPtr(toSeasonYear(maxSalary)),
			refreshes,
			[]string{"salaries"},
		),
		r.datasetStatus(
			"mlbam_players_map",
			"MLBAM player crosswalk",
			"https://github.com/chadwickbureau/register",
			false,
			map[string]int64{
				"player_mlbam_map": r.tableCount(ctx, "player_mlbam_map", `SELECT COUNT(*) FROM player_mlbam_map`),
			},
			nil,
			nil,
			refreshes,
			[]string{"mlbam_players_map"},
		),
		r.datasetStatus(
			"mlbam_teams_map",
			"MLBAM team crosswalk",
			"https://statsapi.mlb.com/api/v1/teams",
			false,
			map[string]int64{
				"team_mlbam_map": r.tableCount(ctx, "team_mlbam_map", `SELECT COUNT(*) FROM team_mlbam_map`),
			},
			nil,
			nil,
			refreshes,
			[]string{"mlbam_teams_map"},
		),
	}, nil
}

func (r *MetaRepository) datasetStatus(id, name, source string, required bool, tables map[string]int64, coverageFrom, coverageTo *core.SeasonYear, refreshes map[string]refreshRecord, refreshKeys []string) core.DatasetStatus {
	rowCount := sumCounts(tables)
	if recorded := sumRefreshRows(refreshes, refreshKeys); recorded > 0 {
		rowCount = recorded
	}

	status := core.DatasetStatus{
		ID:           id,
		Name:         name,
		Source:       source,
		Required:     required,
		Healthy:      datasetHealthy(id, tables),
		CoverageFrom: coverageFrom,
		CoverageTo:   coverageTo,
		RowCount:     rowCount,
		Tables:       tables,
	}

	if loaded := latestRefresh(refreshes, refreshKeys); !loaded.IsZero() {
		ts := loaded
		status.LastLoadedAt = &ts
	}

	return status
}

func datasetHealthy(id string, tables map[string]int64) bool {
	switch id {
	case "lahman":
		return tables["people"] > 0 && tables["teams"] > 0 && tables["batting"] > 0 && tables["pitching"] > 0
	case "retrosheet":
		return tables["games"] > 0 && tables["plays"] > 0
	case "fangraphs_constants":
		return tables["woba_constants"] > 0 && tables["league_constants"] > 0 && tables["park_factors"] > 0
	case "retrosheet_players":
		return tables["retrosheet_players"] > 0
	case "biodata":
		return tables["player_bio_extended"] > 0 || tables["player_relatives"] > 0 || tables["coaches"] > 0 || tables["umpires"] > 0
	case "salary_summary":
		return tables["salary_summary"] > 0
	case "mlbam_players_map":
		return tables["player_mlbam_map"] > 0
	case "mlbam_teams_map":
		return tables["team_mlbam_map"] > 0
	default:
		return sumCounts(tables) > 0
	}
}

func sumCounts(values map[string]int64) int64 {
	var total int64
	for _, count := range values {
		total += count
	}
	return total
}

func latestRefresh(refreshes map[string]refreshRecord, keys []string) time.Time {
	var latest time.Time
	for _, key := range keys {
		if entry, ok := refreshes[key]; ok && entry.lastLoaded.After(latest) {
			latest = entry.lastLoaded
		}
	}
	return latest
}

func sumRefreshRows(refreshes map[string]refreshRecord, keys []string) int64 {
	var total int64
	for _, key := range keys {
		if entry, ok := refreshes[key]; ok {
			total += entry.rowCount
		}
	}
	return total
}

func (r *MetaRepository) SchemaHashes(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM schema_migrations ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema migrations: %w", err)
	}
	defer rows.Close()

	hashers := map[string]*hashAccumulator{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan migration name: %w", err)
		}
		bucket := classifyMigration(name)
		if _, ok := hashers[bucket]; !ok {
			hashers[bucket] = newAccumulator()
		}
		hashers[bucket].write(name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate migrations: %w", err)
	}

	result := make(map[string]string, len(hashers))
	for bucket, hasher := range hashers {
		result[bucket] = hasher.sum()
	}

	return result, nil
}

type refreshRecord struct {
	lastLoaded time.Time
	rowCount   int64
}

func (r *MetaRepository) loadRefreshes(ctx context.Context) (map[string]refreshRecord, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT dataset, last_loaded_at, row_count FROM dataset_refreshes`)
	if err != nil {
		if strings.Contains(err.Error(), "dataset_refreshes") {
			return map[string]refreshRecord{}, nil
		}
		return nil, fmt.Errorf("failed to read dataset refresh metadata: %w", err)
	}
	defer rows.Close()

	result := map[string]refreshRecord{}
	for rows.Next() {
		var dataset string
		var loaded time.Time
		var rowsCount sql.NullInt64
		if err := rows.Scan(&dataset, &loaded, &rowsCount); err != nil {
			return nil, fmt.Errorf("failed to scan refresh metadata: %w", err)
		}
		entry := refreshRecord{
			lastLoaded: loaded,
		}
		if rowsCount.Valid {
			entry.rowCount = rowsCount.Int64
		}
		result[dataset] = entry
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate refresh metadata: %w", err)
	}
	return result, nil
}

func (r *MetaRepository) safeCount(ctx context.Context, query string) (int64, error) {
	var count sql.NullInt64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, err
	}
	if !count.Valid {
		return 0, nil
	}
	return count.Int64, nil
}

func (r *MetaRepository) tableCount(ctx context.Context, tableName, strictQuery string) int64 {
	if core.CountModeFromContext(ctx) == core.CountModeStrict {
		if count, err := r.safeCount(ctx, strictQuery); err == nil {
			return count
		}
		core.MarkCountModeFallback(ctx)
	}

	if estimate, ok := r.estimateTableRows(ctx, tableName); ok {
		if estimate <= 0 {
			if hasRows, existsOK := r.tableHasRows(ctx, tableName); existsOK && hasRows {
				return 1
			}
		}
		return estimate
	}

	if hasRows, ok := r.tableHasRows(ctx, tableName); ok {
		if hasRows {
			return 1
		}
		return 0
	}

	return 0
}

func (r *MetaRepository) estimateTableRows(ctx context.Context, tableName string) (int64, bool) {
	var estimate sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `
		SELECT c.reltuples
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public'
		  AND c.relname = $1
		  AND c.relkind IN ('r', 'm', 'p')
	`, tableName).Scan(&estimate)
	if err != nil || !estimate.Valid {
		return 0, false
	}
	if estimate.Float64 < 0 {
		return 0, false
	}
	return int64(estimate.Float64), true
}

func (r *MetaRepository) tableHasRows(ctx context.Context, tableName string) (bool, bool) {
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s LIMIT 1)`, quoteIdentifier(tableName))
	var exists bool
	if err := r.db.QueryRowContext(ctx, query).Scan(&exists); err != nil {
		return false, false
	}
	return exists, true
}

func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func toSeasonYear(value sql.NullInt64) core.SeasonYear {
	if !value.Valid {
		return 0
	}
	return core.SeasonYear(value.Int64)
}

func seasonFromDate(value sql.NullString) core.SeasonYear {
	if !value.Valid || len(value.String) < 4 {
		return 0
	}
	year, err := strconv.Atoi(value.String[:4])
	if err != nil {
		return 0
	}
	return core.SeasonYear(year)
}

func seasonPtr(year core.SeasonYear) *core.SeasonYear {
	if year == 0 {
		return nil
	}
	y := year
	return &y
}

type hashAccumulator struct {
	hash hash.Hash
}

func newAccumulator() *hashAccumulator {
	return &hashAccumulator{hash: sha256.New()}
}

func (h *hashAccumulator) write(value string) {
	h.hash.Write([]byte(value))
	h.hash.Write([]byte{'\n'})
}

func (h *hashAccumulator) sum() string {
	return hex.EncodeToString(h.hash.Sum(nil))
}

func classifyMigration(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "lahman"):
		return "lahman"
	case strings.Contains(lower, "retrosheet") && strings.Contains(lower, "play"):
		return "retrosheet_plays"
	case strings.Contains(lower, "retrosheet"):
		return "retrosheet_games"
	default:
		return "core"
	}
}

// WOBAConstants returns wOBA calculation constants for seasons.
func (r *MetaRepository) WOBAConstants(ctx context.Context, season *core.SeasonYear) ([]core.WOBAConstant, error) {
	query := `
		SELECT
			season, w_bb, w_hbp, w_1b, w_2b, w_3b, w_hr,
			woba_scale, woba, run_sb, run_cs, r_pa, r_w, c_fip
		FROM woba_constants
		WHERE ($1::int IS NULL OR season = $1)
		ORDER BY season DESC
	`

	var seasonVal sql.NullInt64
	if season != nil {
		seasonVal = sql.NullInt64{Int64: int64(*season), Valid: true}
	}

	rows, err := r.db.QueryContext(ctx, query, seasonVal)
	if err != nil {
		return nil, fmt.Errorf("failed to query wOBA constants: %w", err)
	}
	defer rows.Close()

	var constants []core.WOBAConstant
	for rows.Next() {
		var c core.WOBAConstant
		err := rows.Scan(
			&c.Season, &c.WBB, &c.WHBP, &c.W1B, &c.W2B, &c.W3B, &c.WHR,
			&c.WOBAScale, &c.WOBA, &c.RunSB, &c.RunCS, &c.RPA, &c.RW, &c.CFIP,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wOBA constant: %w", err)
		}
		constants = append(constants, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate wOBA constants: %w", err)
	}

	return constants, nil
}

// LeagueConstants returns league-specific constants for wRC+ calculations.
func (r *MetaRepository) LeagueConstants(ctx context.Context, season *core.SeasonYear, league *core.LeagueID) ([]core.LeagueConstant, error) {
	query := `
		SELECT
			season, league, woba_avg, wrc_per_pa, runs_per_win,
			replacement_runs_per_pa, total_pa, total_runs
		FROM league_constants
		WHERE ($1::int IS NULL OR season = $1)
			AND ($2::text IS NULL OR league = $2)
		ORDER BY season DESC, league
	`

	var seasonVal sql.NullInt64
	if season != nil {
		seasonVal = sql.NullInt64{Int64: int64(*season), Valid: true}
	}

	var leagueVal sql.NullString
	if league != nil {
		leagueVal = sql.NullString{String: string(*league), Valid: true}
	}

	rows, err := r.db.QueryContext(ctx, query, seasonVal, leagueVal)
	if err != nil {
		return nil, fmt.Errorf("failed to query league constants: %w", err)
	}
	defer rows.Close()

	var constants []core.LeagueConstant
	for rows.Next() {
		var c core.LeagueConstant
		var leagueStr string
		var wobaAvg, wrcPerPA, runsPerWin, replacementPA sql.NullFloat64
		var totalPA, totalRuns sql.NullInt64

		err := rows.Scan(
			&c.Season, &leagueStr, &wobaAvg, &wrcPerPA, &runsPerWin,
			&replacementPA, &totalPA, &totalRuns,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan league constant: %w", err)
		}

		c.League = core.LeagueID(leagueStr)
		if wobaAvg.Valid {
			c.WOBAAvg = &wobaAvg.Float64
		}
		if wrcPerPA.Valid {
			c.WRCPerPA = &wrcPerPA.Float64
		}
		if runsPerWin.Valid {
			c.RunsPerWin = &runsPerWin.Float64
		}
		if replacementPA.Valid {
			c.ReplacementPA = &replacementPA.Float64
		}
		if totalPA.Valid {
			c.TotalPA = &totalPA.Int64
		}
		if totalRuns.Valid {
			c.TotalRuns = &totalRuns.Int64
		}

		constants = append(constants, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate league constants: %w", err)
	}

	return constants, nil
}

// ParkFactors returns park factors for seasons.
func (r *MetaRepository) ParkFactors(ctx context.Context, season *core.SeasonYear, teamID *core.TeamID) ([]core.ParkFactorRow, error) {
	query := `
		SELECT
			pf.park_id, pf.season, pf.team_id,
			pf.basic_5yr, pf.basic_3yr, pf.basic_1yr,
			pf.factor_1b, pf.factor_2b, pf.factor_3b, pf.factor_hr,
			pf.factor_so, pf.factor_bb, pf.factor_gb, pf.factor_fb,
			pf.factor_ld, pf.factor_iffb, pf.factor_fip
		FROM park_factors pf
		WHERE ($1::int IS NULL OR pf.season = $1)
			AND ($2::text IS NULL OR pf.team_id = $2)
		ORDER BY pf.season DESC, pf.park_id
	`

	var seasonVal sql.NullInt64
	if season != nil {
		seasonVal = sql.NullInt64{Int64: int64(*season), Valid: true}
	}

	var teamVal sql.NullString
	if teamID != nil {
		teamVal = sql.NullString{String: string(*teamID), Valid: true}
	}

	rows, err := r.db.QueryContext(ctx, query, seasonVal, teamVal)
	if err != nil {
		return nil, fmt.Errorf("failed to query park factors: %w", err)
	}
	defer rows.Close()

	var factors []core.ParkFactorRow
	for rows.Next() {
		var pf core.ParkFactorRow
		var teamIDStr sql.NullString
		var basic5yr, basic3yr, basic1yr sql.NullInt64
		var f1b, f2b, f3b, fhr, fso, fbb, fgb, ffb, fld, fiffb, ffip sql.NullInt64

		err := rows.Scan(
			&pf.ParkID, &pf.Season, &teamIDStr,
			&basic5yr, &basic3yr, &basic1yr,
			&f1b, &f2b, &f3b, &fhr, &fso, &fbb, &fgb, &ffb, &fld, &fiffb, &ffip,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park factor: %w", err)
		}

		if teamIDStr.Valid {
			tid := core.TeamID(teamIDStr.String)
			pf.TeamID = &tid
		}

		if basic5yr.Valid {
			v := int(basic5yr.Int64)
			pf.Basic5yr = &v
		}
		if basic3yr.Valid {
			v := int(basic3yr.Int64)
			pf.Basic3yr = &v
		}
		if basic1yr.Valid {
			v := int(basic1yr.Int64)
			pf.Basic1yr = &v
		}
		if f1b.Valid {
			v := int(f1b.Int64)
			pf.Factor1B = &v
		}
		if f2b.Valid {
			v := int(f2b.Int64)
			pf.Factor2B = &v
		}
		if f3b.Valid {
			v := int(f3b.Int64)
			pf.Factor3B = &v
		}
		if fhr.Valid {
			v := int(fhr.Int64)
			pf.FactorHR = &v
		}
		if fso.Valid {
			v := int(fso.Int64)
			pf.FactorSO = &v
		}
		if fbb.Valid {
			v := int(fbb.Int64)
			pf.FactorBB = &v
		}
		if fgb.Valid {
			v := int(fgb.Int64)
			pf.FactorGB = &v
		}
		if ffb.Valid {
			v := int(ffb.Int64)
			pf.FactorFB = &v
		}
		if fld.Valid {
			v := int(fld.Int64)
			pf.FactorLD = &v
		}
		if fiffb.Valid {
			v := int(fiffb.Int64)
			pf.FactorIFFB = &v
		}
		if ffip.Valid {
			v := int(ffip.Int64)
			pf.FactorFIP = &v
		}

		factors = append(factors, pf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate park factors: %w", err)
	}

	return factors, nil
}
