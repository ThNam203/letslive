-- +goose Up
ALTER TABLE users ALTER COLUMN email TYPE VARCHAR(320);
ALTER TABLE users ALTER COLUMN profile_picture TYPE VARCHAR(2048);
ALTER TABLE users ALTER COLUMN background_picture TYPE VARCHAR(2048);

ALTER TABLE livestream_information ALTER COLUMN thumbnail_url TYPE VARCHAR(2048);

ALTER TABLE user_social_links ALTER COLUMN url TYPE VARCHAR(2048);

ALTER TABLE notifications ALTER COLUMN action_url TYPE VARCHAR(2048);

-- +goose Down
ALTER TABLE users ALTER COLUMN email TYPE TEXT;
ALTER TABLE users ALTER COLUMN profile_picture TYPE TEXT;
ALTER TABLE users ALTER COLUMN background_picture TYPE TEXT;

ALTER TABLE livestream_information ALTER COLUMN thumbnail_url TYPE TEXT;

ALTER TABLE user_social_links ALTER COLUMN url TYPE TEXT;

ALTER TABLE notifications ALTER COLUMN action_url TYPE TEXT;
