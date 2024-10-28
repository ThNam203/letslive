-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "users" (
  "id" uuid DEFAULT uuid_generate_v4(), 
  "username" varchar(20) NOT NULL, 
  "email" text NOT NULL, 
  "password_hash" text, 
  "is_verified" boolean NOT NULL DEFAULT false, 
  "is_online" boolean NOT NULL DEFAULT false, 
  "created_at" timestamptz DEFAULT current_timestamp, 
  "stream_api_key" uuid NOT NULL DEFAULT uuid_generate_v4(), 
  PRIMARY KEY ("id"), 
  CONSTRAINT "uni_users_username" UNIQUE ("username"), 
  CONSTRAINT "uni_users_email" UNIQUE ("email")
);

CREATE TABLE "refresh_tokens" (
  "id" uuid, 
  "value" varchar(255) NOT NULL, 
  "expires_at" timestamptz NOT NULL, 
  "created_at" timestamptz NOT NULL DEFAULT current_timestamp, 
  "revoked_at" timestamptz, 
  "user_id" uuid NOT NULL, 
  PRIMARY KEY ("id"), 
  CONSTRAINT "fk_users_refresh_tokens" FOREIGN KEY ("user_id") REFERENCES "users"("id"), 
  CONSTRAINT "uni_refresh_tokens_value" UNIQUE ("value")
);

CREATE INDEX IF NOT EXISTS "idx_refresh_tokens_user_id" ON "refresh_tokens" ("user_id");

CREATE TABLE "verify_tokens" (
  "id" uuid DEFAULT uuid_generate_v4(), 
  "token" varchar(255) NOT NULL, 
  "expires_at" timestamptz NOT NULL, 
  "created_at" timestamptz DEFAULT current_timestamp, 
  "user_id" uuid NOT NULL, 
  PRIMARY KEY ("id"), 
  CONSTRAINT "fk_users_verify_tokens" FOREIGN KEY ("user_id") REFERENCES "users"("id"), 
  CONSTRAINT "uni_verify_tokens_token" UNIQUE ("token")
);

CREATE INDEX IF NOT EXISTS "idx_verify_tokens_user_id" ON "verify_tokens" ("user_id");

-- +goose Down
DROP INDEX IF EXISTS "idx_verify_tokens_user_id";
DROP TABLE IF EXISTS "verify_tokens";

DROP INDEX IF EXISTS "idx_refresh_tokens_user_id";
DROP TABLE IF EXISTS "refresh_tokens";

DROP TABLE IF EXISTS "users";
DROP EXTENSION IF EXISTS "uuid-ossp";
