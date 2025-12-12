package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/stats_team_batting.sql
var statsTeamBattingQuery string

//go:embed queries/stats_team_pitching.sql
var statsTeamPitchingQuery string

//go:embed queries/stats_team_fielding.sql
var statsTeamFieldingQuery string

// TeamSeasonStats retrieves team season records based on filter criteria
func (r *StatsRepository) TeamSeasonStats(ctx context.Context, filter core.TeamFilter) ([]core.TeamSeason, error) {
	query := `
		SELECT
			"teamID", "yearID", "franchID", "lgID", "name", "park",
			"G", "W", "L", "R", "RA", "attendance", "divID"
		FROM "Teams"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.Year != nil {
		query += fmt.Sprintf(" AND \"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Year))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	if filter.FranchiseID != nil {
		query += fmt.Sprintf(" AND \"franchID\" = $%d", argNum)
		args = append(args, string(*filter.FranchiseID))
		argNum++
	}

	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "wins"
	}

	var orderByClause string
	switch sortBy {
	case "wins":
		orderByClause = "\"W\""
	case "losses":
		orderByClause = "\"L\""
	case "run_diff":
		orderByClause = "(\"R\" - \"RA\")"
	default:
		orderByClause = "\"W\""
	}

	if filter.SortOrder == core.SortAsc {
		query += fmt.Sprintf(" ORDER BY %s ASC", orderByClause)
	} else {
		query += fmt.Sprintf(" ORDER BY %s DESC", orderByClause)
	}

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query team season stats: %w", err)
	}
	defer rows.Close()

	var teams []core.TeamSeason
	for rows.Next() {
		var ts core.TeamSeason
		var attendance sql.NullInt64
		var division sql.NullString

		err := rows.Scan(
			&ts.TeamID, &ts.Year, &ts.FranchiseID, &ts.League, &ts.Name, &ts.ParkID,
			&ts.Games, &ts.Wins, &ts.Losses, &ts.RunsScored, &ts.RunsAllowed, &attendance, &division,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team season: %w", err)
		}

		if attendance.Valid {
			att := int(attendance.Int64)
			ts.Attendance = &att
		}

		if division.Valid {
			ts.Division = &division.String
		}

		teams = append(teams, ts)
	}

	return teams, nil
}

// TeamBattingStats retrieves aggregated batting statistics for teams based on filter criteria
func (r *StatsRepository) TeamBattingStats(ctx context.Context, filter core.TeamStatsFilter) ([]core.TeamBattingStats, error) {
	query := statsTeamBattingQuery

	args := []any{}
	argNum := 1

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND \"teamID\" = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Season != nil {
		query += fmt.Sprintf(" AND \"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND \"yearID\" >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}

	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND \"yearID\" <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	query += ` GROUP BY "teamID", "yearID", "lgID"`

	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "hr"
	}

	var orderByClause string
	switch sortBy {
	case "avg":
		orderByClause = "CASE WHEN SUM(\"AB\") > 0 THEN CAST(SUM(\"H\") AS FLOAT) / SUM(\"AB\") ELSE 0 END"
	case "obp":
		orderByClause = "CASE WHEN (SUM(\"AB\") + SUM(\"BB\") + SUM(\"HBP\") + SUM(\"SF\")) > 0 THEN CAST(SUM(\"H\") + SUM(\"BB\") + SUM(\"HBP\") AS FLOAT) / (SUM(\"AB\") + SUM(\"BB\") + SUM(\"HBP\") + SUM(\"SF\")) ELSE 0 END"
	case "slg":
		orderByClause = "CASE WHEN SUM(\"AB\") > 0 THEN CAST(SUM(\"H\") + SUM(\"2B\") + 2*SUM(\"3B\") + 3*SUM(\"HR\") AS FLOAT) / SUM(\"AB\") ELSE 0 END"
	case "ops":
		orderByClause = "(CASE WHEN SUM(\"AB\") > 0 THEN CAST(SUM(\"H\") AS FLOAT) / SUM(\"AB\") ELSE 0 END) + (CASE WHEN (SUM(\"AB\") + SUM(\"BB\") + SUM(\"HBP\") + SUM(\"SF\")) > 0 THEN CAST(SUM(\"H\") + SUM(\"BB\") + SUM(\"HBP\") AS FLOAT) / (SUM(\"AB\") + SUM(\"BB\") + SUM(\"HBP\") + SUM(\"SF\")) ELSE 0 END)"
	case "h":
		orderByClause = "SUM(\"H\")"
	case "r":
		orderByClause = "SUM(\"R\")"
	case "rbi":
		orderByClause = "SUM(\"RBI\")"
	case "sb":
		orderByClause = "SUM(\"SB\")"
	case "bb":
		orderByClause = "SUM(\"BB\")"
	case "so":
		orderByClause = "SUM(\"SO\")"
	default: // hr
		orderByClause = "SUM(\"HR\")"
	}

	if filter.SortOrder == core.SortAsc {
		query += fmt.Sprintf(" ORDER BY %s ASC", orderByClause)
	} else {
		query += fmt.Sprintf(" ORDER BY %s DESC", orderByClause)
	}

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query team batting stats: %w", err)
	}
	defer rows.Close()

	var results []core.TeamBattingStats
	for rows.Next() {
		var stats core.TeamBattingStats
		var ab, h, bb, hbp, sf int

		err := rows.Scan(
			&stats.TeamID, &stats.Year, &stats.League,
			&stats.G, &ab, &stats.R, &h, &stats.Doubles, &stats.Triples,
			&stats.HR, &stats.RBI, &stats.SB, &stats.CS, &bb, &stats.SO, &hbp, &sf,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team batting stats: %w", err)
		}

		stats.AB = ab
		stats.H = h
		stats.BB = bb
		stats.HBP = hbp
		stats.SF = sf

		if ab > 0 {
			stats.AVG = float64(h) / float64(ab)
		}
		if ab+bb+hbp+sf > 0 {
			stats.OBP = float64(h+bb+hbp) / float64(ab+bb+hbp+sf)
		}
		if ab > 0 {
			totalBases := h + stats.Doubles + 2*stats.Triples + 3*stats.HR
			stats.SLG = float64(totalBases) / float64(ab)
		}
		stats.OPS = stats.OBP + stats.SLG

		results = append(results, stats)
	}

	return results, nil
}

// TeamBattingStatsCount returns the count of team batting stat records matching the filter
func (r *StatsRepository) TeamBattingStatsCount(ctx context.Context, filter core.TeamStatsFilter) (int, error) {
	query := `
		SELECT COUNT(DISTINCT ("teamID", "yearID", "lgID"))
		FROM "Batting"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND \"teamID\" = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Season != nil {
		query += fmt.Sprintf(" AND \"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND \"yearID\" >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}

	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND \"yearID\" <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count team batting stats: %w", err)
	}

	return count, nil
}

