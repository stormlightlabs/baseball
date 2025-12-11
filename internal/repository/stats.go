// TODO: construct param structs
package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) SeasonBattingLeaders(ctx context.Context, year core.SeasonYear, stat string, limit, offset int, league *core.LeagueID) ([]core.PlayerBattingSeason, error) {
	orderColumn := "\"HR\""

	switch stat {
	case "avg":
		orderColumn = "CASE WHEN \"AB\" > 0 THEN CAST(\"H\" AS FLOAT) / \"AB\" ELSE 0 END"
	case "hr":
		orderColumn = "\"HR\""
	case "rbi":
		orderColumn = "\"RBI\""
	case "sb":
		orderColumn = "\"SB\""
	case "h":
		orderColumn = "\"H\""
	case "r":
		orderColumn = "\"R\""
	}

	query := `
		SELECT
			"playerID", "yearID", "teamID", "lgID",
			"G", "AB", "R", "H", "2B", "3B", "HR", "RBI", "SB", "CS", "BB", "SO", "HBP", "SF"
		FROM "Batting"
		WHERE "yearID" = $1 AND "AB" >= 300
	`

	args := []any{int(year)}
	argNum := 2

	if league != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*league))
		argNum++
	}

	query += fmt.Sprintf(" ORDER BY %s DESC, \"H\" DESC LIMIT $%d OFFSET $%d", orderColumn, argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get batting leaders: %w", err)
	}
	defer rows.Close()

	var leaders []core.PlayerBattingSeason
	for rows.Next() {
		var s core.PlayerBattingSeason
		var doubles, triples sql.NullInt64

		err := rows.Scan(
			&s.PlayerID, &s.Year, &s.TeamID, &s.League,
			&s.G, &s.AB, &s.R, &s.H, &doubles, &triples, &s.HR, &s.RBI, &s.SB, &s.CS, &s.BB, &s.SO, &s.HBP, &s.SF,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batting leader: %w", err)
		}

		if doubles.Valid {
			s.Doubles = int(doubles.Int64)
		}
		if triples.Valid {
			s.Triples = int(triples.Int64)
		}

		s.PA = s.AB + s.BB + s.HBP + s.SF

		if s.AB > 0 {
			s.AVG = float64(s.H) / float64(s.AB)
			singles := s.H - s.Doubles - s.Triples - s.HR
			totalBases := singles + (s.Doubles * 2) + (s.Triples * 3) + (s.HR * 4)
			s.SLG = float64(totalBases) / float64(s.AB)
		}

		if s.PA > 0 {
			s.OBP = float64(s.H+s.BB+s.HBP) / float64(s.PA)
		}

		s.OPS = s.OBP + s.SLG

		leaders = append(leaders, s)
	}

	return leaders, nil
}

func (r *StatsRepository) CareerBattingLeaders(ctx context.Context, stat string, limit, offset int) ([]core.PlayerBattingSeason, error) {
	return nil, nil
}

func (r *StatsRepository) SeasonPitchingLeaders(ctx context.Context, year core.SeasonYear, stat string, limit, offset int, league *core.LeagueID) ([]core.PlayerPitchingSeason, error) {
	orderColumn := "\"W\""

	switch stat {
	case "era":
		orderColumn = "\"ERA\""
	case "so", "k":
		orderColumn = "\"SO\""
	case "w", "wins":
		orderColumn = "\"W\""
	case "sv", "saves":
		orderColumn = "\"SV\""
	case "ip":
		orderColumn = "\"IPouts\""
	}

	query := `
		SELECT
			"playerID", "yearID", "teamID", "lgID",
			"W", "L", "G", "GS", "CG", "SHO", "SV", "IPouts", "H", "ER", "HR", "BB", "SO", "HBP", "BK", "WP", "ERA"
		FROM "Pitching"
		WHERE "yearID" = $1 AND "IPouts" >= 450
	`

	args := []any{int(year)}
	argNum := 2

	if league != nil {
		query += fmt.Sprintf(" AND \"lgID\" = $%d", argNum)
		args = append(args, string(*league))
		argNum++
	}

	sortDir := "DESC"
	if stat == "era" {
		sortDir = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", orderColumn, sortDir, argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get pitching leaders: %w", err)
	}
	defer rows.Close()

	var leaders []core.PlayerPitchingSeason
	for rows.Next() {
		var s core.PlayerPitchingSeason
		var era sql.NullFloat64

		err := rows.Scan(
			&s.PlayerID, &s.Year, &s.TeamID, &s.League,
			&s.W, &s.L, &s.G, &s.GS, &s.CG, &s.SHO, &s.SV, &s.IPOuts, &s.H, &s.ER, &s.HR, &s.BB, &s.SO, &s.HBP, &s.BK, &s.WP, &era,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pitching leader: %w", err)
		}

		if era.Valid {
			s.ERA = era.Float64
		}

		ip := float64(s.IPOuts) / 3.0
		if ip > 0 {
			s.WHIP = float64(s.H+s.BB) / ip
			s.KPer9 = (float64(s.SO) / ip) * 9.0
			s.BBPer9 = (float64(s.BB) / ip) * 9.0
			s.HRPer9 = (float64(s.HR) / ip) * 9.0
		}

		leaders = append(leaders, s)
	}

	return leaders, nil
}

