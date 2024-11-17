-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "users" (
  "id" uuid DEFAULT uuid_generate_v4(), 
  "username" varchar(20) NOT NULL, 
  "email" text NOT NULL, 
  "is_verified" boolean NOT NULL DEFAULT false, 
  "is_online" boolean NOT NULL DEFAULT false, 
  "created_at" timestamptz DEFAULT current_timestamp, 
  "stream_api_key" uuid NOT NULL DEFAULT uuid_generate_v4(), 
  PRIMARY KEY ("id"), 
  CONSTRAINT "uni_users_username" UNIQUE ("username"), 
  CONSTRAINT "uni_users_email" UNIQUE ("email")
);

-- +goose Down
DROP TABLE IF EXISTS "users";
DROP EXTENSION IF EXISTS "uuid-ossp"
