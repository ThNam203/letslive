-- +goose Up
CREATE TABLE IF NOT EXISTS disabled_authors (
    user_id UUID PRIMARY KEY,
    disabled_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS disabled_authors;
