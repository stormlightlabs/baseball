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

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id core.UserID) (*core.User, error) {
	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE id = $1
	`

	var user core.User
	var name, avatarURL sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&name,
		&avatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
		&user.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if name.Valid {
		user.Name = &name.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*core.User, error) {
	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at, last_login_at, is_active
		FROM users
		WHERE email = $1
	`

	var user core.User
	var name, avatarURL sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&name,
		&avatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
		&user.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if name.Valid {
		user.Name = &name.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *UserRepository) Create(ctx context.Context, email string, name *string, avatarURL *string) (*core.User, error) {
	query := `
		INSERT INTO users (email, name, avatar_url, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, NOW(), NOW(), true)
		RETURNING id, email, name, avatar_url, created_at, updated_at, last_login_at, is_active
	`

	var user core.User
	var dbName, dbAvatarURL sql.NullString
	var lastLoginAt sql.NullTime

	if name != nil {
		dbName.String = *name
		dbName.Valid = true
	}
	if avatarURL != nil {
		dbAvatarURL.String = *avatarURL
		dbAvatarURL.Valid = true
	}

	err := r.db.QueryRowContext(ctx, query, email, dbName, dbAvatarURL).Scan(
		&user.ID,
		&user.Email,
		&dbName,
		&dbAvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
		&user.IsActive,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if dbName.Valid {
		user.Name = &dbName.String
	}
	if dbAvatarURL.Valid {
		user.AvatarURL = &dbAvatarURL.String
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *core.User) error {
	query := `
		UPDATE users
		SET name = $2, avatar_url = $3, updated_at = NOW(), is_active = $4
		WHERE id = $1
	`

	var name, avatarURL sql.NullString
	if user.Name != nil {
		name.String = *user.Name
		name.Valid = true
	}
	if user.AvatarURL != nil {
		avatarURL.String = *user.AvatarURL
		avatarURL.Valid = true
	}

	result, err := r.db.ExecContext(ctx, query, user.ID, name, avatarURL, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id core.UserID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) List(ctx context.Context, p core.Pagination) ([]core.User, error) {
	query := `
		SELECT id, email, name, avatar_url, created_at, updated_at, last_login_at, is_active
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	offset := (p.Page - 1) * p.PerPage
	rows, err := r.db.QueryContext(ctx, query, p.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []core.User
	for rows.Next() {
		var user core.User
		var name, avatarURL sql.NullString
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&name,
			&avatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if name.Valid {
			user.Name = &name.String
		}
		if avatarURL.Valid {
			user.AvatarURL = &avatarURL.String
		}
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

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
		return nil, fmt.Errorf("API key not found: %w", err)
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
		return nil, fmt.Errorf("API key not found: %w", err)
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
		return fmt.Errorf("API key not found")
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

type OAuthTokenRepository struct {
	db *sql.DB
}

func NewOAuthTokenRepository(db *sql.DB) *OAuthTokenRepository {
	return &OAuthTokenRepository{db: db}
}

func (r *OAuthTokenRepository) Create(ctx context.Context, userID core.UserID, accessToken string, refreshToken *string, expiresAt time.Time) (*core.OAuthToken, error) {
	query := `
		INSERT INTO oauth_tokens (user_id, access_token, refresh_token, token_type, expires_at, created_at)
		VALUES ($1, $2, $3, 'Bearer', $4, NOW())
		RETURNING id, user_id, access_token, refresh_token, token_type, expires_at, created_at
	`

	var token core.OAuthToken
	var dbRefreshToken sql.NullString

	if refreshToken != nil {
		dbRefreshToken.String = *refreshToken
		dbRefreshToken.Valid = true
	}

	err := r.db.QueryRowContext(ctx, query, userID, accessToken, dbRefreshToken, expiresAt).Scan(
		&token.ID,
		&token.UserID,
		&token.AccessToken,
		&dbRefreshToken,
		&token.TokenType,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create OAuth token: %w", err)
	}

	if dbRefreshToken.Valid {
		token.RefreshToken = &dbRefreshToken.String
	}

	return &token, nil
}

func (r *OAuthTokenRepository) GetByAccessToken(ctx context.Context, accessToken string) (*core.OAuthToken, error) {
	query := `
		SELECT id, user_id, access_token, refresh_token, token_type, expires_at, created_at
		FROM oauth_tokens
		WHERE access_token = $1 AND expires_at > NOW()
	`

	var token core.OAuthToken
	var refreshToken sql.NullString

	err := r.db.QueryRowContext(ctx, query, accessToken).Scan(
		&token.ID,
		&token.UserID,
		&token.AccessToken,
		&refreshToken,
		&token.TokenType,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found or expired: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	if refreshToken.Valid {
		token.RefreshToken = &refreshToken.String
	}

	return &token, nil
}

func (r *OAuthTokenRepository) GetByUserID(ctx context.Context, userID core.UserID) (*core.OAuthToken, error) {
	query := `
		SELECT id, user_id, access_token, refresh_token, token_type, expires_at, created_at
		FROM oauth_tokens
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var token core.OAuthToken
	var refreshToken sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&token.ID,
		&token.UserID,
		&token.AccessToken,
		&refreshToken,
		&token.TokenType,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("token not found: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	if refreshToken.Valid {
		token.RefreshToken = &refreshToken.String
	}

	return &token, nil
}

func (r *OAuthTokenRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM oauth_tokens WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("token not found")
	}

	return nil
}

func (r *OAuthTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM oauth_tokens WHERE expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

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
