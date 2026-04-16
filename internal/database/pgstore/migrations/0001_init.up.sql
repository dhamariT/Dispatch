-- 0001_init.up.sql
-- Bootstrap schema matching every type in database.Store.

CREATE TABLE IF NOT EXISTS experiments (
    id              TEXT PRIMARY KEY,
    deploy_id       TEXT        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'collecting',
    decision        TEXT        NOT NULL DEFAULT '',
    hold_reason     TEXT        NOT NULL DEFAULT '',
    canary_devices  TEXT[]      NOT NULL,
    control_devices TEXT[]      NOT NULL,
    window_minutes  INT         NOT NULL DEFAULT 5,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    results         JSONB       NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS samples (
    id              TEXT PRIMARY KEY,
    experiment_id   TEXT        NOT NULL REFERENCES experiments(id),
    device_id       TEXT        NOT NULL,
    metric_name     TEXT        NOT NULL,
    "group"         TEXT        NOT NULL,
    value           DOUBLE PRECISION NOT NULL,
    ts              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_samples_experiment
    ON samples(experiment_id);

CREATE TABLE IF NOT EXISTS metric_directions (
    experiment_id TEXT NOT NULL REFERENCES experiments(id),
    metric_name   TEXT NOT NULL,
    direction     INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (experiment_id, metric_name)
);

CREATE TABLE IF NOT EXISTS devices (
    id          TEXT PRIMARY KEY,
    name        TEXT        NOT NULL DEFAULT '',
    fleet       TEXT        NOT NULL DEFAULT '',
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id           TEXT PRIMARY KEY,
    hashed_key   TEXT        NOT NULL,
    subject      TEXT        NOT NULL,
    device_id    TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS audit_log (
    id          TEXT PRIMARY KEY,
    ts          TIMESTAMPTZ NOT NULL DEFAULT now(),
    subject_id  TEXT        NOT NULL,
    action      TEXT        NOT NULL,
    target_type TEXT        NOT NULL,
    target_id   TEXT        NOT NULL,
    metadata    JSONB       NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_audit_log_ts
    ON audit_log(ts DESC);
