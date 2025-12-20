package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

type CoachRepository struct {
	db *sql.DB
}

func NewCoachRepository(db *sql.DB) *CoachRepository {
	return &CoachRepository{db: db}
}

// GetByID retrieves a coach by their player ID.
func (r *CoachRepository) GetByID(ctx context.Context, id core.PlayerID) (*core.Coach, error) {
	query := `
		SELECT DISTINCT
			p."playerID",
			p."retroID",
			p."nameFirst",
			p."nameLast"
		FROM "People" p
		INNER JOIN coaches c ON c.retro_id = p."retroID"
		WHERE p."playerID" = $1
	`

	var coach core.Coach
	var retroID, firstName, lastName sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&coach.ID,
		&retroID,
		&firstName,
		&lastName,
	)

	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("coach", string(id))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get coach: %w", err)
	}

	if retroID.Valid {
		rid := core.RetroPlayerID(retroID.String)
		coach.RetroID = &rid
	}
	if firstName.Valid {
		coach.FirstName = &firstName.String
	}
	if lastName.Valid {
		coach.LastName = &lastName.String
	}

	return &coach, nil
}

// List retrieves all coaches with pagination.
func (r *CoachRepository) List(ctx context.Context, p core.Pagination) ([]core.Coach, error) {
	query := `
		SELECT DISTINCT
			p."playerID",
			p."retroID",
			p."nameFirst",
			p."nameLast"
		FROM "People" p
		INNER JOIN coaches c ON c.retro_id = p."retroID"
		ORDER BY p."nameLast", p."nameFirst"
	`

	args := []any{}
	if p.PerPage > 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, p.PerPage, (p.Page-1)*p.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list coaches: %w", err)
	}
	defer rows.Close()

	var coaches []core.Coach
	for rows.Next() {
		var coach core.Coach
		var retroID, firstName, lastName sql.NullString

		err := rows.Scan(
			&coach.ID,
			&retroID,
			&firstName,
			&lastName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coach: %w", err)
		}

		if retroID.Valid {
			rid := core.RetroPlayerID(retroID.String)
			coach.RetroID = &rid
		}
		if firstName.Valid {
			coach.FirstName = &firstName.String
		}
		if lastName.Valid {
			coach.LastName = &lastName.String
		}

		coaches = append(coaches, coach)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate coaches: %w", err)
	}

	return coaches, nil
}

// SeasonRecords retrieves all season coaching records for a coach.
func (r *CoachRepository) SeasonRecords(ctx context.Context, id core.PlayerID) ([]core.CoachSeasonRecord, error) {
	query := `
		SELECT
			c.retro_id,
			p."playerID",
			c.year,
			c.team_id,
			c.role,
			c.first_game,
			c.last_game
		FROM coaches c
		LEFT JOIN "People" p ON p."retroID" = c.retro_id
		WHERE p."playerID" = $1
		ORDER BY c.year DESC, c.team_id
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get coach season records: %w", err)
	}
	defer rows.Close()

	var records []core.CoachSeasonRecord
	for rows.Next() {
		var rec core.CoachSeasonRecord
		var playerID sql.NullString
		var role sql.NullString
		var firstGame, lastGame sql.NullTime

		err := rows.Scan(
			&rec.RetroID,
			&playerID,
			&rec.Year,
			&rec.TeamID,
			&role,
			&firstGame,
			&lastGame,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan coach season record: %w", err)
		}

		if playerID.Valid {
			pid := core.PlayerID(playerID.String)
			rec.PlayerID = &pid
		}
		if role.Valid {
			rec.Role = &role.String
		}
		if firstGame.Valid {
			rec.FirstGame = &firstGame.Time
		}
		if lastGame.Valid {
			rec.LastGame = &lastGame.Time
		}

		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate coach records: %w", err)
	}

	return records, nil
}
