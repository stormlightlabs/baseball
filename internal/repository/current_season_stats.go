package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

// CurrentSeasonStatsRepository serves current-season stats from current_season schema.
type CurrentSeasonStatsRepository struct {
	db *sql.DB
}

func NewCurrentSeasonStatsRepository(db *sql.DB) *CurrentSeasonStatsRepository {
	return &CurrentSeasonStatsRepository{db: db}
}

func (r *CurrentSeasonStatsRepository) ShouldUseBattingSeason(ctx context.Context, year core.SeasonYear) (bool, error) {
	hasCurrent, err := r.currentSeasonRowsForYear(ctx, "batting", year)
	if err != nil {
		return false, err
	}
	if !hasCurrent {
		return false, nil
	}
	if int(year) >= currentYear() {
		return true, nil
	}
	hasLahman, err := r.lahmanRowsForYear(ctx, "Batting", "yearID", year)
	if err != nil {
		return false, err
	}
	return !hasLahman, nil
}

func (r *CurrentSeasonStatsRepository) ShouldUsePitchingSeason(ctx context.Context, year core.SeasonYear) (bool, error) {
	hasCurrent, err := r.currentSeasonRowsForYear(ctx, "pitching", year)
	if err != nil {
		return false, err
	}
	if !hasCurrent {
		return false, nil
	}
	if int(year) >= currentYear() {
		return true, nil
	}
	hasLahman, err := r.lahmanRowsForYear(ctx, "Pitching", "yearID", year)
	if err != nil {
		return false, err
	}
	return !hasLahman, nil
}

func (r *CurrentSeasonStatsRepository) lahmanRowsForYear(ctx context.Context, table, yearColumn string, year core.SeasonYear) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM "%s" WHERE "%s" = $1 LIMIT 1)`, table, yearColumn)
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, int(year)).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed checking %s rows for season %d: %w", table, year, err)
	}
	return exists, nil
}

func (r *CurrentSeasonStatsRepository) currentSeasonRowsForYear(ctx context.Context, table string, year core.SeasonYear) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM current_season.%s WHERE season = $1 LIMIT 1)`, table)
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, int(year)).Scan(&exists); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "current_season") {
			return false, nil
		}
		return false, fmt.Errorf("failed checking current_season.%s rows for season %d: %w", table, year, err)
	}
	return exists, nil
}

func (r *CurrentSeasonStatsRepository) SeasonBattingLeaders(ctx context.Context, year core.SeasonYear, stat string, limit, offset int, league *core.LeagueID) ([]core.PlayerBattingSeason, error) {
	orderExpr := map[string]string{
		"avg": "COALESCE(cs.avg, 0)",
		"hr":  "COALESCE(cs.hr, 0)",
		"rbi": "COALESCE(cs.rbi, 0)",
		"sb":  "COALESCE(cs.sb, 0)",
		"h":   "COALESCE(cs.h, 0)",
		"r":   "COALESCE(cs.r, 0)",
	}[strings.ToLower(strings.TrimSpace(stat))]
	if orderExpr == "" {
		orderExpr = "COALESCE(cs.hr, 0)"
	}

	leagueValue := ""
	if league != nil {
		leagueValue = string(*league)
	}

	query := fmt.Sprintf(`
		SELECT
			cs.player_id,
			cs.season,
			COALESCE(cs.team_id, tm.local_team_id, ''),
			COALESCE(tm.local_league, ''),
			COALESCE(cs.g, 0),
			COALESCE(cs.pa, 0),
			COALESCE(cs.ab, 0),
			COALESCE(cs.r, 0),
			COALESCE(cs.h, 0),
			COALESCE(cs."2b", 0),
			COALESCE(cs."3b", 0),
			COALESCE(cs.hr, 0),
			COALESCE(cs.rbi, 0),
			COALESCE(cs.sb, 0),
			COALESCE(cs.cs, 0),
			COALESCE(cs.bb, 0),
			COALESCE(cs.so, 0),
			COALESCE(cs.hbp, 0),
			COALESCE(cs.sf, 0),
			cs.avg,
			cs.obp,
			cs.slg,
			cs.ops,
			cs.fetched_at
		FROM current_season.batting cs
		LEFT JOIN team_mlbam_map tm
			ON tm.season = cs.season
			AND tm.mlbam_team_id = cs.team_mlb_id
		WHERE cs.season = $1
			AND cs.player_id IS NOT NULL
			AND ($2::text = '' OR COALESCE(tm.local_league, '') = $2)
		ORDER BY %s DESC, COALESCE(cs.h, 0) DESC
		LIMIT $3 OFFSET $4
	`, orderExpr)

	rows, err := r.db.QueryContext(ctx, query, int(year), leagueValue, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query current-season batting leaders: %w", err)
	}
	defer rows.Close()

	leaders := []core.PlayerBattingSeason{}
	for rows.Next() {
		season, err := scanCurrentSeasonBattingRow(rows)
		if err != nil {
			return nil, err
		}
		leaders = append(leaders, season)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating current-season batting leaders: %w", err)
	}
	return leaders, nil
}

