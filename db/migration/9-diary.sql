-- +migrate Up notransaction

CREATE TYPE diary_source AS ENUM('telegram');

CREATE TABLE IF NOT EXISTS diaries (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    note TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    time_zone VARCHAR(100) NOT NULL,
    source diary_source NOT NULL
);

-- +migrate Down

DROP TABLE IF EXISTS diaries;