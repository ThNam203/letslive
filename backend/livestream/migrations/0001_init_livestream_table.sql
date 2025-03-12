-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE livestream_visibility AS ENUM('public', 'private');

CREATE TABLE livestreams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    status VARCHAR(20) NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'public',
    view_count BIGINT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    ended_at TIMESTAMP WITH TIME ZONE,
    playback_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- +goose Down
DROP TABLE livestreams;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TYPE IF EXISTS livestream_visibility;
