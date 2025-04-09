-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "auths" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "email" text NOT NULL, 
  "password_hash" text, 
  "created_at" timestamptz DEFAULT current_timestamp, 
  CONSTRAINT "uni_auths_email" UNIQUE ("email"),
  CONSTRAINT "uni_auths_user_id" UNIQUE ("user_id")
);

CREATE INDEX IF NOT EXISTS "idx_auths_email" ON "auths"("email");

CREATE TABLE "refresh_tokens" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY, 
  "token" varchar(1024) NOT NULL, 
  "expires_at" timestamptz NOT NULL, 
  "created_at" timestamptz NOT NULL DEFAULT current_timestamp, 
  "revoked_at" timestamptz, 
  "user_id" uuid NOT NULL, 
  CONSTRAINT "fk_auths_refresh_tokens" FOREIGN KEY ("user_id") REFERENCES "auths"("user_id")
);

CREATE INDEX IF NOT EXISTS "idx_refresh_tokens_user_id" ON "refresh_tokens" ("user_id");

CREATE TABLE "sign_up_otps" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY, 
  "code" char(6) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz DEFAULT current_timestamp,
  "email" text NOT NULL,
  "used_at" timestamptz
);

-- +goose Down
DROP TABLE IF EXISTS "sign_up_otps";

DROP INDEX IF EXISTS "idx_refresh_tokens_user_id";
DROP TABLE IF EXISTS "refresh_tokens";

DROP INDEX IF EXISTS "idx_auths_email";
DROP TABLE IF EXISTS "auths";

DROP EXTENSION IF EXISTS "uuid-ossp";
