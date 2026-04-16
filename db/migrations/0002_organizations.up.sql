CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Roles are stored as a plain text column guarded by a CHECK constraint.
-- We intentionally avoid a Postgres ENUM so adding a new role later is
-- an ALTER TABLE on a constraint, not a type migration.
CREATE TABLE organization_members (
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, user_id),
    CONSTRAINT organization_members_role_check
        CHECK (role IN ('admin', 'operator', 'viewer'))
);

CREATE INDEX idx_organization_members_user_id ON organization_members(user_id);
