-- +migrate Up notransaction\

CREATE TABLE IF NOT EXISTS siakadu_scraping_results (
    id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    location storage_location NOT NULL
);

-- +migrate Down

DROP TABLE IF EXISTS siakadu_scraping_results;