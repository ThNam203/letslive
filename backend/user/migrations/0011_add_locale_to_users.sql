-- +goose Up
ALTER TABLE users ADD COLUMN locale VARCHAR(35);

-- +goose Down
ALTER TABLE users DROP COLUMN locale;
