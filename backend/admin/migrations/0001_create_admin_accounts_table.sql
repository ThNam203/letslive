-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "admin_accounts" (
  "id" uuid DEFAULT uuid_generate_v4() PRIMARY KEY,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT current_timestamp,
  CONSTRAINT "uni_admin_accounts_email" UNIQUE ("email")
);

CREATE INDEX IF NOT EXISTS "idx_admin_accounts_email" ON "admin_accounts"("email");

-- +goose Down
DROP INDEX IF EXISTS "idx_admin_accounts_email";
DROP TABLE IF EXISTS "admin_accounts";
DROP EXTENSION IF EXISTS "uuid-ossp";
