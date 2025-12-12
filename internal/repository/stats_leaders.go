package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

// TODO: param struct
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
	orderColumn := "SUM(\"HR\")"
	switch stat {
	case "avg":
		orderColumn = "CASE WHEN SUM(\"AB\") > 0 THEN CAST(SUM(\"H\") AS FLOAT) / SUM(\"AB\") ELSE 0 END"
	case "hr":
		orderColumn = "SUM(\"HR\")"
	case "rbi":
		orderColumn = "SUM(\"RBI\")"
	case "sb":
		orderColumn = "SUM(\"SB\")"
	case "h":
		orderColumn = "SUM(\"H\")"
	case "r":
		orderColumn = "SUM(\"R\")"
	case "ops":
		// OPS requires calculating OBP and SLG from aggregates
		orderColumn = "(CASE WHEN (SUM(\"AB\") + SUM(\"BB\") + SUM(\"HBP\") + SUM(\"SF\")) > 0 THEN CAST(SUM(\"H\") + SUM(\"BB\") + SUM(\"HBP\") AS FLOAT) / (SUM(\"AB\") + SUM(\"BB\") + SUM(\"HBP\") + SUM(\"SF\")) ELSE 0 END) + (CASE WHEN SUM(\"AB\") > 0 THEN CAST(SUM(\"H\") + SUM(\"2B\") + (SUM(\"3B\") * 2) + (SUM(\"HR\") * 3) AS FLOAT) / SUM(\"AB\") ELSE 0 END)"
	}

	query := fmt.Sprintf(`
		SELECT
			"playerID",
			MAX("yearID") as last_year,
			'' as "teamID",
			'' as "lgID",
			SUM("G") as g,
			SUM("AB") as ab,
			SUM("R") as r,
			SUM("H") as h,
			SUM("2B") as doubles,
			SUM("3B") as triples,
			SUM("HR") as hr,
			SUM("RBI") as rbi,
			SUM("SB") as sb,
			SUM("CS") as cs,
			SUM("BB") as bb,
			SUM("SO") as so,
			SUM("HBP") as hbp,
			SUM("SF") as sf
		FROM "Batting"
		WHERE "AB" > 0
		GROUP BY "playerID"
		HAVING SUM("AB") >= 1000
		ORDER BY %s DESC, SUM("H") DESC
		LIMIT $1 OFFSET $2
	`, orderColumn)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get career batting leaders: %w", err)
	}
	defer rows.Close()

	var leaders []core.PlayerBattingSeason
	for rows.Next() {
		var s core.PlayerBattingSeason
		var doubles, triples, sb, cs, bb, so, hbp, sf sql.NullInt64
		var teamID, lgID string

		err := rows.Scan(
			&s.PlayerID, &s.Year, &teamID, &lgID,
			&s.G, &s.AB, &s.R, &s.H, &doubles, &triples, &s.HR, &s.RBI, &sb, &cs, &bb, &so, &hbp, &sf,
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
		if sf.Valid {
			s.SF = int(sf.Int64)
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
	orderColumn := "SUM(\"W\")"
	sortDir := "DESC"

	switch stat {
	case "era":
		orderColumn = "CASE WHEN SUM(\"IPouts\") > 0 THEN (CAST(SUM(\"ER\") AS FLOAT) * 27.0) / SUM(\"IPouts\") ELSE 999 END"
		sortDir = "ASC"
	case "so", "k":
		orderColumn = "SUM(\"SO\")"
	case "w", "wins":
		orderColumn = "SUM(\"W\")"
	case "sv", "saves":
		orderColumn = "SUM(\"SV\")"
	case "ip":
		orderColumn = "SUM(\"IPouts\")"
	}

	query := fmt.Sprintf(`
		SELECT
			"playerID",
			MAX("yearID") as last_year,
			'' as "teamID",
			'' as "lgID",
			SUM("W") as w,
			SUM("L") as l,
			SUM("G") as g,
			SUM("GS") as gs,
			SUM("CG") as cg,
			SUM("SHO") as sho,
			SUM("SV") as sv,
			SUM("IPouts") as ipouts,
			SUM("H") as h,
			SUM("ER") as er,
			SUM("HR") as hr,
			SUM("BB") as bb,
			SUM("SO") as so,
			SUM("HBP") as hbp,
			SUM("BK") as bk,
			SUM("WP") as wp,
			CASE WHEN SUM("IPouts") > 0 THEN (CAST(SUM("ER") AS FLOAT) * 27.0) / SUM("IPouts") ELSE 0 END as era
		FROM "Pitching"
		WHERE "IPouts" > 0
		GROUP BY "playerID"
		HAVING SUM("IPouts") >= 1500
		ORDER BY %s %s
		LIMIT $1 OFFSET $2
	`, orderColumn, sortDir)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get career pitching leaders: %w", err)
	}
	defer rows.Close()

	var leaders []core.PlayerPitchingSeason
	for rows.Next() {
		var s core.PlayerPitchingSeason
		var era sql.NullFloat64
		var hbp, bk, wp sql.NullInt64
		var teamID, lgID string

		err := rows.Scan(
			&s.PlayerID, &s.Year, &teamID, &lgID,
			&s.W, &s.L, &s.G, &s.GS, &s.CG, &s.SHO, &s.SV, &s.IPOuts, &s.H, &s.ER, &s.HR, &s.BB, &s.SO, &hbp, &bk, &wp, &era,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pitching leader: %w", err)
		}

		if era.Valid {
			s.ERA = era.Float64
		}
		if hbp.Valid {
			s.HBP = int(hbp.Int64)
		}
		if bk.Valid {
			s.BK = int(bk.Int64)
		}
		if wp.Valid {
			s.WP = int(wp.Int64)
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
