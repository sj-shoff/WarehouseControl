-- +goose Up
CREATE TABLE IF NOT EXISTS apps (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    secret TEXT NOT NULL UNIQUE
);

-- +goose Down
DROP TABLE IF EXISTS apps;