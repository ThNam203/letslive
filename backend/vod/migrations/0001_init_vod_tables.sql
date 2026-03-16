-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ENUM Type
DO $$ BEGIN
    CREATE TYPE vod_visibility AS ENUM ('public', 'private');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- VODs Table
CREATE TABLE IF NOT EXISTS vods (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    livestream_id UUID,
    user_id UUID NOT NULL,
    title VARCHAR(255),
    description VARCHAR(1000),
    thumbnail_url VARCHAR(2048),
    visibility vod_visibility NOT NULL DEFAULT 'private',
    view_count BIGINT NOT NULL DEFAULT 0,
    duration BIGINT NOT NULL DEFAULT 0,
    playback_url VARCHAR(2048),
    status VARCHAR(20) NOT NULL DEFAULT 'ready',
    original_file_url VARCHAR(2048),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_vod_livestream_id ON vods(livestream_id);
CREATE INDEX IF NOT EXISTS idx_vod_user_id ON vods(user_id);

-- Transcode Jobs Table
CREATE TABLE IF NOT EXISTS transcode_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vod_id UUID NOT NULL REFERENCES vods(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_transcode_jobs_status ON transcode_jobs(status, created_at);

-- VOD Comments Table
CREATE TABLE IF NOT EXISTS vod_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vod_id UUID NOT NULL,
    user_id UUID NOT NULL,
    parent_id UUID,
    content VARCHAR(2000) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    like_count BIGINT NOT NULL DEFAULT 0,
    reply_count BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_vod_comment_vod
        FOREIGN KEY (vod_id)
        REFERENCES vods(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_vod_comment_parent
        FOREIGN KEY (parent_id)
        REFERENCES vod_comments(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vod_comment_vod_id ON vod_comments(vod_id);
CREATE INDEX IF NOT EXISTS idx_vod_comment_parent_id ON vod_comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_vod_comment_user_id ON vod_comments(user_id);

-- VOD Comment Likes Table
CREATE TABLE IF NOT EXISTS vod_comment_likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_vod_comment_like_comment
        FOREIGN KEY (comment_id)
        REFERENCES vod_comments(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_vod_comment_like_user
        UNIQUE (comment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_vod_comment_like_comment_id ON vod_comment_likes(comment_id);
CREATE INDEX IF NOT EXISTS idx_vod_comment_like_user_id ON vod_comment_likes(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_vod_comment_like_user_id;
DROP INDEX IF EXISTS idx_vod_comment_like_comment_id;
DROP INDEX IF EXISTS idx_vod_comment_user_id;
DROP INDEX IF EXISTS idx_vod_comment_parent_id;
DROP INDEX IF EXISTS idx_vod_comment_vod_id;
DROP INDEX IF EXISTS idx_transcode_jobs_status;
DROP INDEX IF EXISTS idx_vod_user_id;
DROP INDEX IF EXISTS idx_vod_livestream_id;

DROP TABLE IF EXISTS vod_comment_likes;
DROP TABLE IF EXISTS vod_comments;
DROP TABLE IF EXISTS transcode_jobs;
DROP TABLE IF EXISTS vods;

DROP TYPE IF EXISTS vod_visibility;
DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd
