-- +goose Up
ALTER TABLE users add is_chirpy_red BOOLEAN DEFAULT false;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS is_chirpy_red;