func (r *CurrentSeasonStatsRepository) SeasonPitchingLeaders(ctx context.Context, year core.SeasonYear, stat string, limit, offset int, league *core.LeagueID) ([]core.PlayerPitchingSeason, error) {
	orderExpr := map[string]string{
		"era":   "COALESCE(cs.era, 999)",
		"so":    "COALESCE(cs.so, 0)",
		"k":     "COALESCE(cs.so, 0)",
		"w":     "COALESCE(cs.w, 0)",
		"wins":  "COALESCE(cs.w, 0)",
		"sv":    "COALESCE(cs.sv, 0)",
		"saves": "COALESCE(cs.sv, 0)",
		"ip":    "COALESCE(cs.ip, 0)",
	}[strings.ToLower(strings.TrimSpace(stat))]
	if orderExpr == "" {
		orderExpr = "COALESCE(cs.w, 0)"
	}
	sortDir := "DESC"
	if strings.EqualFold(strings.TrimSpace(stat), "era") {
		sortDir = "ASC"
	}

	leagueValue := ""
	if league != nil {
		leagueValue = string(*league)
	}

	query := fmt.Sprintf(`
		SELECT
			cs.player_id,
			cs.season,
			COALESCE(cs.team_id, tm.local_team_id, ''),
			COALESCE(tm.local_league, ''),
			COALESCE(cs.w, 0),
			COALESCE(cs.l, 0),
			COALESCE(cs.g, 0),
			COALESCE(cs.gs, 0),
			COALESCE(cs.sv, 0),
			COALESCE(cs.ip, 0),
			COALESCE(cs.h, 0),
			COALESCE(cs.er, 0),
			COALESCE(cs.hr, 0),
			COALESCE(cs.bb, 0),
			COALESCE(cs.so, 0),
			COALESCE(cs.hbp, 0),
			cs.era,
			cs.whip,
			cs.fetched_at
		FROM current_season.pitching cs
		LEFT JOIN team_mlbam_map tm
			ON tm.season = cs.season
			AND tm.mlbam_team_id = cs.team_mlb_id
		WHERE cs.season = $1
			AND cs.player_id IS NOT NULL
			AND ($2::text = '' OR COALESCE(tm.local_league, '') = $2)
		ORDER BY %s %s, COALESCE(cs.so, 0) DESC
		LIMIT $3 OFFSET $4
	`, orderExpr, sortDir)

	rows, err := r.db.QueryContext(ctx, query, int(year), leagueValue, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query current-season pitching leaders: %w", err)
	}
	defer rows.Close()

	leaders := []core.PlayerPitchingSeason{}
	for rows.Next() {
		season, err := scanCurrentSeasonPitchingRow(rows)
		if err != nil {
			return nil, err
		}
		leaders = append(leaders, season)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating current-season pitching leaders: %w", err)
	}
	return leaders, nil
}

func (r *CurrentSeasonStatsRepository) QueryBattingStats(ctx context.Context, filter core.BattingStatsFilter) ([]core.PlayerBattingSeason, error) {
	query := `
		SELECT
			cs.player_id,
			cs.season,
			COALESCE(cs.team_id, tm.local_team_id, ''),
			COALESCE(tm.local_league, ''),
			COALESCE(cs.g, 0),
			COALESCE(cs.pa, 0),
			COALESCE(cs.ab, 0),
			COALESCE(cs.r, 0),
			COALESCE(cs.h, 0),
			COALESCE(cs."2b", 0),
			COALESCE(cs."3b", 0),
			COALESCE(cs.hr, 0),
			COALESCE(cs.rbi, 0),
			COALESCE(cs.sb, 0),
			COALESCE(cs.cs, 0),
			COALESCE(cs.bb, 0),
			COALESCE(cs.so, 0),
			COALESCE(cs.hbp, 0),
			COALESCE(cs.sf, 0),
			cs.avg,
			cs.obp,
			cs.slg,
			cs.ops,
			cs.fetched_at
		FROM current_season.batting cs
		LEFT JOIN team_mlbam_map tm
			ON tm.season = cs.season
			AND tm.mlbam_team_id = cs.team_mlb_id
		WHERE cs.player_id IS NOT NULL
	`

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND cs.player_id = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}
	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.team_id, tm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}
	if filter.Season != nil {
		query += fmt.Sprintf(" AND cs.season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}
	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND cs.season >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}
	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND cs.season <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}
	if filter.League != nil {
		query += fmt.Sprintf(" AND COALESCE(tm.local_league, '') = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}
	if filter.MinAB != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.ab, 0) >= $%d", argNum)
		args = append(args, *filter.MinAB)
		argNum++
	}

	sortExpr := map[string]string{
		"avg": "COALESCE(cs.avg, 0)",
		"hr":  "COALESCE(cs.hr, 0)",
		"rbi": "COALESCE(cs.rbi, 0)",
		"sb":  "COALESCE(cs.sb, 0)",
		"h":   "COALESCE(cs.h, 0)",
		"r":   "COALESCE(cs.r, 0)",
	}[strings.ToLower(strings.TrimSpace(filter.SortBy))]
	if sortExpr == "" {
		sortExpr = "COALESCE(cs.h, 0)"
	}

	sortDir := "DESC"
	if filter.SortOrder == core.SortAsc {
		sortDir = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s, COALESCE(cs.h, 0) DESC LIMIT $%d OFFSET $%d", sortExpr, sortDir, argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query current-season batting stats: %w", err)
	}
	defer rows.Close()

	stats := []core.PlayerBattingSeason{}
	for rows.Next() {
		season, err := scanCurrentSeasonBattingRow(rows)
		if err != nil {
			return nil, err
		}
		stats = append(stats, season)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating current-season batting stats: %w", err)
	}
	return stats, nil
}

