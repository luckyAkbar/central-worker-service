-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS secret_message_nodes (
    id BIGINT PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    text TEXT DEFAULT '',
    previous_secret_message_id BIGINT DEFAULT NULL
);

ALTER TABLE secret_message_nodes ADD FOREIGN KEY (session_id) REFERENCES "secret_messaging_sessions" (id);
ALTER TABLE secret_message_nodes ADD FOREIGN KEY (previous_secret_message_id) REFERENCES "secret_message_nodes" (id);

-- +migrate Down

DROP TABLE IF EXISTS secret_message_nodes;