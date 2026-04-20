-- +goose Up
-- +goose StatementBegin

-- Remove legacy VOD objects from livestream database after VOD service split.
-- Keep livestreams.vod_id column because livestream service still uses it.
ALTER TABLE IF EXISTS livestreams
DROP CONSTRAINT IF EXISTS fk_livestream_vod;

DROP INDEX IF EXISTS idx_transcode_jobs_status;
DROP INDEX IF EXISTS idx_vod_comment_like_user_id;
DROP INDEX IF EXISTS idx_vod_comment_like_comment_id;
DROP INDEX IF EXISTS idx_vod_comment_user_id;
DROP INDEX IF EXISTS idx_vod_comment_parent_id;
DROP INDEX IF EXISTS idx_vod_comment_vod_id;
DROP INDEX IF EXISTS idx_vod_user_id;
DROP INDEX IF EXISTS idx_vod_livestream_id;

DROP TABLE IF EXISTS transcode_jobs;
DROP TABLE IF EXISTS vod_comment_likes;
DROP TABLE IF EXISTS vod_comments;
DROP TABLE IF EXISTS vods;

DROP TYPE IF EXISTS vod_visibility;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Intentionally left blank.
-- Restoring dropped VOD tables and data is not supported by this rollback.
-- +goose StatementEnd