// TeamPitchingStats retrieves aggregated pitching statistics for teams based on filter criteria
func (r *StatsRepository) TeamPitchingStats(ctx context.Context, filter core.TeamStatsFilter) ([]core.TeamPitchingStats, error) {
	query := statsTeamPitchingQuery

	args := []any{}
	argNum := 1

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND \"teamID\" = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Season != nil {
		query += fmt.Sprintf(" AND \"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND \"yearID\" >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}

	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND \"yearID\" <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	query += ` GROUP BY "teamID", "yearID", "lgID"`

	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "era"
	}

	var orderByClause string
	switch sortBy {
	case "era":
		orderByClause = "CASE WHEN SUM(\"IPouts\") > 0 THEN (CAST(SUM(\"ER\") AS FLOAT) * 27.0) / SUM(\"IPouts\") ELSE 999.99 END"
	case "whip":
		orderByClause = "CASE WHEN SUM(\"IPouts\") > 0 THEN (CAST(SUM(\"H\") + SUM(\"BB\") AS FLOAT) * 3.0) / SUM(\"IPouts\") ELSE 999.99 END"
	case "w":
		orderByClause = "SUM(\"W\")"
	case "l":
		orderByClause = "SUM(\"L\")"
	case "sv":
		orderByClause = "SUM(\"SV\")"
	case "so":
		orderByClause = "SUM(\"SO\")"
	case "ip":
		orderByClause = "SUM(\"IPouts\")"
	case "cg":
		orderByClause = "SUM(\"CG\")"
	case "sho":
		orderByClause = "SUM(\"SHO\")"
	default: // era
		orderByClause = "CASE WHEN SUM(\"IPouts\") > 0 THEN (CAST(SUM(\"ER\") AS FLOAT) * 27.0) / SUM(\"IPouts\") ELSE 999.99 END"
	}

	if filter.SortOrder == core.SortAsc {
		query += fmt.Sprintf(" ORDER BY %s ASC", orderByClause)
	} else {
		query += fmt.Sprintf(" ORDER BY %s DESC", orderByClause)
	}

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query team pitching stats: %w", err)
	}
	defer rows.Close()

	var results []core.TeamPitchingStats
	for rows.Next() {
		var stats core.TeamPitchingStats

		err := rows.Scan(
			&stats.TeamID, &stats.Year, &stats.League,
			&stats.W, &stats.L, &stats.G, &stats.GS, &stats.CG, &stats.SHO, &stats.SV,
			&stats.IPOuts, &stats.H, &stats.ER, &stats.HR, &stats.BB, &stats.SO,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team pitching stats: %w", err)
		}

		if stats.IPOuts > 0 {
			stats.ERA = (float64(stats.ER) * 27.0) / float64(stats.IPOuts)
			stats.WHIP = (float64(stats.H+stats.BB) * 3.0) / float64(stats.IPOuts)
		}

		results = append(results, stats)
	}

	return results, nil
}

