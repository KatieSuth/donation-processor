-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS donations (
    uuid UUID PRIMARY KEY,
    amount BIGINT NOT NULL CHECK (amount >= 0),
    currency TEXT NOT NULL,
    payment_method TEXT NOT NULL CHECK (payment_method IN ('cc', 'ach', 'crypto', 'venmo')),
    nonprofit_id TEXT NOT NULL,
    donor_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('new', 'pending', 'success', 'failure')),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS donations;
