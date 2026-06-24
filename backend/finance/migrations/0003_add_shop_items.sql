-- +goose Up
-- +goose StatementBegin
CREATE TABLE shop_items (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(100) NOT NULL,
    description   TEXT,
    image_url     TEXT NOT NULL,
    animation_url TEXT NOT NULL,
    price         INTEGER NOT NULL CHECK (price > 0),
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO shop_items (name, description, image_url, animation_url, price) VALUES
    ('Rose',   'A beautiful red rose', 'https://placeholder.co/rose.png',   'https://placeholder.co/rose.json',   100),
    ('Crown',  'A golden crown',       'https://placeholder.co/crown.png',  'https://placeholder.co/crown.json',  500),
    ('Rocket', 'A blazing rocket',     'https://placeholder.co/rocket.png', 'https://placeholder.co/rocket.json', 1000);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shop_items;
-- +goose StatementEnd