func (r *CurrentSeasonStatsRepository) QueryBattingStatsCount(ctx context.Context, filter core.BattingStatsFilter) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM current_season.batting cs
		LEFT JOIN team_mlbam_map tm
			ON tm.season = cs.season
			AND tm.mlbam_team_id = cs.team_mlb_id
		WHERE cs.player_id IS NOT NULL
	`
	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND cs.player_id = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}
	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.team_id, tm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}
	if filter.Season != nil {
		query += fmt.Sprintf(" AND cs.season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}
	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND cs.season >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}
	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND cs.season <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}
	if filter.League != nil {
		query += fmt.Sprintf(" AND COALESCE(tm.local_league, '') = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}
	if filter.MinAB != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.ab, 0) >= $%d", argNum)
		args = append(args, *filter.MinAB)
		argNum++
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count current-season batting stats: %w", err)
	}
	return count, nil
}

func (r *CurrentSeasonStatsRepository) QueryPitchingStats(ctx context.Context, filter core.PitchingStatsFilter) ([]core.PlayerPitchingSeason, error) {
	query := `
		SELECT
			cs.player_id,
			cs.season,
			COALESCE(cs.team_id, tm.local_team_id, ''),
			COALESCE(tm.local_league, ''),
			COALESCE(cs.w, 0),
			COALESCE(cs.l, 0),
			COALESCE(cs.g, 0),
			COALESCE(cs.gs, 0),
			COALESCE(cs.sv, 0),
			COALESCE(cs.ip, 0),
			COALESCE(cs.h, 0),
			COALESCE(cs.er, 0),
			COALESCE(cs.hr, 0),
			COALESCE(cs.bb, 0),
			COALESCE(cs.so, 0),
			COALESCE(cs.hbp, 0),
			cs.era,
			cs.whip,
			cs.fetched_at
		FROM current_season.pitching cs
		LEFT JOIN team_mlbam_map tm
			ON tm.season = cs.season
			AND tm.mlbam_team_id = cs.team_mlb_id
		WHERE cs.player_id IS NOT NULL
	`

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND cs.player_id = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}
	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.team_id, tm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}
	if filter.Season != nil {
		query += fmt.Sprintf(" AND cs.season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}
	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND cs.season >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}
	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND cs.season <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}
	if filter.League != nil {
		query += fmt.Sprintf(" AND COALESCE(tm.local_league, '') = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}
	if filter.MinIP != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.ip, 0) >= $%d", argNum)
		args = append(args, *filter.MinIP)
		argNum++
	}
	if filter.MinGS != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.gs, 0) >= $%d", argNum)
		args = append(args, *filter.MinGS)
		argNum++
	}

	sortExpr := map[string]string{
		"era": "COALESCE(cs.era, 999)",
		"w":   "COALESCE(cs.w, 0)",
		"so":  "COALESCE(cs.so, 0)",
		"sv":  "COALESCE(cs.sv, 0)",
		"ip":  "COALESCE(cs.ip, 0)",
	}[strings.ToLower(strings.TrimSpace(filter.SortBy))]
	if sortExpr == "" {
		sortExpr = "COALESCE(cs.so, 0)"
	}

	sortDir := "DESC"
	if filter.SortOrder == core.SortAsc || strings.EqualFold(strings.TrimSpace(filter.SortBy), "era") {
		sortDir = "ASC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s, COALESCE(cs.so, 0) DESC LIMIT $%d OFFSET $%d", sortExpr, sortDir, argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query current-season pitching stats: %w", err)
	}
	defer rows.Close()

	stats := []core.PlayerPitchingSeason{}
	for rows.Next() {
		season, err := scanCurrentSeasonPitchingRow(rows)
		if err != nil {
			return nil, err
		}
		stats = append(stats, season)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating current-season pitching stats: %w", err)
	}
	return stats, nil
}

func (r *CurrentSeasonStatsRepository) QueryPitchingStatsCount(ctx context.Context, filter core.PitchingStatsFilter) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM current_season.pitching cs
		LEFT JOIN team_mlbam_map tm
			ON tm.season = cs.season
			AND tm.mlbam_team_id = cs.team_mlb_id
		WHERE cs.player_id IS NOT NULL
	`
	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND cs.player_id = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}
	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.team_id, tm.local_team_id, '') = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}
	if filter.Season != nil {
		query += fmt.Sprintf(" AND cs.season = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}
	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND cs.season >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}
	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND cs.season <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}
	if filter.League != nil {
		query += fmt.Sprintf(" AND COALESCE(tm.local_league, '') = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}
	if filter.MinIP != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.ip, 0) >= $%d", argNum)
		args = append(args, *filter.MinIP)
		argNum++
	}
	if filter.MinGS != nil {
		query += fmt.Sprintf(" AND COALESCE(cs.gs, 0) >= $%d", argNum)
		args = append(args, *filter.MinGS)
		argNum++
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count current-season pitching stats: %w", err)
	}
	return count, nil
}

