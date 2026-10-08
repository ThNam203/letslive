-- +goose Up
-- +goose StatementBegin

ALTER TABLE vods ADD COLUMN IF NOT EXISTS like_count BIGINT NOT NULL DEFAULT 0;
-- kept for ranking and the owner's own stats; the public api never returns it
ALTER TABLE vods ADD COLUMN IF NOT EXISTS dislike_count BIGINT NOT NULL DEFAULT 0;

-- One row per user per VOD: a user either likes or dislikes a VOD, never both.
CREATE TABLE IF NOT EXISTS vod_reactions (
    vod_id UUID NOT NULL,
    user_id UUID NOT NULL,
    reaction VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (vod_id, user_id),

    CONSTRAINT fk_vod_reaction_vod
        FOREIGN KEY (vod_id)
        REFERENCES vods(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_vod_reaction_value
        CHECK (reaction IN ('like', 'dislike'))
);

CREATE INDEX IF NOT EXISTS idx_vod_reaction_user_id ON vod_reactions(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_vod_reaction_user_id;
DROP TABLE IF EXISTS vod_reactions;
ALTER TABLE vods DROP COLUMN IF EXISTS dislike_count;
ALTER TABLE vods DROP COLUMN IF EXISTS like_count;

-- +goose StatementEnd
