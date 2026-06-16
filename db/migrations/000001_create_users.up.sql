CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(20) NOT NULL,
    display_name VARCHAR(50) NOT NULL,
    bio VARCHAR(500) NOT NULL DEFAULT '',
    timezone VARCHAR(64) NOT NULL,
    role VARCHAR(16) NOT NULL DEFAULT 'user',
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT users_username_format_chk
        CHECK (username ~ '^[a-z]{1,20}$'),
    CONSTRAINT users_role_chk
        CHECK (role IN ('user', 'moderator', 'admin')),
    CONSTRAINT users_status_chk
        CHECK (status IN ('active', 'blocked'))
);

CREATE UNIQUE INDEX users_username_unique_idx ON users (username);
CREATE INDEX users_status_idx ON users (status);

CREATE TABLE audit_log (
    id UUID PRIMARY KEY,
    actor_id UUID NOT NULL REFERENCES users (id),
    target_id UUID NOT NULL REFERENCES users (id),
    action VARCHAR(64) NOT NULL,
    old_value TEXT NOT NULL,
    new_value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX audit_log_target_created_idx
    ON audit_log (target_id, created_at DESC);
