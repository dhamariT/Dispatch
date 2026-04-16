package db

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const orgColumns = `id, slug, name, created_at, updated_at`

func scanOrg(row pgx.Row) (Organization, error) {
	var o Organization
	err := row.Scan(&o.ID, &o.Slug, &o.Name, &o.CreatedAt, &o.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Organization{}, ErrNotFound
	}
	if err != nil {
		return Organization{}, err
	}
	return o, nil
}

func (d *DB) CreateOrganization(ctx context.Context, slug, name string) (Organization, error) {
	const q = `INSERT INTO organizations (slug, name) VALUES ($1, $2) RETURNING ` + orgColumns
	return scanOrg(d.pool.QueryRow(ctx, q, slug, name))
}

func (d *DB) GetOrganizationBySlug(ctx context.Context, slug string) (Organization, error) {
	const q = `SELECT ` + orgColumns + ` FROM organizations WHERE slug = $1`
	return scanOrg(d.pool.QueryRow(ctx, q, slug))
}

func (d *DB) GetOrganizationByID(ctx context.Context, id uuid.UUID) (Organization, error) {
	const q = `SELECT ` + orgColumns + ` FROM organizations WHERE id = $1`
	return scanOrg(d.pool.QueryRow(ctx, q, id))
}

func (d *DB) ListOrganizationsForUser(ctx context.Context, userID uuid.UUID) ([]OrganizationWithRole, error) {
	const q = `
		SELECT o.id, o.slug, o.name, o.created_at, o.updated_at, m.role
		FROM organizations o
		JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.created_at ASC`
	rows, err := d.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]OrganizationWithRole, 0)
	for rows.Next() {
		var o OrganizationWithRole
		if err := rows.Scan(&o.ID, &o.Slug, &o.Name, &o.CreatedAt, &o.UpdatedAt, &o.Role); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (d *DB) AddOrganizationMember(ctx context.Context, orgID, userID uuid.UUID, role string) (OrganizationMember, error) {
	const q = `
		INSERT INTO organization_members (organization_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (organization_id, user_id) DO UPDATE
		    SET role = EXCLUDED.role
		RETURNING organization_id, user_id, role, created_at`
	var m OrganizationMember
	err := d.pool.QueryRow(ctx, q, orgID, userID, role).Scan(&m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt)
	return m, err
}

func (d *DB) GetOrganizationMembership(ctx context.Context, orgID, userID uuid.UUID) (OrganizationMember, error) {
	const q = `
		SELECT organization_id, user_id, role, created_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2`
	var m OrganizationMember
	err := d.pool.QueryRow(ctx, q, orgID, userID).Scan(&m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrganizationMember{}, ErrNotFound
	}
	return m, err
}

func (d *DB) ListOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]OrganizationMemberWithUser, error) {
	const q = `
		SELECT m.organization_id, m.user_id, m.role, m.created_at,
		       u.email, u.login, u.name, u.avatar_url
		FROM organization_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.organization_id = $1
		ORDER BY m.created_at ASC`
	rows, err := d.pool.Query(ctx, q, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]OrganizationMemberWithUser, 0)
	for rows.Next() {
		var m OrganizationMemberWithUser
		if err := rows.Scan(
			&m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt,
			&m.Email, &m.Login, &m.Name, &m.AvatarURL,
		); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (d *DB) RemoveOrganizationMember(ctx context.Context, orgID, userID uuid.UUID) error {
	const q = `DELETE FROM organization_members WHERE organization_id = $1 AND user_id = $2`
	_, err := d.pool.Exec(ctx, q, orgID, userID)
	return err
}
