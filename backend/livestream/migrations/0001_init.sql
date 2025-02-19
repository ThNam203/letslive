-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE livestreams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'offline', 
    view_count BIGINT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    playback_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- +goose Down
DROP TABLE livestreams;
DROP EXTENSION IF EXISTS "uuid-ossp";
