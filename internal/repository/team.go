package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	"stormlightlabs.org/baseball/internal/core"
)

//go:embed queries/team_roster.sql
var teamRosterQuery string

//go:embed queries/list_seasons.sql
var listSeasonsQuery string

//go:embed queries/team_batting_agg.sql
var teamBattingAggQuery string

//go:embed queries/team_batting_players.sql
var teamBattingPlayersQuery string

//go:embed queries/team_pitching_agg.sql
var teamPitchingAggQuery string

//go:embed queries/team_pitching_players.sql
var teamPitchingPlayersQuery string

//go:embed queries/team_fielding_agg.sql
var teamFieldingAggQuery string

//go:embed queries/team_fielding_players.sql
var teamFieldingPlayersQuery string

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) GetTeamSeason(ctx context.Context, teamID core.TeamID, year core.SeasonYear) (*core.TeamSeason, error) {
	query := `
		SELECT
			"teamID", "yearID", "franchID", "lgID", "name", "park",
			"G", "W", "L", "R", "RA", "attendance", "divID"
		FROM "Teams"
		WHERE "teamID" = $1 AND "yearID" = $2
	`

	var ts core.TeamSeason
	var attendance sql.NullInt64
	var division sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(teamID), int(year)).Scan(
		&ts.TeamID, &ts.Year, &ts.FranchiseID, &ts.League, &ts.Name, &ts.ParkID,
		&ts.Games, &ts.Wins, &ts.Losses, &ts.RunsScored, &ts.RunsAllowed, &attendance, &division,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("team season not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team season: %w", err)
	}

	if attendance.Valid {
		att := int(attendance.Int64)
		ts.Attendance = &att
	}

	if division.Valid {
		ts.Division = &division.String
	}

	return &ts, nil
}

