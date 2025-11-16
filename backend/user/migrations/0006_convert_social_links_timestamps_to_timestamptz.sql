-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_social_links
ALTER COLUMN created_at TYPE TIMESTAMPTZ USING created_at AT TIME ZONE 'UTC',
ALTER COLUMN updated_at TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE 'UTC';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_social_links
ALTER COLUMN created_at TYPE TIMESTAMP,
ALTER COLUMN updated_at TYPE TIMESTAMP;
-- +goose StatementEnd

