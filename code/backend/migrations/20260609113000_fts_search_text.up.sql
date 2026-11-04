CREATE OR REPLACE FUNCTION chunks_fts_update()
RETURNS trigger
AS $$
DECLARE
	source_text TEXT;
BEGIN
	source_text := coalesce(nullif(trim(NEW.metadata->>'search_text'), ''), NEW.content, '');
	NEW.fts :=
		setweight(to_tsvector('russian', unaccent(source_text)), 'A') ||
		setweight(to_tsvector('simple', unaccent(source_text)), 'B');
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_chunks_fts_update ON chunks;

CREATE TRIGGER trg_chunks_fts_update
BEFORE INSERT OR UPDATE OF content, metadata
ON chunks
FOR EACH ROW
EXECUTE FUNCTION chunks_fts_update();

UPDATE chunks c
SET fts =
	setweight(to_tsvector('russian', unaccent(coalesce(nullif(trim(c.metadata->>'search_text'), ''), c.content, ''))), 'A') ||
	setweight(to_tsvector('simple', unaccent(coalesce(nullif(trim(c.metadata->>'search_text'), ''), c.content, ''))), 'B')
WHERE c.fts IS NULL OR (c.metadata ? 'search_text');
