-- +goose Up
-- Add status field to vods table
-- 'ready' default preserves backward compat for existing livestream VODs
-- Possible values: 'uploading', 'processing', 'ready', 'failed'
ALTER TABLE vods ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'ready';

-- Original file URL for uploaded videos (raw file in MinIO before transcoding)
ALTER TABLE vods ADD COLUMN IF NOT EXISTS original_file_url VARCHAR(2048);

-- Make livestream_id nullable so uploaded VODs don't need a livestream
ALTER TABLE vods ALTER COLUMN livestream_id DROP NOT NULL;

-- Transcode job queue (PostgreSQL as simple job queue)
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

-- +goose Down
DROP INDEX IF EXISTS idx_transcode_jobs_status;
DROP TABLE IF EXISTS transcode_jobs;
ALTER TABLE vods ALTER COLUMN livestream_id SET NOT NULL;
ALTER TABLE vods DROP COLUMN IF EXISTS original_file_url;
ALTER TABLE vods DROP COLUMN IF EXISTS status;
