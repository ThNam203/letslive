-- +goose Up
ALTER TABLE users RENAME TO auths;

-- +goose Down 
ALTER TABLE auths RENAME TO users;
