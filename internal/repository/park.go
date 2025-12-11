package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

type ParkRepository struct {
	db *sql.DB
}

func NewParkRepository(db *sql.DB) *ParkRepository {
	return &ParkRepository{db: db}
}

// GetByID retrieves a park by its park key/ID.
func (r *ParkRepository) GetByID(ctx context.Context, id core.ParkID) (*core.Park, error) {
	query := `
		SELECT
			"parkkey",
			"parkname",
			"city",
			"state",
			"country"
		FROM "Parks"
		WHERE "parkkey" = $1
	`

	var park core.Park
	var state sql.NullString
	var country sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&park.ID,
		&park.Name,
		&park.City,
		&state,
		&country,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("park not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get park: %w", err)
	}

	if state.Valid {
		park.State = state.String
	}
	if country.Valid {
		park.Country = country.String
	}

	return &park, nil
}

// List retrieves all parks with pagination.
func (r *ParkRepository) List(ctx context.Context, p core.Pagination) ([]core.Park, error) {
	query := `
		SELECT DISTINCT
			"parkkey",
			"parkname",
			"city",
			"state",
			"country"
		FROM "Parks"
		WHERE "parkkey" IS NOT NULL
		  AND "parkname" IS NOT NULL
		ORDER BY "parkname"
	`

	args := []any{}
	if p.PerPage > 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, p.PerPage, (p.Page-1)*p.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list parks: %w", err)
	}
	defer rows.Close()

	var parks []core.Park
	for rows.Next() {
		var park core.Park
		var state sql.NullString
		var country sql.NullString

		err := rows.Scan(
			&park.ID,
			&park.Name,
			&park.City,
			&state,
			&country,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan park: %w", err)
		}

		if state.Valid {
			park.State = state.String
		}
		if country.Valid {
			park.Country = country.String
		}

		parks = append(parks, park)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate parks: %w", err)
	}

	return parks, nil
}

// GamesAtPark retrieves games played at a specific park.
// TODO: This should delegate to GameRepository
func (r *ParkRepository) GamesAtPark(ctx context.Context, id core.ParkID, filter core.GameFilter) ([]core.Game, error) {
	return []core.Game{}, nil
}
