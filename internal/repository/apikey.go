package repository

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

type APIKeyRepository struct {
	db *sql.DB
}

func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// generateAPIKey creates a cryptographically secure random API key
func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sk_" + base64.URLEncoding.EncodeToString(b), nil
}

// hashAPIKey creates a SHA-256 hash of the API key for storage
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return base64.URLEncoding.EncodeToString(hash[:])
}

func (r *APIKeyRepository) Create(ctx context.Context, userID core.UserID, name *string, expiresAt *time.Time) (*core.APIKey, string, error) {
	key, err := generateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	keyHash := hashAPIKey(key)
	keyPrefix := key[:10]

	query := `
		INSERT INTO api_keys (user_id, key_hash, key_prefix, name, expires_at, created_at, is_active)
		VALUES ($1, $2, $3, $4, $5, NOW(), true)
		RETURNING id, user_id, key_prefix, name, created_at, last_used_at, expires_at, is_active
	`

	var apiKey core.APIKey
	var dbName sql.NullString
	var lastUsedAt, dbExpiresAt sql.NullTime

	if name != nil {
		dbName.String = *name
		dbName.Valid = true
	}
	if expiresAt != nil {
		dbExpiresAt.Time = *expiresAt
		dbExpiresAt.Valid = true
	}

	err = r.db.QueryRowContext(ctx, query, userID, keyHash, keyPrefix, dbName, dbExpiresAt).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyPrefix,
		&dbName,
		&apiKey.CreatedAt,
		&lastUsedAt,
		&dbExpiresAt,
		&apiKey.IsActive,
	)

	if err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	if dbName.Valid {
		apiKey.Name = &dbName.String
	}
	if lastUsedAt.Valid {
		apiKey.LastUsedAt = &lastUsedAt.Time
	}
	if dbExpiresAt.Valid {
		apiKey.ExpiresAt = &dbExpiresAt.Time
	}

	return &apiKey, key, nil
}

func (r *APIKeyRepository) GetByKey(ctx context.Context, key string) (*core.APIKey, error) {
	keyHash := hashAPIKey(key)

	query := `
		SELECT id, user_id, key_prefix, name, created_at, last_used_at, expires_at, is_active
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true
	`

	var apiKey core.APIKey
	var name sql.NullString
	var lastUsedAt, expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyPrefix,
		&name,
		&apiKey.CreatedAt,
		&lastUsedAt,
		&expiresAt,
		&apiKey.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("API key", "")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if name.Valid {
		apiKey.Name = &name.String
	}
	if lastUsedAt.Valid {
		apiKey.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
		if time.Now().After(*apiKey.ExpiresAt) {
			return nil, fmt.Errorf("API key has expired")
		}
	}

	return &apiKey, nil
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id string) (*core.APIKey, error) {
	query := `
		SELECT id, user_id, key_prefix, name, created_at, last_used_at, expires_at, is_active
		FROM api_keys
		WHERE id = $1
	`

	var apiKey core.APIKey
	var name sql.NullString
	var lastUsedAt, expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.KeyPrefix,
		&name,
		&apiKey.CreatedAt,
		&lastUsedAt,
		&expiresAt,
		&apiKey.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, core.NewNotFoundError("API key", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	if name.Valid {
		apiKey.Name = &name.String
	}
	if lastUsedAt.Valid {
		apiKey.LastUsedAt = &lastUsedAt.Time
	}
	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}

	return &apiKey, nil
}

func (r *APIKeyRepository) ListByUser(ctx context.Context, userID core.UserID) ([]core.APIKey, error) {
	query := `
		SELECT id, user_id, key_prefix, name, created_at, last_used_at, expires_at, is_active
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []core.APIKey
	for rows.Next() {
		var apiKey core.APIKey
		var name sql.NullString
		var lastUsedAt, expiresAt sql.NullTime

		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.KeyPrefix,
			&name,
			&apiKey.CreatedAt,
			&lastUsedAt,
			&expiresAt,
			&apiKey.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}

		if name.Valid {
			apiKey.Name = &name.String
		}
		if lastUsedAt.Valid {
			apiKey.LastUsedAt = &lastUsedAt.Time
		}
		if expiresAt.Valid {
			apiKey.ExpiresAt = &expiresAt.Time
		}

		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate API keys: %w", err)
	}

	return apiKeys, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to revoke API key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return core.NewNotFoundError("API key", "")
	}

	return nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	return nil
}
