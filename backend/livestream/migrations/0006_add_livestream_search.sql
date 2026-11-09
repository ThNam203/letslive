-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

-- unaccent() is only STABLE (its dictionary can change), so it cannot appear
-- in an index expression. Pinning the dictionary makes this wrapper safe to
-- declare IMMUTABLE. Search queries must use exactly the same expressions as
-- the indexes below, otherwise the indexes are not used.
CREATE OR REPLACE FUNCTION immutable_unaccent(text)
RETURNS text
LANGUAGE sql
IMMUTABLE PARALLEL SAFE STRICT
AS $$
    SELECT public.unaccent('public.unaccent'::regdictionary, $1);
$$;

CREATE INDEX IF NOT EXISTS idx_livestream_title_search_trgm
    ON livestreams USING GIN (immutable_unaccent(lower(coalesce(title, ''))) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_livestream_description_search_trgm
    ON livestreams USING GIN (immutable_unaccent(lower(coalesce(description, ''))) gin_trgm_ops);
-- every OR branch of the search must be indexable, including the streamer match
CREATE INDEX IF NOT EXISTS idx_livestream_user_id ON livestreams(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_livestream_user_id;
DROP INDEX IF EXISTS idx_livestream_description_search_trgm;
DROP INDEX IF EXISTS idx_livestream_title_search_trgm;
DROP FUNCTION IF EXISTS immutable_unaccent(text);
DROP EXTENSION IF EXISTS unaccent;
DROP EXTENSION IF EXISTS pg_trgm;

-- +goose StatementEnd
