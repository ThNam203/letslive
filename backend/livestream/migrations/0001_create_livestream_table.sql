-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE livestream_visibility AS ENUM('public', 'private');

CREATE TABLE livestreams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    thumbnail_url TEXT,
    visibility livestream_visibility NOT NULL DEFAULT 'public',
    view_count BIGINT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
    vod_id UUID,

    CONSTRAINT fk_livestream_vod
        FOREIGN KEY (vod_id)
        REFERENCES vod(id)
        ON DELETE SET NULL
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_livestreams_updated_at
BEFORE UPDATE ON livestreams
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TABLE livestreams;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TYPE IF EXISTS livestream_visibility;
DROP FUNCTION IF EXISTS update_updated_at_column();
