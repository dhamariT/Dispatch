package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const sessionColumns = `id, user_id, hashed_secret, created_at, expires_at, last_used_at`

func scanSession(row pgx.Row) (Session, error) {
	var s Session
	err := row.Scan(&s.ID, &s.UserID, &s.HashedSecret, &s.CreatedAt, &s.ExpiresAt, &s.LastUsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, err
	}
	return s, nil
}

func (d *DB) CreateSession(ctx context.Context, userID uuid.UUID, hashedSecret []byte, expiresAt time.Time) (Session, error) {
	const q = `
		INSERT INTO sessions (user_id, hashed_secret, expires_at)
		VALUES ($1, $2, $3)
		RETURNING ` + sessionColumns
	return scanSession(d.pool.QueryRow(ctx, q, userID, hashedSecret, expiresAt))
}

func (d *DB) GetSessionByID(ctx context.Context, id uuid.UUID) (Session, error) {
	const q = `SELECT ` + sessionColumns + ` FROM sessions WHERE id = $1`
	return scanSession(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) TouchSession(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE sessions SET last_used_at = now() WHERE id = $1`
	_, err := d.pool.Exec(ctx, q, id)
	return err
}

func (d *DB) DeleteSession(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM sessions WHERE id = $1`
	_, err := d.pool.Exec(ctx, q, id)
	return err
}

func (d *DB) DeleteExpiredSessions(ctx context.Context) error {
	const q = `DELETE FROM sessions WHERE expires_at < now()`
	_, err := d.pool.Exec(ctx, q)
	return err
}
