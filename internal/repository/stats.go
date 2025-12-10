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
	orderColumn := "\"HR\"" // default to home runs

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

	args := []interface{}{int(year)}
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
	orderColumn := "\"W\"" // default to wins

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

	args := []interface{}{int(year)}
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
