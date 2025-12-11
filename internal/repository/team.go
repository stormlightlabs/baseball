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
	query := `
		SELECT
			"yearID",
			string_agg(DISTINCT "lgID", ',' ORDER BY "lgID") as leagues,
			COUNT(*) as team_count
		FROM "Teams"
		GROUP BY "yearID"
		ORDER BY "yearID" DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
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

		// Parse comma-separated leagues
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
