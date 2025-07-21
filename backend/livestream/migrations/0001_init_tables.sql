-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ENUM Types
CREATE TYPE livestream_visibility AS ENUM ('public', 'private');
CREATE TYPE vod_visibility AS ENUM ('public', 'private');

-- Livestreams Table
CREATE TABLE livestreams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    visibility livestream_visibility NOT NULL DEFAULT 'public',
    view_count BIGINT DEFAULT 0,
    started_at TIMESTAMPTZ DEFAULT now(),
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- VODs Table
CREATE TABLE vods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    livestream_id UUID NOT NULL,
    user_id UUID NOT NULL,
    title VARCHAR(255),
    description TEXT,
    thumbnail_url VARCHAR(1024),
    visibility vod_visibility NOT NULL DEFAULT 'private',
    view_count BIGINT NOT NULL DEFAULT 0,
    duration BIGINT NOT NULL,
    playback_url VARCHAR(2048),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add foreign keys after both tables are created
ALTER TABLE livestreams
ADD COLUMN vod_id UUID,
ADD CONSTRAINT fk_livestream_vod
    FOREIGN KEY (vod_id)
    REFERENCES vods(id)
    ON DELETE SET NULL;

ALTER TABLE vods
ADD CONSTRAINT fk_vod_livestream
    FOREIGN KEY (livestream_id)
    REFERENCES livestreams(id)
    ON DELETE SET NULL;

-- Indexes
CREATE INDEX idx_vod_livestream_id ON vods(livestream_id);
CREATE INDEX idx_vod_user_id ON vods(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_vod_user_id;
DROP INDEX IF EXISTS idx_vod_livestream_id;

ALTER TABLE vods DROP CONSTRAINT IF EXISTS fk_vod_livestream;
ALTER TABLE livestreams DROP CONSTRAINT IF EXISTS fk_livestream_vod;
ALTER TABLE livestreams DROP COLUMN IF EXISTS vod_id;

DROP TABLE IF EXISTS vods;
DROP TABLE IF EXISTS livestreams;

DROP TYPE IF EXISTS vod_visibility;
DROP TYPE IF EXISTS livestream_visibility;
DROP EXTENSION IF EXISTS "uuid-ossp";
