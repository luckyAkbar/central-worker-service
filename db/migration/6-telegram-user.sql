-- +migrate Up notransaction

CREATE TABLE IF NOT EXISTS telegram_users (
    id BIGSERIAL PRIMARY KEY,
    is_bot BOOLEAN NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) DEFAULT NULL,
    username VARCHAR(255) DEFAULT NULL,
    language_code VARCHAR(25) DEFAULT NULL,
    is_premium BOOLEAN NOT NULL
);

-- +migrate Down

DROP TABLE IF EXISTS telegram_users;