func scanCurrentSeasonBattingRow(scanner interface {
	Scan(dest ...any) error
}) (core.PlayerBattingSeason, error) {
	var s core.PlayerBattingSeason
	var avg, obp, slg, ops sql.NullFloat64
	var fetchedAt time.Time
	if err := scanner.Scan(
		&s.PlayerID,
		&s.Year,
		&s.TeamID,
		&s.League,
		&s.G,
		&s.PA,
		&s.AB,
		&s.R,
		&s.H,
		&s.Doubles,
		&s.Triples,
		&s.HR,
		&s.RBI,
		&s.SB,
		&s.CS,
		&s.BB,
		&s.SO,
		&s.HBP,
		&s.SF,
		&avg,
		&obp,
		&slg,
		&ops,
		&fetchedAt,
	); err != nil {
		return core.PlayerBattingSeason{}, fmt.Errorf("failed to scan current-season batting row: %w", err)
	}

	if avg.Valid {
		s.AVG = avg.Float64
	} else if s.AB > 0 {
		s.AVG = float64(s.H) / float64(s.AB)
	}
	if obp.Valid {
		s.OBP = obp.Float64
	} else if s.PA > 0 {
		s.OBP = float64(s.H+s.BB+s.HBP) / float64(s.PA)
	}
	if slg.Valid {
		s.SLG = slg.Float64
	} else if s.AB > 0 {
		singles := s.H - s.Doubles - s.Triples - s.HR
		totalBases := singles + (s.Doubles * 2) + (s.Triples * 3) + (s.HR * 4)
		s.SLG = float64(totalBases) / float64(s.AB)
	}
	if ops.Valid {
		s.OPS = ops.Float64
	} else {
		s.OPS = s.OBP + s.SLG
	}

	s.Source = "current_season"
	s.DataSources = []string{"current_season"}
	s.FetchedAt = &fetchedAt
	return s, nil
}

func scanCurrentSeasonPitchingRow(scanner interface {
	Scan(dest ...any) error
}) (core.PlayerPitchingSeason, error) {
	var s core.PlayerPitchingSeason
	var ip float64
	var era, whip sql.NullFloat64
	var fetchedAt time.Time
	if err := scanner.Scan(
		&s.PlayerID,
		&s.Year,
		&s.TeamID,
		&s.League,
		&s.W,
		&s.L,
		&s.G,
		&s.GS,
		&s.SV,
		&ip,
		&s.H,
		&s.ER,
		&s.HR,
		&s.BB,
		&s.SO,
		&s.HBP,
		&era,
		&whip,
		&fetchedAt,
	); err != nil {
		return core.PlayerPitchingSeason{}, fmt.Errorf("failed to scan current-season pitching row: %w", err)
	}

	s.IPOuts = int(math.Round(ip * 3))
	innings := float64(s.IPOuts) / 3.0
	if era.Valid {
		s.ERA = era.Float64
	} else if innings > 0 {
		s.ERA = (float64(s.ER) * 9.0) / innings
	}
	if whip.Valid {
		s.WHIP = whip.Float64
	} else if innings > 0 {
		s.WHIP = float64(s.H+s.BB) / innings
	}
	if innings > 0 {
		s.KPer9 = (float64(s.SO) / innings) * 9.0
		s.BBPer9 = (float64(s.BB) / innings) * 9.0
		s.HRPer9 = (float64(s.HR) / innings) * 9.0
	}

	s.Source = "current_season"
	s.DataSources = []string{"current_season"}
	s.FetchedAt = &fetchedAt
	return s, nil
}
