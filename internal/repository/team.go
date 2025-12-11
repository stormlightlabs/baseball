package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

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

func (r *TeamRepository) Roster(ctx context.Context, year core.SeasonYear, teamID core.TeamID) ([]core.PlayerBattingSeason, error) {
	return nil, nil
}
