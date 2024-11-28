-- +goose Up
ALTER TABLE users DROP CONSTRAINT "uni_users_username";

-- +goose Down
ALTER TABLE users ADD CONSTRAINT "uni_users_username" UNIQUE ("username");
