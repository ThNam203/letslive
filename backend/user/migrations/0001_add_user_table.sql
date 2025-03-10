-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE user_auth_provider_enum AS ENUM ('google', 'local');
CREATE TYPE user_status_enum AS ENUM ('normal', 'disabled');

CREATE TABLE "users" (
  "id" UUID DEFAULT uuid_generate_v4(), 
  "username" VARCHAR(20) NOT NULL, 
  "email" TEXT NOT NULL, 
  "is_verified" BOOLEAN NOT NULL DEFAULT false, 
  "created_at" TIMESTAMPTZ  DEFAULT current_timestamp, 
  "stream_api_key" UUID NOT NULL DEFAULT uuid_generate_v4(), 
  "display_name" VARCHAR(50),
  "phone_number" VARCHAR(15),
  "bio" TEXT,
  "status" user_status_enum NOT NULL DEFAULT 'normal',
  "profile_picture" TEXT,
  "background_picture" TEXT,
  "auth_provider" user_auth_provider_enum NOT NULL DEFAULT 'local',
  PRIMARY KEY ("id"), 
  CONSTRAINT "uni_users_username" UNIQUE ("username"), 
  CONSTRAINT "uni_users_email" UNIQUE ("email")
);

-- +goose Down
DROP TABLE IF EXISTS "users";
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TYPE IF EXISTS user_auth_provider_enum;
DROP TYPE IF EXISTS user_status_enum;
