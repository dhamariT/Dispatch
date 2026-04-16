-- invites carries single-use tokens for adding teammates to an org.
-- Only hashed_token is stored. The raw token is handed to the inviter once
-- (in the response) and embedded in the share link; the server never sees
-- it again until acceptance, at which point we hash the incoming value and
-- look it up.
CREATE TABLE invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL,
    hashed_token BYTEA NOT NULL UNIQUE,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    accepted_by UUID REFERENCES users(id),
    CONSTRAINT invites_role_check
        CHECK (role IN ('admin', 'operator', 'viewer'))
);

CREATE INDEX idx_invites_organization_id ON invites(organization_id);
