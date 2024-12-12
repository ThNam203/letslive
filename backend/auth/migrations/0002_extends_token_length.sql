-- +goose Up
ALTER TABLE "refresh_tokens" ALTER COLUMN value TYPE varchar(1000);
ALTER TABLE "refresh_tokens" ALTER COLUMN value SET NOT NULL;

-- +goose Down
ALTER TABLE "refresh_tokens" ALTER COLUMN value TYPE varchar(255) USING substr("value", 1, 255);
ALTER TABLE "refresh_tokens" ALTER COLUMN value SET NOT NULL;
