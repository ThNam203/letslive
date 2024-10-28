-- +goose Up
ALTER TABLE users DROP COLUMN is_online;
ALTER TABLE users DROP COLUMN stream_api_key;

-- -goose Down
ALTER TABLE users ADD COLUMN is_online boolean;
ALTER TABLE users ADD COLUMN stream_api_key uuid not null default uuid_generate_v4();
