package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

type UsageRepository struct {
	db *sql.DB
}

func NewUsageRepository(db *sql.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) Record(ctx context.Context, userID *core.UserID, apiKeyID *string, endpoint string, method string, statusCode int, responseTimeMs *int) error {
	query := `
		INSERT INTO api_usage (user_id, api_key_id, endpoint, method, status_code, response_time_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	var dbUserID, dbAPIKeyID sql.NullString
	var dbResponseTimeMs sql.NullInt64

	if userID != nil {
		dbUserID.String = string(*userID)
		dbUserID.Valid = true
	}
	if apiKeyID != nil {
		dbAPIKeyID.String = *apiKeyID
		dbAPIKeyID.Valid = true
	}
	if responseTimeMs != nil {
		dbResponseTimeMs.Int64 = int64(*responseTimeMs)
		dbResponseTimeMs.Valid = true
	}

	_, err := r.db.ExecContext(ctx, query, dbUserID, dbAPIKeyID, endpoint, method, statusCode, dbResponseTimeMs)
	if err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	return nil
}

func (r *UsageRepository) GetUserUsage(ctx context.Context, userID core.UserID, since time.Time) ([]core.APIUsage, error) {
	query := `
		SELECT id, user_id, api_key_id, endpoint, method, status_code, response_time_ms, created_at
		FROM api_usage
		WHERE user_id = $1 AND created_at >= $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get user usage: %w", err)
	}
	defer rows.Close()

	var usageList []core.APIUsage
	for rows.Next() {
		var usage core.APIUsage
		var dbUserID, apiKeyID sql.NullString
		var responseTimeMs sql.NullInt64

		err := rows.Scan(
			&usage.ID,
			&dbUserID,
			&apiKeyID,
			&usage.Endpoint,
			&usage.Method,
			&usage.StatusCode,
			&responseTimeMs,
			&usage.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage: %w", err)
		}

		if dbUserID.Valid {
			uid := core.UserID(dbUserID.String)
			usage.UserID = &uid
		}
		if apiKeyID.Valid {
			usage.APIKeyID = &apiKeyID.String
		}
		if responseTimeMs.Valid {
			ms := int(responseTimeMs.Int64)
			usage.ResponseTimeMs = &ms
		}

		usageList = append(usageList, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate usage: %w", err)
	}

	return usageList, nil
}

func (r *UsageRepository) GetAPIKeyUsage(ctx context.Context, apiKeyID string, since time.Time) ([]core.APIUsage, error) {
	query := `
		SELECT id, user_id, api_key_id, endpoint, method, status_code, response_time_ms, created_at
		FROM api_usage
		WHERE api_key_id = $1 AND created_at >= $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, apiKeyID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key usage: %w", err)
	}
	defer rows.Close()

	var usageList []core.APIUsage
	for rows.Next() {
		var usage core.APIUsage
		var userID, dbAPIKeyID sql.NullString
		var responseTimeMs sql.NullInt64

		err := rows.Scan(
			&usage.ID,
			&userID,
			&dbAPIKeyID,
			&usage.Endpoint,
			&usage.Method,
			&usage.StatusCode,
			&responseTimeMs,
			&usage.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage: %w", err)
		}

		if userID.Valid {
			uid := core.UserID(userID.String)
			usage.UserID = &uid
		}
		if dbAPIKeyID.Valid {
			usage.APIKeyID = &dbAPIKeyID.String
		}
		if responseTimeMs.Valid {
			ms := int(responseTimeMs.Int64)
			usage.ResponseTimeMs = &ms
		}

		usageList = append(usageList, usage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate usage: %w", err)
	}

	return usageList, nil
}
