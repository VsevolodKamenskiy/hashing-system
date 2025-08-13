-- +goose Up
CREATE TABLE IF NOT EXISTS hashes (
    id   BIGSERIAL PRIMARY KEY,
    hash TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS hashes;
