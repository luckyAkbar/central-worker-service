-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    access_token_expired_at TIMESTAMPTZ NOT NULL,
    refresh_token_expired_at TIMESTAMPTZ NOT NULL,
    access_token VARCHAR(255) NOT NULL,
    refresh_token VARCHAR(255) NOT NULL
);

ALTER TABLE sessions ADD CONSTRAINT session_user_id_access_refresh_token_unique UNIQUE (user_id, access_token, refresh_token);

-- +migrate Down

DROP TABLES IF EXISTS sessions;