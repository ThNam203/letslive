-- +goose Up
alter table "users" add column "display_name" varchar(50);
alter table "users" add column "phone_number" varchar(15);
alter table "users" add column "bio" text;
alter table "users" add column "is_active" boolean default true;
alter table "users" add column "profile_picture" text;
alter table "users" add column "background_picture" text;

-- +goose Down
alter table "users" drop column "display_name";
aLTER TABLE "users" DROP COLUMN "phone_number";
ALTER TABLE "users" DROP COLUMN "bio";
ALTER TABLE "users" DROP COLUMN "profile_picture";
ALTER TABLE "users" DROP COLUMN "is_active";
ALTER TABLE "users" DROP COLUMN "background_picture";
ALTER TABLE "users" DROP COLUMN "is_active";
