-- +migrate Up notransaction

ALTER TABLE secret_messaging_sessions ADD COLUMN is_blocked BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down

ALTER TABLE secret_messaging_sessions DROP COLUMN is_blocked;