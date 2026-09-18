-- +goose Up
-- +goose StatementBegin

ALTER TABLE vod_comments ADD COLUMN IF NOT EXISTS is_edited BOOLEAN NOT NULL DEFAULT false;

-- Edit history for VOD comments.
--
-- Each row is the "before" image of one edit: the content the comment held
-- until that edit replaced it. The newest version is never stored here, it
-- lives in vod_comments.content. So the full history of a comment is
--   SELECT previous_content FROM vod_comment_edits ... ORDER BY version ASC
-- followed by the comment's current content as the latest version.
--
-- Nothing reads this table yet: it is written so an edit history feature has
-- the data when it is built.
CREATE TABLE IF NOT EXISTS vod_comment_edits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    comment_id UUID NOT NULL,
    version INT NOT NULL,
    previous_content VARCHAR(2000) NOT NULL,
    edited_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_vod_comment_edit_comment
        FOREIGN KEY (comment_id)
        REFERENCES vod_comments(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_vod_comment_edit_version
        UNIQUE (comment_id, version)
);

CREATE INDEX IF NOT EXISTS idx_vod_comment_edit_comment_id ON vod_comment_edits(comment_id, version);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_vod_comment_edit_comment_id;
DROP TABLE IF EXISTS vod_comment_edits;
ALTER TABLE vod_comments DROP COLUMN IF EXISTS is_edited;

-- +goose StatementEnd
