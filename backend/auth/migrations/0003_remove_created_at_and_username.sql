-- +goose Up
ALTER TABLE users DROP COLUMN username;
ALTER TABLE users DROP COLUMN created_at;

-- +goose Down
ALTER TABLE users ADD COLUMN username VARCHAR(20) NOT NULL;
ALTER TABLE users ADD COLUMN created_at TIMESTAMPTZ DEFAULT current_timestamp;
