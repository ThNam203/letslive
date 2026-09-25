-- +goose Up
CREATE INDEX IF NOT EXISTS "idx_refresh_tokens_token" ON "refresh_tokens" ("token");

-- +goose Down
DROP INDEX IF EXISTS "idx_refresh_tokens_token";
