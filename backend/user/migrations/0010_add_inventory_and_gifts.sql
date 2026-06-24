-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_inventory (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shop_item_id  UUID NOT NULL,
    quantity      INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, shop_item_id)
);

CREATE TABLE gifts (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_user_id    UUID NOT NULL REFERENCES users(id),
    recipient_user_id UUID NOT NULL REFERENCES users(id),
    shop_item_id      UUID NOT NULL,
    quantity          INTEGER NOT NULL DEFAULT 1 CHECK (quantity >= 1),
    message           TEXT,
    sent_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX gifts_recipient_idx ON gifts (recipient_user_id, sent_at DESC);
CREATE INDEX gifts_sender_idx    ON gifts (sender_user_id,    sent_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS gifts_sender_idx;
DROP INDEX IF EXISTS gifts_recipient_idx;
DROP TABLE IF EXISTS gifts;
DROP TABLE IF EXISTS user_inventory;
-- +goose StatementEnd
