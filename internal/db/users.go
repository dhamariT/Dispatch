package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("not found")

const userColumns = `id, email, login, name, avatar_url, created_at, updated_at`

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.Login, &u.Name, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

func (d *DB) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	return scanUser(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) GetUserByEmail(ctx context.Context, email string) (User, error) {
	const q = `SELECT ` + userColumns + ` FROM users WHERE email = $1`
	return scanUser(d.pool.QueryRow(ctx, q, email))
}

func (d *DB) CreateUser(ctx context.Context, email, login, name, avatarURL string) (User, error) {
	const q = `
		INSERT INTO users (email, login, name, avatar_url)
		VALUES ($1, $2, $3, $4)
		RETURNING ` + userColumns
	return scanUser(d.pool.QueryRow(ctx, q, email, login, name, avatarURL))
}

func (d *DB) UpdateUserProfile(ctx context.Context, id uuid.UUID, login, name, avatarURL string) (User, error) {
	const q = `
		UPDATE users
		SET login = $2, name = $3, avatar_url = $4, updated_at = now()
		WHERE id = $1
		RETURNING ` + userColumns
	return scanUser(d.pool.QueryRow(ctx, q, id, login, name, avatarURL))
}

// FindUserByIdentity looks up a user via (provider, provider_user_id) in the
// user_identities join table. This is the primary GitHub → local user path.
func (d *DB) FindUserByIdentity(ctx context.Context, provider, providerUserID string) (User, error) {
	const q = `
		SELECT u.id, u.email, u.login, u.name, u.avatar_url, u.created_at, u.updated_at
		FROM users u
		JOIN user_identities i ON i.user_id = u.id
		WHERE i.provider = $1 AND i.provider_user_id = $2`
	return scanUser(d.pool.QueryRow(ctx, q, provider, providerUserID))
}

func (d *DB) LinkUserIdentity(ctx context.Context, provider, providerUserID string, userID uuid.UUID) error {
	const q = `
		INSERT INTO user_identities (provider, provider_user_id, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, provider_user_id) DO NOTHING`
	_, err := d.pool.Exec(ctx, q, provider, providerUserID, userID)
	if err != nil {
		return fmt.Errorf("link identity: %w", err)
	}
	return nil
}
