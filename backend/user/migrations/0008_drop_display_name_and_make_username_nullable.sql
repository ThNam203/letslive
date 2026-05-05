-- +goose Up
UPDATE users
SET username = LEFT(display_name, 30)
WHERE display_name IS NOT NULL
  AND display_name <> '';

ALTER TABLE users DROP COLUMN display_name;
ALTER TABLE users ALTER COLUMN username DROP NOT NULL;
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(30);

-- +goose Down
ALTER TABLE users ALTER COLUMN username TYPE VARCHAR(50);
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
ALTER TABLE users ADD COLUMN display_name VARCHAR(50) UNIQUE;
