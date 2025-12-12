package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"stormlightlabs.org/baseball/internal/core"
)

type EjectionRepository struct {
	db *sql.DB
}

func NewEjectionRepository(db *sql.DB) *EjectionRepository {
	return &EjectionRepository{db: db}
}

// List retrieves ejections based on filter criteria
func (r *EjectionRepository) List(ctx context.Context, filter core.EjectionFilter) ([]core.Ejection, error) {
	query := `
		SELECT
			game_id,
			date,
			game_number,
			ejectee_id,
			ejectee_name,
			team,
			role,
			umpire_id,
			umpire_name,
			inning,
			reason
		FROM ejections
		WHERE 1=1
	`

	args := []any{}
	argNum := 1

	if filter.Season != nil {
		query += fmt.Sprintf(" AND SUBSTRING(date, 7, 4) = $%d", argNum)
		args = append(args, fmt.Sprintf("%04d", int(*filter.Season)))
		argNum++
	}

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND ejectee_id = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

	if filter.UmpireID != nil {
		query += fmt.Sprintf(" AND umpire_id = $%d", argNum)
		args = append(args, string(*filter.UmpireID))
		argNum++
	}

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND team = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Role != nil {
		query += fmt.Sprintf(" AND role = $%d", argNum)
		args = append(args, *filter.Role)
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	// Sorting
	sortBy := "date"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "DESC"
	if filter.SortOrder == core.SortAsc {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if filter.Pagination.PerPage > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, filter.Pagination.PerPage, (filter.Pagination.Page-1)*filter.Pagination.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list ejections: %w", err)
	}
	defer rows.Close()

	var ejections []core.Ejection
	for rows.Next() {
		ejection, err := scanEjection(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ejection: %w", err)
		}
		ejections = append(ejections, ejection)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate ejections: %w", err)
	}

	return ejections, nil
}

// Count returns the number of ejections matching the filter
func (r *EjectionRepository) Count(ctx context.Context, filter core.EjectionFilter) (int, error) {
	query := "SELECT COUNT(*) FROM ejections WHERE 1=1"

	args := []any{}
	argNum := 1

	if filter.Season != nil {
		query += fmt.Sprintf(" AND SUBSTRING(date, 7, 4) = $%d", argNum)
		args = append(args, fmt.Sprintf("%04d", int(*filter.Season)))
		argNum++
	}

	if filter.PlayerID != nil {
		query += fmt.Sprintf(" AND ejectee_id = $%d", argNum)
		args = append(args, string(*filter.PlayerID))
		argNum++
	}

	if filter.UmpireID != nil {
		query += fmt.Sprintf(" AND umpire_id = $%d", argNum)
		args = append(args, string(*filter.UmpireID))
		argNum++
	}

	if filter.TeamID != nil {
		query += fmt.Sprintf(" AND team = $%d", argNum)
		args = append(args, string(*filter.TeamID))
		argNum++
	}

	if filter.Role != nil {
		query += fmt.Sprintf(" AND role = $%d", argNum)
		args = append(args, *filter.Role)
		argNum++
	}

	if filter.DateFrom != nil {
		query += fmt.Sprintf(" AND date >= $%d", argNum)
		args = append(args, *filter.DateFrom)
		argNum++
	}

	if filter.DateTo != nil {
		query += fmt.Sprintf(" AND date <= $%d", argNum)
		args = append(args, *filter.DateTo)
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count ejections: %w", err)
	}

	return count, nil
}

// ListBySeason retrieves all ejections for a specific season
func (r *EjectionRepository) ListBySeason(ctx context.Context, year core.SeasonYear, p core.Pagination) ([]core.Ejection, error) {
	filter := core.EjectionFilter{
		Season:     &year,
		Pagination: p,
	}
	return r.List(ctx, filter)
}

// CountBySeason returns the number of ejections in a specific season
func (r *EjectionRepository) CountBySeason(ctx context.Context, year core.SeasonYear) (int, error) {
	filter := core.EjectionFilter{
		Season: &year,
	}
	return r.Count(ctx, filter)
}

func scanEjection(scanner interface {
	Scan(dest ...any) error
}) (core.Ejection, error) {
	var e core.Ejection
	var gameNumber sql.NullInt64
	var team, umpireID, umpireName, reason sql.NullString
	var inning sql.NullInt64

	err := scanner.Scan(
		&e.GameID,
		&e.Date,
		&gameNumber,
		&e.EjecteeID,
		&e.EjecteeName,
		&team,
		&e.Role,
		&umpireID,
		&umpireName,
		&inning,
		&reason,
	)
	if err != nil {
		return e, err
	}

	if gameNumber.Valid {
		gn := int(gameNumber.Int64)
		e.GameNumber = &gn
	}

	if team.Valid {
		t := core.TeamID(strings.TrimSpace(team.String))
		e.Team = &t
	}

	if umpireID.Valid {
		u := core.UmpireID(strings.TrimSpace(umpireID.String))
		e.UmpireID = &u
	}

	if umpireName.Valid {
		un := strings.TrimSpace(umpireName.String)
		e.UmpireName = &un
	}

	if inning.Valid {
		i := int(inning.Int64)
		e.Inning = &i
	}

	if reason.Valid {
		r := strings.TrimSpace(reason.String)
		e.Reason = &r
	}

	return e, nil
}
