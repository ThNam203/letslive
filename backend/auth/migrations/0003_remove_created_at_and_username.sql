-- +goose Up
ALTER TABLE USER DROP COLUMN username;
ALTER TABLE USER DROP COLUMN created_at;

-- +goose Down
ALTER TABLE USER ADD COLUMN username VARCHAR(20) NOT NULL;
ALTER TABLE USER ADD COLUMN created_at TIMESTAMPTZ DEFAULT current_timestamp;
