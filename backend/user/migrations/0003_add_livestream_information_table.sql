-- +goose Up
CREATE TABLE "livestream_information" (
  "user_id" UUID NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "title" VARCHAR(50),
  "description" VARCHAR(500),
  "thumbnail_url" text,
  PRIMARY KEY ("user_id")
);

-- +goose Down
DROP TABLE "livestream_information";