// TeamPitchingStatsCount returns the count of team pitching stat records matching the filter
func (r *StatsRepository) TeamPitchingStatsCount(ctx context.Context, filter core.TeamStatsFilter) (int, error) {
	query := `
		SELECT COUNT(DISTINCT ("teamID", "yearID", "lgID"))
		FROM "Pitching"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND \"teamID\" = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Season != nil {
		query += fmt.Sprintf(" AND \"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND \"yearID\" >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}

	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND \"yearID\" <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count team pitching stats: %w", err)
	}

	return count, nil
}

// TeamFieldingStats retrieves aggregated fielding statistics for teams based on filter criteria
func (r *StatsRepository) TeamFieldingStats(ctx context.Context, filter core.TeamStatsFilter) ([]core.TeamFieldingStats, error) {
	query := statsTeamFieldingQuery

	args := []any{}
	argNum := 1

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND \"teamID\" = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Season != nil {
		query += fmt.Sprintf(" AND \"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND \"yearID\" >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}

	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND \"yearID\" <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	query += ` GROUP BY "teamID", "yearID", "lgID"`

	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = "po"
	}

	var orderByClause string
	switch sortBy {
	case "fpct":
		orderByClause = "CASE WHEN (SUM(\"PO\") + SUM(\"A\") + SUM(\"E\")) > 0 THEN CAST(SUM(\"PO\") + SUM(\"A\") AS FLOAT) / (SUM(\"PO\") + SUM(\"A\") + SUM(\"E\")) ELSE 0 END"
	case "po":
		orderByClause = "SUM(\"PO\")"
	case "a":
		orderByClause = "SUM(\"A\")"
	case "e":
		orderByClause = "SUM(\"E\")"
	case "dp":
		orderByClause = "SUM(\"DP\")"
	case "pb":
		orderByClause = "SUM(CASE WHEN \"POS\" = 'C' THEN \"PB\" ELSE 0 END)"
	default: // po
		orderByClause = "SUM(\"PO\")"
	}

	if filter.SortOrder == core.SortAsc {
		query += fmt.Sprintf(" ORDER BY %s ASC", orderByClause)
	} else {
		query += fmt.Sprintf(" ORDER BY %s DESC", orderByClause)
	}

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query team fielding stats: %w", err)
	}
	defer rows.Close()

	var results []core.TeamFieldingStats
	for rows.Next() {
		var stats core.TeamFieldingStats
		var wp sql.NullInt64

		err := rows.Scan(
			&stats.TeamID, &stats.Year, &stats.League,
			&stats.G, &stats.PO, &stats.A, &stats.E, &stats.DP,
			&stats.PB, &wp, &stats.SB, &stats.CS,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team fielding stats: %w", err)
		}

		if stats.PO+stats.A+stats.E > 0 {
			stats.FPct = float64(stats.PO+stats.A) / float64(stats.PO+stats.A+stats.E)
		}

		results = append(results, stats)
	}

	return results, nil
}

// TeamFieldingStatsCount returns the count of team fielding stat records matching the filter
func (r *StatsRepository) TeamFieldingStatsCount(ctx context.Context, filter core.TeamStatsFilter) (int, error) {
	query := `
		SELECT COUNT(DISTINCT ("teamID", "yearID", "lgID"))
		FROM "Fielding"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND \"teamID\" = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Season != nil {
		query += fmt.Sprintf(" AND \"yearID\" = $%d", argNum)
		args = append(args, int(*filter.Season))
		argNum++
	}

	if filter.SeasonFrom != nil {
		query += fmt.Sprintf(" AND \"yearID\" >= $%d", argNum)
		args = append(args, int(*filter.SeasonFrom))
		argNum++
	}

	if filter.SeasonTo != nil {
		query += fmt.Sprintf(" AND \"yearID\" <= $%d", argNum)
		args = append(args, int(*filter.SeasonTo))
		argNum++
	}

	if filter.League != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*filter.League))
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count team fielding stats: %w", err)
	}

	return count, nil
}
