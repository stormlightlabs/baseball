package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"stormlightlabs.org/baseball/internal/core"
)

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
		return nil, core.NewNotFoundError("token", "")
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
		return nil, core.NewNotFoundError("token", string(userID))
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
		return core.NewNotFoundError("token", "")
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
