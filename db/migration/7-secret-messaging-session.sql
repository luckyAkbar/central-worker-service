-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS secret_messaging_sessions (
    id VARCHAR(255) PRIMARY KEY,
    sender_id BIGSERIAL NOT NULL,
    target_id BIGSERIAL NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expired_at TIMESTAMPTZ NOT NULL
);

ALTER TABLE secret_messaging_sessions ADD FOREIGN KEY (sender_id) REFERENCES "telegram_users" (id);
ALTER TABLE secret_messaging_sessions ADD FOREIGN KEY (target_id) REFERENCES "telegram_users" (id);
ALTER TABLE secret_messaging_sessions ADD CONSTRAINT sender_id_target_id_must_not_same_check CHECK (sender_id != target_id);

-- +migrate Down

DROP TABLE IF EXISTS secret_messaging_sessions;