-- +migrate Up notransaction

CREATE TYPE "subscription_type" AS ENUM ('meme');
CREATE TYPE "subscription_channel" AS ENUM  ('telegram');

CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(255) PRIMARY KEY,
    "type" subscription_type NOT NULL,
    "channel" subscription_channel NOT NULL,
    user_reference_id VARCHAR(255) NOT NULL -- not a foreign key to ensure any user from any service / channel can register 
);

-- +migrate Down

DROP TABLE IF EXISTS subscriptions;