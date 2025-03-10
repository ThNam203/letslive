-- +goose Up
ALTER TABLE livestreams ADD COLUMN duration BIGINT;

-- +goose Down
ALTER TABLE livestreams DROP COLUMN duration;
