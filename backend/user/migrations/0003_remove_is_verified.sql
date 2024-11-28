-- +goose Up 
ALTER TABLE users DROP COLUMN "is_verified";

-- +goose Down
ALTER TABLE users ADD COLUMN "is_verified" boolean NOT NULL DEFAULT false;
