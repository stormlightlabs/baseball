package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/core"
)

type ManagerRepository struct {
	db *sql.DB
}

func NewManagerRepository(db *sql.DB) *ManagerRepository {
	return &ManagerRepository{db: db}
}

// GetByID retrieves a manager by their manager ID (playerID).
func (r *ManagerRepository) GetByID(ctx context.Context, id core.ManagerID) (*core.Manager, error) {
	query := `
		SELECT
			p."playerID",
			p."nameFirst",
			p."nameLast"
		FROM "People" p
		WHERE p."playerID" = $1
		  AND EXISTS (
		    SELECT 1 FROM "Managers" m
		    WHERE m."playerID" = p."playerID"
		  )
	`

	var mgr core.Manager
	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&mgr.ID,
		&mgr.FirstName,
		&mgr.LastName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("manager not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get manager: %w", err)
	}

	// Check if this person was also a player
	var playerID string
	checkPlayerQuery := `SELECT "playerID" FROM "People" WHERE "playerID" = $1`
	if err := r.db.QueryRowContext(ctx, checkPlayerQuery, string(id)).Scan(&playerID); err == nil {
		pid := core.PlayerID(playerID)
		mgr.PlayerID = &pid
	}

	return &mgr, nil
}

// List retrieves all managers with pagination.
func (r *ManagerRepository) List(ctx context.Context, p core.Pagination) ([]core.Manager, error) {
	query := `
		SELECT DISTINCT
			p."playerID",
			p."nameFirst",
			p."nameLast"
		FROM "People" p
		INNER JOIN "Managers" m ON m."playerID" = p."playerID"
		ORDER BY p."nameLast", p."nameFirst"
	`

	args := []any{}
	if p.PerPage > 0 {
		query += " LIMIT $1 OFFSET $2"
		args = append(args, p.PerPage, (p.Page-1)*p.PerPage)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list managers: %w", err)
	}
	defer rows.Close()

	var managers []core.Manager
	for rows.Next() {
		var mgr core.Manager
		err := rows.Scan(
			&mgr.ID,
			&mgr.FirstName,
			&mgr.LastName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manager: %w", err)
		}

		pid := core.PlayerID(mgr.ID)
		mgr.PlayerID = &pid

		managers = append(managers, mgr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate managers: %w", err)
	}

	return managers, nil
}

// SeasonRecords retrieves all season records for a manager.
func (r *ManagerRepository) SeasonRecords(ctx context.Context, id core.ManagerID) ([]core.ManagerSeasonRecord, error) {
	query := `
		SELECT
			m."playerID",
			m."teamID",
			m."yearID",
			m."G",
			m."W",
			m."L",
			m."rank"
		FROM "Managers" m
		WHERE m."playerID" = $1
		ORDER BY m."yearID" DESC, m."inseason"
	`

	rows, err := r.db.QueryContext(ctx, query, string(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get manager season records: %w", err)
	}
	defer rows.Close()

	var records []core.ManagerSeasonRecord
	for rows.Next() {
		var rec core.ManagerSeasonRecord
		var rank sql.NullInt64

		err := rows.Scan(
			&rec.ManagerID,
			&rec.TeamID,
			&rec.Year,
			&rec.G,
			&rec.W,
			&rec.L,
			&rank,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manager season record: %w", err)
		}

		if rank.Valid {
			r := int(rank.Int64)
			rec.Rank = &r
		}

		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate manager records: %w", err)
	}

	return records, nil
}
