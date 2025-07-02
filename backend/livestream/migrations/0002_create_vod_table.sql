-- +goose Up
CREATE TYPE vod_visibility AS ENUM ('public', 'private');

CREATE TABLE vods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    livestream_id UUID,
    user_id UUID NOT NULL,
    title VARCHAR(255),
    description TEXT,
    thumbnail_url VARCHAR(1024),
    visibility vod_visibility NOT NULL DEFAULT 'private',
    view_count BIGINT NOT NULL DEFAULT 0,
    duration BIGINT,
    playback_url VARCHAR(2048),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_vod_livestream
        FOREIGN KEY (livestream_id)
        REFERENCES livestreams(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_vod_livestream_id ON vods(livestream_id);
CREATE INDEX idx_vod_user_id ON vods(user_id);

CREATE TRIGGER trigger_vods_updated_at
BEFORE UPDATE ON vods
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_vods_updated_at ON vods;
DROP INDEX IF EXISTS idx_vod_user_id;
DROP INDEX IF EXISTS idx_vod_livestream_id;
DROP TABLE IF EXISTS vods;
DROP TYPE IF EXISTS vod_visibility;
