package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

type UmpireRepository struct {
	db *sql.DB
}

func NewUmpireRepository(db *sql.DB) *UmpireRepository {
	return &UmpireRepository{db: db}
}

// GetByID retrieves an umpire by their ID.
// Since umpires don't have a dedicated table, we extract from games table.
func (r *UmpireRepository) GetByID(ctx context.Context, id core.UmpireID) (*core.Umpire, error) {
	// TODO: move to embedded query
	query := `
		SELECT DISTINCT
			hp_ump_id,
			hp_ump_name
		FROM games
		WHERE hp_ump_id = $1
		UNION
		SELECT DISTINCT
			b1_ump_id,
			b1_ump_name
		FROM games
		WHERE b1_ump_id = $1
		UNION
		SELECT DISTINCT
			b2_ump_id,
			b2_ump_name
		FROM games
		WHERE b2_ump_id = $1
		UNION
		SELECT DISTINCT
			b3_ump_id,
			b3_ump_name
		FROM games
		WHERE b3_ump_id = $1
		UNION
		SELECT DISTINCT
			lf_ump_id,
			lf_ump_name
		FROM games
		WHERE lf_ump_id = $1
		UNION
		SELECT DISTINCT
			rf_ump_id,
			rf_ump_name
		FROM games
		WHERE rf_ump_id = $1
		LIMIT 1
	`

	var umpire core.Umpire
	var fullName sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&umpire.ID,
		&fullName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("umpire not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get umpire: %w", err)
	}

	if fullName.Valid {
		name := fullName.String
		umpire.FirstName = ""
		umpire.LastName = name
	}

	return &umpire, nil
}

// List retrieves all umpires with pagination.
func (r *UmpireRepository) List(ctx context.Context, p core.Pagination) ([]core.Umpire, error) {
	// TODO: move to embedded query
	query := `
		SELECT ump_id, ump_name
		FROM (
			SELECT hp_ump_id AS ump_id, hp_ump_name AS ump_name FROM games WHERE hp_ump_id IS NOT NULL AND hp_ump_id != ''
			UNION
			SELECT b1_ump_id, b1_ump_name FROM games WHERE b1_ump_id IS NOT NULL AND b1_ump_id != ''
			UNION
			SELECT b2_ump_id, b2_ump_name FROM games WHERE b2_ump_id IS NOT NULL AND b2_ump_id != ''
			UNION
			SELECT b3_ump_id, b3_ump_name FROM games WHERE b3_ump_id IS NOT NULL AND b3_ump_id != ''
			UNION
			SELECT lf_ump_id, lf_ump_name FROM games WHERE lf_ump_id IS NOT NULL AND lf_ump_id != ''
			UNION
			SELECT rf_ump_id, rf_ump_name FROM games WHERE rf_ump_id IS NOT NULL AND rf_ump_id != ''
		) AS all_umps
		ORDER BY ump_name
	`

	args := []any{}
	if p.PerPage > 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, p.PerPage, (p.Page-1)*p.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list umpires: %w", err)
	}
	defer rows.Close()

	var umpires []core.Umpire
	for rows.Next() {
		var umpire core.Umpire
		var fullName sql.NullString

		err := rows.Scan(
			&umpire.ID,
			&fullName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan umpire: %w", err)
		}

		if fullName.Valid {
			umpire.FirstName = ""
			umpire.LastName = fullName.String
		}

		umpires = append(umpires, umpire)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate umpires: %w", err)
	}

	return umpires, nil
}

// GamesForUmpire retrieves games where the umpire officiated.
// TODO: This should delegate to GameRepository or be implemented with direct query
func (r *UmpireRepository) GamesForUmpire(ctx context.Context, id core.UmpireID, filter core.GameFilter) ([]core.Game, error) {
	return []core.Game{}, nil
}
