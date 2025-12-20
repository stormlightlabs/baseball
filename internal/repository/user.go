package repository

import (
	"context"
	"database/sql"
	"fmt"

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
		return nil, core.NewNotFoundError("user", string(id))
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
		return nil, core.NewNotFoundError("user", email)
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
		return core.NewNotFoundError("user", "")
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
		return core.NewNotFoundError("user", "")
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
