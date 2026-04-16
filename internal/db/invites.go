package db

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const inviteColumns = `id, organization_id, email, role, hashed_token, created_by,
	created_at, expires_at, accepted_at, accepted_by`

func scanInvite(row pgx.Row) (Invite, error) {
	var inv Invite
	err := row.Scan(
		&inv.ID, &inv.OrganizationID, &inv.Email, &inv.Role, &inv.HashedToken,
		&inv.CreatedBy, &inv.CreatedAt, &inv.ExpiresAt, &inv.AcceptedAt, &inv.AcceptedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invite{}, ErrNotFound
	}
	if err != nil {
		return Invite{}, err
	}
	return inv, nil
}

func (d *DB) CreateInvite(
	ctx context.Context,
	orgID uuid.UUID,
	email, role string,
	hashedToken []byte,
	createdBy uuid.UUID,
	expiresAt time.Time,
) (Invite, error) {
	const q = `
		INSERT INTO invites (organization_id, email, role, hashed_token, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING ` + inviteColumns
	return scanInvite(d.pool.QueryRow(ctx, q, orgID, email, role, hashedToken, createdBy, expiresAt))
}

// GetLiveInviteByHashedToken returns an invite only if it is unaccepted and
// unexpired. Consumers never need to see dead invites.
func (d *DB) GetLiveInviteByHashedToken(ctx context.Context, hashed []byte) (Invite, error) {
	const q = `
		SELECT ` + inviteColumns + ` FROM invites
		WHERE hashed_token = $1
		  AND accepted_at IS NULL
		  AND expires_at > now()`
	return scanInvite(d.pool.QueryRow(ctx, q, hashed))
}

func (d *DB) ListInvitesForOrganization(ctx context.Context, orgID uuid.UUID) ([]Invite, error) {
	const q = `SELECT ` + inviteColumns + ` FROM invites WHERE organization_id = $1 ORDER BY created_at DESC`
	rows, err := d.pool.Query(ctx, q, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Invite, 0)
	for rows.Next() {
		var inv Invite
		if err := rows.Scan(
			&inv.ID, &inv.OrganizationID, &inv.Email, &inv.Role, &inv.HashedToken,
			&inv.CreatedBy, &inv.CreatedAt, &inv.ExpiresAt, &inv.AcceptedAt, &inv.AcceptedBy,
		); err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (d *DB) MarkInviteAccepted(ctx context.Context, id, acceptedBy uuid.UUID) error {
	const q = `UPDATE invites SET accepted_at = now(), accepted_by = $2 WHERE id = $1`
	_, err := d.pool.Exec(ctx, q, id, acceptedBy)
	return err
}

func (d *DB) DeleteInvite(ctx context.Context, id, orgID uuid.UUID) error {
	const q = `DELETE FROM invites WHERE id = $1 AND organization_id = $2`
	_, err := d.pool.Exec(ctx, q, id, orgID)
	return err
}
