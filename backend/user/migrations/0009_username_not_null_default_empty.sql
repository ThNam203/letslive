-- +goose Up
UPDATE users SET username = '' WHERE username IS NULL;

ALTER TABLE users ALTER COLUMN username SET DEFAULT '';
ALTER TABLE users ALTER COLUMN username SET NOT NULL;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
CREATE UNIQUE INDEX users_username_unique_nonempty ON users (username) WHERE username <> '';

-- +goose Down
DROP INDEX IF EXISTS users_username_unique_nonempty;
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);

ALTER TABLE users ALTER COLUMN username DROP NOT NULL;
ALTER TABLE users ALTER COLUMN username DROP DEFAULT;