func (r *TeamRepository) ListTeamSeasons(ctx context.Context, filter core.TeamFilter) ([]core.TeamSeason, error) {
	query := `
		SELECT
			"teamID", "yearID", "franchID", "lgID", "name", "park",
			"G", "W", "L", "R", "RA", "attendance", "divID"
		FROM "Teams"
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.NameQuery != "" {
		query += fmt.Sprintf(` AND (
			"name" ILIKE $%d OR
			"teamID" ILIKE $%d OR
			"franchID" ILIKE $%d
		)`, argNum, argNum, argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

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

	query += " ORDER BY \"yearID\" DESC, \"W\" DESC"

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list team seasons: %w", err)
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

func (r *TeamRepository) CountTeamSeasons(ctx context.Context, filter core.TeamFilter) (int, error) {
	query := `SELECT COUNT(*) FROM "Teams" WHERE 1=1`
	args := []any{}
	argNum := 1

	if filter.NameQuery != "" {
		query += fmt.Sprintf(` AND (
			"name" ILIKE $%d OR
			"teamID" ILIKE $%d OR
			"franchID" ILIKE $%d
		)`, argNum, argNum, argNum)
		args = append(args, "%"+filter.NameQuery+"%")
		argNum++
	}

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
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *TeamRepository) GetFranchise(ctx context.Context, id core.FranchiseID) (*core.Franchise, error) {
	query := `
		SELECT
			"franchID", "franchName", "active", "NAassoc"
		FROM "TeamsFranchises"
		WHERE "franchID" = $1
	`

	var f core.Franchise
	var naAssoc, active sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&f.ID, &f.Name, &active, &naAssoc,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("franchise not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get franchise: %w", err)
	}

	if active.Valid {
		f.Active = active.String == "Y"
	}

	return &f, nil
}

func (r *TeamRepository) ListFranchises(ctx context.Context, onlyActive bool) ([]core.Franchise, error) {
	query := `
		SELECT
			"franchID", "franchName", "active", "NAassoc"
		FROM "TeamsFranchises"
		WHERE 1=1
	`

	args := []any{}
	if onlyActive {
		query += " AND \"active\" = 'Y'"
	}

	query += " ORDER BY \"franchName\""

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list franchises: %w", err)
	}
	defer rows.Close()

	var franchises []core.Franchise
	for rows.Next() {
		var f core.Franchise
		var naAssoc, active sql.NullString

		err := rows.Scan(
			&f.ID, &f.Name, &active, &naAssoc,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan franchise: %w", err)
		}

		if active.Valid {
			f.Active = active.String == "Y"
		}

		franchises = append(franchises, f)
	}

	return franchises, nil
}

func (r *TeamRepository) ListSeasons(ctx context.Context) ([]core.Season, error) {
	rows, err := r.db.QueryContext(ctx, listSeasonsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list seasons: %w", err)
	}
	defer rows.Close()

	var seasons []core.Season
	for rows.Next() {
		var s core.Season
		var leaguesStr string

		err := rows.Scan(&s.Year, &leaguesStr, &s.TeamCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan season: %w", err)
		}

		if leaguesStr != "" {
			for _, lg := range splitAndTrim(leaguesStr, ",") {
				s.Leagues = append(s.Leagues, core.LeagueID(lg))
			}
		}

		seasons = append(seasons, s)
	}

	return seasons, nil
}

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func (r *TeamRepository) Roster(ctx context.Context, year core.SeasonYear, teamID core.TeamID) ([]core.RosterPlayer, error) {
	rows, err := r.db.QueryContext(ctx, teamRosterQuery, string(teamID), int(year))
	if err != nil {
		return nil, fmt.Errorf("failed to get roster: %w", err)
	}
	defer rows.Close()

	var roster []core.RosterPlayer
	for rows.Next() {
		var rp core.RosterPlayer
		var position sql.NullString
		var battingG, ab, h, hr, rbi sql.NullInt64
		var avg sql.NullFloat64
		var pitchingG, w, l, so sql.NullInt64
		var era sql.NullFloat64

		err := rows.Scan(
			&rp.PlayerID, &rp.FirstName, &rp.LastName, &position,
			&battingG, &ab, &h, &hr, &rbi, &avg,
			&pitchingG, &w, &l, &era, &so,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan roster player: %w", err)
		}

		if position.Valid {
			rp.Position = &position.String
		}

		if battingG.Valid {
			g := int(battingG.Int64)
			rp.BattingG = &g
		}
		if ab.Valid {
			a := int(ab.Int64)
			rp.AB = &a
		}
		if h.Valid {
			hits := int(h.Int64)
			rp.H = &hits
		}
		if hr.Valid {
			hrs := int(hr.Int64)
			rp.HR = &hrs
		}
		if rbi.Valid {
			rbis := int(rbi.Int64)
			rp.RBI = &rbis
		}
		if avg.Valid {
			rp.AVG = &avg.Float64
		}

		if pitchingG.Valid {
			g := int(pitchingG.Int64)
			rp.PitchingG = &g
		}
		if w.Valid {
			wins := int(w.Int64)
			rp.W = &wins
		}
		if l.Valid {
			losses := int(l.Int64)
			rp.L = &losses
		}
		if era.Valid {
			rp.ERA = &era.Float64
		}
		if so.Valid {
			strikeouts := int(so.Int64)
			rp.SO = &strikeouts
		}

		roster = append(roster, rp)
	}

	return roster, nil
}

func (r *TeamRepository) BattingStats(ctx context.Context, year core.SeasonYear, teamID core.TeamID, includePlayers bool) (*core.TeamBattingStats, error) {
	var stats core.TeamBattingStats
	var ab, h, bb, hbp, sf int

	err := r.db.QueryRowContext(ctx, teamBattingAggQuery, string(teamID), int(year)).Scan(
		&stats.TeamID, &stats.Year, &stats.League,
		&stats.G, &ab, &stats.R, &h, &stats.Doubles, &stats.Triples,
		&stats.HR, &stats.RBI, &stats.SB, &stats.CS, &bb, &stats.SO, &hbp, &sf,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no batting stats found for team %s in year %d", teamID, year)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team batting stats: %w", err)
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
	singles := h - stats.Doubles - stats.Triples - stats.HR
	totalBases := singles + (stats.Doubles * 2) + (stats.Triples * 3) + (stats.HR * 4)
	if ab > 0 {
		stats.SLG = float64(totalBases) / float64(ab)
		stats.OPS = stats.OBP + stats.SLG
	}

	if includePlayers {
		rows, err := r.db.QueryContext(ctx, teamBattingPlayersQuery, string(teamID), int(year))
		if err != nil {
			return nil, fmt.Errorf("failed to get player batting stats: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ps core.PlayerBattingSeason
			var pa, ab, h, bb, hbp, sf int

			err := rows.Scan(
				&ps.PlayerID, &ps.Year, &ps.TeamID, &ps.League,
				&ps.G, &pa, &ab, &ps.R, &h, &ps.Doubles, &ps.Triples,
				&ps.HR, &ps.RBI, &ps.SB, &ps.CS, &bb, &ps.SO, &hbp, &sf,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan player batting: %w", err)
			}

			ps.PA = pa
			ps.AB = ab
			ps.H = h
			ps.BB = bb
			ps.HBP = hbp
			ps.SF = sf

			if ab > 0 {
				ps.AVG = float64(h) / float64(ab)
			}
			if ab+bb+hbp+sf > 0 {
				ps.OBP = float64(h+bb+hbp) / float64(ab+bb+hbp+sf)
			}
			singles := h - ps.Doubles - ps.Triples - ps.HR
			totalBases := singles + (ps.Doubles * 2) + (ps.Triples * 3) + (ps.HR * 4)
			if ab > 0 {
				ps.SLG = float64(totalBases) / float64(ab)
				ps.OPS = ps.OBP + ps.SLG
			}

			stats.Players = append(stats.Players, ps)
		}
	}

	return &stats, nil
}

func (r *TeamRepository) PitchingStats(ctx context.Context, year core.SeasonYear, teamID core.TeamID, includePlayers bool) (*core.TeamPitchingStats, error) {
	var stats core.TeamPitchingStats

	err := r.db.QueryRowContext(ctx, teamPitchingAggQuery, string(teamID), int(year)).Scan(
		&stats.TeamID, &stats.Year, &stats.League,
		&stats.W, &stats.L, &stats.G, &stats.GS, &stats.CG, &stats.SHO, &stats.SV,
		&stats.IPOuts, &stats.H, &stats.ER, &stats.HR, &stats.BB, &stats.SO,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no pitching stats found for team %s in year %d", teamID, year)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team pitching stats: %w", err)
	}

	if stats.IPOuts > 0 {
		stats.ERA = (float64(stats.ER) * 27.0) / float64(stats.IPOuts)
		innings := float64(stats.IPOuts) / 3.0
		stats.WHIP = (float64(stats.H) + float64(stats.BB)) / innings
	}

	if includePlayers {
		rows, err := r.db.QueryContext(ctx, teamPitchingPlayersQuery, string(teamID), int(year))
		if err != nil {
			return nil, fmt.Errorf("failed to get player pitching stats: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ps core.PlayerPitchingSeason

			err := rows.Scan(
				&ps.PlayerID, &ps.Year, &ps.TeamID, &ps.League,
				&ps.W, &ps.L, &ps.G, &ps.GS, &ps.CG, &ps.SHO, &ps.SV,
				&ps.IPOuts, &ps.H, &ps.ER, &ps.HR, &ps.BB, &ps.SO,
				&ps.HBP, &ps.BK, &ps.WP,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan player pitching: %w", err)
			}

			if ps.IPOuts > 0 {
				ps.ERA = (float64(ps.ER) * 27.0) / float64(ps.IPOuts)
				innings := float64(ps.IPOuts) / 3.0
				ps.WHIP = (float64(ps.H) + float64(ps.BB)) / innings
				ps.KPer9 = (float64(ps.SO) * 27.0) / float64(ps.IPOuts)
				ps.BBPer9 = (float64(ps.BB) * 27.0) / float64(ps.IPOuts)
				ps.HRPer9 = (float64(ps.HR) * 27.0) / float64(ps.IPOuts)
			}

			stats.Players = append(stats.Players, ps)
		}
	}

	return &stats, nil
}

func (r *TeamRepository) FieldingStats(ctx context.Context, year core.SeasonYear, teamID core.TeamID, includePlayers bool) (*core.TeamFieldingStats, error) {
	var stats core.TeamFieldingStats
	var wp int

	err := r.db.QueryRowContext(ctx, teamFieldingAggQuery, string(teamID), int(year)).Scan(
		&stats.TeamID, &stats.Year, &stats.League,
		&stats.G, &stats.PO, &stats.A, &stats.E, &stats.DP, &stats.PB, &wp,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no fielding stats found for team %s in year %d", teamID, year)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get team fielding stats: %w", err)
	}

	stats.SB = 0
	stats.CS = 0

	totalChances := stats.PO + stats.A + stats.E
	if totalChances > 0 {
		stats.FPct = float64(stats.PO+stats.A) / float64(totalChances)
	}

	if includePlayers {
		rows, err := r.db.QueryContext(ctx, teamFieldingPlayersQuery, string(teamID), int(year))
		if err != nil {
			return nil, fmt.Errorf("failed to get player fielding stats: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var ps core.PlayerFieldingSeason

			err := rows.Scan(
				&ps.PlayerID, &ps.Year, &ps.TeamID, &ps.League, &ps.Position,
				&ps.G, &ps.GS, &ps.Inn, &ps.PO, &ps.A, &ps.E, &ps.DP,
				&ps.PB, &ps.SB, &ps.CS,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan player fielding: %w", err)
			}

			if ps.Inn > 0 {
				innings := float64(ps.Inn) / 3.0
				ps.RF9 = (float64(ps.PO) + float64(ps.A)) * 9.0 / innings
			}

			stats.Players = append(stats.Players, ps)
		}
	}

	return &stats, nil
}
