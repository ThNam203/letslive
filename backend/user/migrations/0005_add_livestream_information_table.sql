-- +goose Up
CREATE TABLE "livestream_information" (
  "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "title" text,
  "description" text,
  "thumbnail_url" text,
  PRIMARY KEY ("user_id")
);

-- +goose Down
DROP TABLE "livestream_information";