func (r *StatsRepository) CareerPitchingLeaders(ctx context.Context, stat string, limit, offset int) ([]core.PlayerPitchingSeason, error) {
	return nil, nil
}

func (r *StatsRepository) TeamSeasonStats(ctx context.Context, filter core.TeamFilter) ([]core.TeamSeason, error) {
	return nil, nil
}

// QueryBattingStats provides flexible batting stats querying with various filters.
func (r *StatsRepository) QueryBattingStats(ctx context.Context, filter core.BattingStatsFilter) ([]core.PlayerBattingSeason, error) {
	query := `
		SELECT
			"playerID", "yearID", "teamID", "lgID",
			"G", "AB", "R", "H", "2B", "3B", "HR", "RBI", "SB", "CS", "BB", "SO", "HBP", "SF"
		FROM "Batting"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND \"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

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

	if filter.MinAB != nil {
		query += fmt.Sprintf(" AND \"AB\" >= $%d", argNum)
		args = append(args, *filter.MinAB)
		argNum++
	}

	orderColumn := "\"H\""
	sortDir := "DESC"

	if filter.SortBy != "" {
		switch filter.SortBy {
		case "avg":
			orderColumn = "CASE WHEN \"AB\" > 0 THEN CAST(\"H\" AS FLOAT) / \"AB\" ELSE 0 END"
		case "hr":
			orderColumn = "\"HR\""
		case "rbi":
			orderColumn = "\"RBI\""
		case "sb":
			orderColumn = "\"SB\""
		case "h":
			orderColumn = "\"H\""
		case "r":
			orderColumn = "\"R\""
		}
	}

	if filter.SortOrder == core.SortAsc {
		sortDir = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", orderColumn, sortDir, argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query batting stats: %w", err)
	}
	defer rows.Close()

	var stats []core.PlayerBattingSeason
	for rows.Next() {
		var s core.PlayerBattingSeason
		var doubles, triples, sf, sb, cs, bb, so, hbp sql.NullInt64

		err := rows.Scan(
			&s.PlayerID, &s.Year, &s.TeamID, &s.League,
			&s.G, &s.AB, &s.R, &s.H, &doubles, &triples, &s.HR, &s.RBI, &sb, &cs, &bb, &so, &hbp, &sf,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan batting stats: %w", err)
		}

		if doubles.Valid {
			s.Doubles = int(doubles.Int64)
		}
		if triples.Valid {
			s.Triples = int(triples.Int64)
		}
		if sf.Valid {
			s.SF = int(sf.Int64)
		}
		if sb.Valid {
			s.SB = int(sb.Int64)
		}
		if cs.Valid {
			s.CS = int(cs.Int64)
		}
		if bb.Valid {
			s.BB = int(bb.Int64)
		}
		if so.Valid {
			s.SO = int(so.Int64)
		}
		if hbp.Valid {
			s.HBP = int(hbp.Int64)
		}

		s.PA = s.AB + s.BB + s.HBP + s.SF
		if s.AB > 0 {
			s.AVG = float64(s.H) / float64(s.AB)
			singles := s.H - s.Doubles - s.Triples - s.HR
			s.SLG = float64(singles+(s.Doubles*2)+(s.Triples*3)+(s.HR*4)) / float64(s.AB)
		}
		if s.PA > 0 {
			s.OBP = float64(s.H+s.BB+s.HBP) / float64(s.PA)
		}
		s.OPS = s.OBP + s.SLG

		stats = append(stats, s)
	}

	return stats, nil
}

// QueryBattingStatsCount returns the total count for the filter.
func (r *StatsRepository) QueryBattingStatsCount(ctx context.Context, filter core.BattingStatsFilter) (int, error) {
	query := `SELECT COUNT(*) FROM "Batting" WHERE 1=1`

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND \"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

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

	if filter.MinAB != nil {
		query += fmt.Sprintf(" AND \"AB\" >= $%d", argNum)
		args = append(args, *filter.MinAB)
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count batting stats: %w", err)
	}

	return count, nil
}

// QueryPitchingStats provides flexible pitching stats querying with various filters.
func (r *StatsRepository) QueryPitchingStats(ctx context.Context, filter core.PitchingStatsFilter) ([]core.PlayerPitchingSeason, error) {
	query := `
		SELECT
			"playerID", "yearID", "teamID", "lgID",
			"W", "L", "G", "GS", "CG", "SHO", "SV", "IPouts", "H", "ER", "HR", "BB", "SO", "HBP", "BK", "WP",
			CASE WHEN "IPouts" > 0 THEN (CAST("ER" AS FLOAT) * 27.0) / "IPouts" ELSE 0 END as "ERA"
		FROM "Pitching"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND \"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

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

	if filter.MinIP != nil {
		minOuts := int(*filter.MinIP * 3)
		query += fmt.Sprintf(" AND \"IPouts\" >= $%d", argNum)
		args = append(args, minOuts)
		argNum++
	}

	if filter.MinGS != nil {
		query += fmt.Sprintf(" AND \"GS\" >= $%d", argNum)
		args = append(args, *filter.MinGS)
		argNum++
	}

	orderColumn := "\"SO\""
	sortDir := "DESC"

	if filter.SortBy != "" {
		switch filter.SortBy {
		case "era":
			orderColumn = "CASE WHEN \"IPouts\" > 0 THEN (CAST(\"ER\" AS FLOAT) * 27.0) / \"IPouts\" ELSE 999 END"
			sortDir = "ASC"
		case "w":
			orderColumn = "\"W\""
		case "so":
			orderColumn = "\"SO\""
		case "sv":
			orderColumn = "\"SV\""
		case "ip":
			orderColumn = "\"IPouts\""
		}
	}

	if filter.SortOrder == core.SortAsc {
		sortDir = "ASC"
	} else if filter.SortBy == "era" {
		sortDir = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", orderColumn, sortDir, argNum, argNum+1)
	args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pitching stats: %w", err)
	}
	defer rows.Close()

	var stats []core.PlayerPitchingSeason
	for rows.Next() {
		var s core.PlayerPitchingSeason
		var era sql.NullFloat64

		err := rows.Scan(
			&s.PlayerID, &s.Year, &s.TeamID, &s.League,
			&s.W, &s.L, &s.G, &s.GS, &s.CG, &s.SHO, &s.SV, &s.IPOuts, &s.H, &s.ER, &s.HR, &s.BB, &s.SO, &s.HBP, &s.BK, &s.WP, &era,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pitching stats: %w", err)
		}

		if era.Valid {
			s.ERA = era.Float64
		}

		ip := float64(s.IPOuts) / 3.0
		if ip > 0 {
			s.WHIP = float64(s.H+s.BB) / ip
			s.KPer9 = (float64(s.SO) / ip) * 9.0
			s.BBPer9 = (float64(s.BB) / ip) * 9.0
			s.HRPer9 = (float64(s.HR) / ip) * 9.0
		}

		stats = append(stats, s)
	}

	return stats, nil
}

// QueryPitchingStatsCount returns the total count for the filter.
func (r *StatsRepository) QueryPitchingStatsCount(ctx context.Context, filter core.PitchingStatsFilter) (int, error) {
	query := `SELECT COUNT(*) FROM "Pitching" WHERE 1=1`

	args := []any{}
	argNum := 1

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND \"playerID\" = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

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

	if filter.MinIP != nil {
		minOuts := int(*filter.MinIP * 3)
		query += fmt.Sprintf(" AND \"IPouts\" >= $%d", argNum)
		args = append(args, minOuts)
		argNum++
	}

	if filter.MinGS != nil {
		query += fmt.Sprintf(" AND \"GS\" >= $%d", argNum)
		args = append(args, *filter.MinGS)
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pitching stats: %w", err)
	}

	return count, nil
}
