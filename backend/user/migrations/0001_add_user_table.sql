-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE user_auth_provider_enum AS ENUM ('google', 'local');
CREATE TYPE user_status_enum AS ENUM ('normal', 'disabled');

CREATE TABLE "users" (
  "id" UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  "username" VARCHAR(50) NOT NULL UNIQUE,
  "email" TEXT NOT NULL UNIQUE,
  "created_at" TIMESTAMPTZ NOT NULL DEFAULT current_timestamp, 
  "stream_api_key" UUID NOT NULL DEFAULT uuid_generate_v4() UNIQUE, 
  "display_name" VARCHAR(50) UNIQUE,
  "phone_number" VARCHAR(15),
  "bio" VARCHAR(300),
  "status" user_status_enum NOT NULL DEFAULT 'normal',
  "profile_picture" TEXT,
  "background_picture" TEXT,
  "auth_provider" user_auth_provider_enum NOT NULL DEFAULT 'local'
);

CREATE INDEX "idx_users_stream_api_key" ON users(stream_api_key);

-- +goose Down
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP TYPE IF EXISTS user_auth_provider_enum;
DROP TYPE IF EXISTS user_status_enum;
DROP INDEX IF EXISTS idx_users_stream_api_key;
