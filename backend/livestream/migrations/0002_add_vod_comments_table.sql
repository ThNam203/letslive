-- +goose Up
-- +goose StatementBegin

CREATE TABLE vod_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vod_id UUID NOT NULL,
    user_id UUID NOT NULL,
    parent_id UUID,
    content TEXT NOT NULL,
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

CREATE TABLE vod_comment_likes (
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

CREATE INDEX idx_vod_comment_vod_id ON vod_comments(vod_id);
CREATE INDEX idx_vod_comment_parent_id ON vod_comments(parent_id);
CREATE INDEX idx_vod_comment_user_id ON vod_comments(user_id);
CREATE INDEX idx_vod_comment_like_comment_id ON vod_comment_likes(comment_id);
CREATE INDEX idx_vod_comment_like_user_id ON vod_comment_likes(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_vod_comment_like_user_id;
DROP INDEX IF EXISTS idx_vod_comment_like_comment_id;
DROP INDEX IF EXISTS idx_vod_comment_user_id;
DROP INDEX IF EXISTS idx_vod_comment_parent_id;
DROP INDEX IF EXISTS idx_vod_comment_vod_id;

DROP TABLE IF EXISTS vod_comment_likes;
DROP TABLE IF EXISTS vod_comments;

-- +goose StatementEnd
