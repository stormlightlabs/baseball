package repository

import (
	"context"
	"database/sql"
	"fmt"

	"stormlightlabs.org/baseball/internal/cache"
	"stormlightlabs.org/baseball/internal/core"
)

type ManagerRepository struct {
	db    *sql.DB
	cache *cache.CachedRepository
}

func NewManagerRepository(db *sql.DB, cacheClient *cache.Client) *ManagerRepository {
	return &ManagerRepository{
		db:    db,
		cache: cache.NewCachedRepository(cacheClient, "manager"),
	}
}

// GetByID retrieves a manager by their manager ID (playerID).
// Includes extended biographical data from Retrosheet when available.
func (r *ManagerRepository) GetByID(ctx context.Context, id core.ManagerID) (*core.Manager, error) {
	var cached core.Manager
	if r.cache.Entity.Get(ctx, string(id), &cached) {
		return &cached, nil
	}

	query := `
		SELECT
			p."playerID",
			p."nameFirst",
			p."nameLast",
			pbe.debut_manager,
			pbe.last_manager,
			pbe.use_name,
			pbe.full_name
		FROM "People" p
		LEFT JOIN player_bio_extended pbe ON pbe.retro_id = p."retroID"
		WHERE p."playerID" = $1
		  AND EXISTS (
		    SELECT 1 FROM "Managers" m
		    WHERE m."playerID" = p."playerID"
		  )
	`

	var mgr core.Manager
	var debutGame, lastGame sql.NullTime
	var useName, fullName sql.NullString

	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(
		&mgr.ID,
		&mgr.FirstName,
		&mgr.LastName,
		&debutGame,
		&lastGame,
		&useName,
		&fullName,
	)

	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("manager", string(id))
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get manager: %w", err)
	}

	pid := core.PlayerID(mgr.ID)
	mgr.PlayerID = &pid

	if debutGame.Valid {
		mgr.DebutGame = &debutGame.Time
	}
	if lastGame.Valid {
		mgr.LastGame = &lastGame.Time
	}
	if useName.Valid {
		mgr.UseName = &useName.String
	}
	if fullName.Valid {
		mgr.FullName = &fullName.String
	}

	_ = r.cache.Entity.Set(ctx, string(id), &mgr)
	return &mgr, nil
}

// List retrieves all managers with pagination.
// Includes extended biographical data from Retrosheet when available.
func (r *ManagerRepository) List(ctx context.Context, p core.Pagination) ([]core.Manager, error) {
	query := `
		SELECT DISTINCT
			p."playerID",
			p."nameFirst",
			p."nameLast",
			pbe.debut_manager,
			pbe.last_manager,
			pbe.use_name,
			pbe.full_name
		FROM "People" p
		INNER JOIN "Managers" m ON m."playerID" = p."playerID"
		LEFT JOIN player_bio_extended pbe ON pbe.retro_id = p."retroID"
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
		var debutGame, lastGame sql.NullTime
		var useName, fullName sql.NullString

		err := rows.Scan(
			&mgr.ID,
			&mgr.FirstName,
			&mgr.LastName,
			&debutGame,
			&lastGame,
			&useName,
			&fullName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan manager: %w", err)
		}

		pid := core.PlayerID(mgr.ID)
		mgr.PlayerID = &pid

		if debutGame.Valid {
			mgr.DebutGame = &debutGame.Time
		}
		if lastGame.Valid {
			mgr.LastGame = &lastGame.Time
		}
		if useName.Valid {
			mgr.UseName = &useName.String
		}
		if fullName.Valid {
			mgr.FullName = &fullName.String
		}

		managers = append(managers, mgr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate managers: %w", err)
	}

	return managers, nil
}

// SeasonRecords retrieves all season records for a manager.
func (r *ManagerRepository) SeasonRecords(ctx context.Context, id core.ManagerID) ([]core.ManagerSeasonRecord, error) {
	var cached []core.ManagerSeasonRecord
	if r.cache.Entity.Get(ctx, string(id)+":seasons", &cached) {
		return cached, nil
	}

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

	_ = r.cache.Entity.Set(ctx, string(id)+":seasons", records)
	return records, nil
}
