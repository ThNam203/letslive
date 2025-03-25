-- +goose Up
CREATE TABLE "followers" (
  "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "follower_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "followed_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
  PRIMARY KEY ("user_id", "follower_id")
);

-- for fetching user's followers
CREATE INDEX idx_followers_user ON followers(user_id);

-- +goose Down
DROP TABLE followers;
DROP INDEX idx_followers_user;
