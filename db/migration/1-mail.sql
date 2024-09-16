-- +migrate Up notransaction

CREATE TYPE mail_status AS ENUM('SUCCESS', 'FAILED', 'ON_PROGRESS');

CREATE TABLE IF NOT EXISTS mails (
    id VARCHAR(255) PRIMARY KEY,
    "to" TEXT NOT NULL,
    cc TEXT DEFAULT NULL,
    bcc TEXT DEFAULT NULL,
    html_content TEXT NOT NULL,
    subject VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    delivered_at TIMESTAMPTZ DEFAULT NULL,
    metadata TEXT DEFAULT NULL,
    status mail_status DEFAULT 'ON_PROGRESS'::mail_status
);

-- +migrate Down

DROP TABLE IF EXISTS mails;
