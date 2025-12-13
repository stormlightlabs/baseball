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

	minLahman, maxLahman, minRetro, maxRetro, err := r.SeasonCoverage(ctx)
	if err != nil {
		return nil, err
	}

	lahmanTables := map[string]int64{
		"people":      r.safeCount(ctx, `SELECT COUNT(*) FROM "People"`),
		"teams":       r.safeCount(ctx, `SELECT COUNT(*) FROM "Teams"`),
		"batting":     r.safeCount(ctx, `SELECT COUNT(*) FROM "Batting"`),
		"pitching":    r.safeCount(ctx, `SELECT COUNT(*) FROM "Pitching"`),
		"appearances": r.safeCount(ctx, `SELECT COUNT(*) FROM "Appearances"`),
		"salaries":    r.safeCount(ctx, `SELECT COUNT(*) FROM "Salaries"`),
	}

	var lahmanRowFallback int64
	for _, count := range lahmanTables {
		lahmanRowFallback += count
	}

	lahmanStatus := core.DatasetStatus{
		ID:           "lahman",
		Name:         "Lahman Baseball Database",
		Source:       "https://sabr.org/lahman-database/",
		CoverageFrom: seasonPtr(minLahman),
		CoverageTo:   seasonPtr(maxLahman),
		RowCount:     lahmanRowFallback,
		Tables:       lahmanTables,
	}
	if entry, ok := refreshes["lahman"]; ok {
		if !entry.lastLoaded.IsZero() {
			ts := entry.lastLoaded
			lahmanStatus.LastLoadedAt = &ts
		}
		if entry.rowCount > 0 {
			lahmanStatus.RowCount = entry.rowCount
		}
	}

	retroTables := map[string]int64{
		"games": r.safeCount(ctx, `SELECT COUNT(*) FROM games`),
		"plays": r.safeCount(ctx, `SELECT COUNT(*) FROM plays`),
	}

	retroRowFallback := retroTables["games"] + retroTables["plays"]
	retroStatus := core.DatasetStatus{
		ID:           "retrosheet",
		Name:         "Retrosheet Game Logs & Plays",
		Source:       "https://www.retrosheet.org/",
		CoverageFrom: seasonPtr(minRetro),
		CoverageTo:   seasonPtr(maxRetro),
		RowCount:     retroRowFallback,
		Tables:       retroTables,
	}

	var retroRecordedRows int64
	for _, key := range []string{"retrosheet_games", "retrosheet_plays"} {
		if entry, ok := refreshes[key]; ok {
			retroRecordedRows += entry.rowCount
			if retroStatus.LastLoadedAt == nil || entry.lastLoaded.After(*retroStatus.LastLoadedAt) {
				ts := entry.lastLoaded
				retroStatus.LastLoadedAt = &ts
			}
		}
	}
	if retroRecordedRows > 0 {
		retroStatus.RowCount = retroRecordedRows
	}

	return []core.DatasetStatus{lahmanStatus, retroStatus}, nil
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

func (r *MetaRepository) safeCount(ctx context.Context, query string) int64 {
	var count sql.NullInt64
	if err := r.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0
	}
	if !count.Valid {
		return 0
	}
	return count.Int64
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
