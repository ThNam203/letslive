-- +goose Up
ALTER TABLE users ADD COLUMN user_id uuid;
ALTER TABLE users ADD CONSTRAINT uni_user_id UNIQUE ("user_id");

-- +goose Down
ALTER TABLE users DROP COLUMN user_id;
ALTER TABLE users DROP CONSTRAINT uni_user_id;
