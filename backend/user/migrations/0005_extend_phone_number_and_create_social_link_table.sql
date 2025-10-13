-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
ALTER COLUMN phone_number TYPE character varying(20);

CREATE TABLE IF NOT EXISTS user_social_links (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    platform VARCHAR(50) NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (user_id, platform)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_social_links;

ALTER TABLE users
ALTER COLUMN phone_number TYPE character varying(15);
-- +goose StatementEnd

