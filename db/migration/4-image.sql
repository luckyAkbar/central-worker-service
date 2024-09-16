-- +migrate Up notransaction

CREATE TYPE storage_location AS ENUM ('LOCAL');

CREATE TABLE IF NOT EXISTS images (
    id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    file_size_bytes BIGINT NOT NULL,
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    access_key VARCHAR(255) NOT NULL,
    location storage_location NOT NULL
);

-- +migrate Down

DROP TABLE IF EXISTS images;