CREATE OR REPLACE FUNCTION chunks_fts_update()
RETURNS trigger
AS $$
BEGIN
	NEW.fts :=
		setweight(to_tsvector('russian', unaccent(coalesce(NEW.content, ''))), 'A') ||
		setweight(to_tsvector('simple', unaccent(coalesce(NEW.content, ''))), 'B');
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_chunks_fts_update ON chunks;

CREATE TRIGGER trg_chunks_fts_update
BEFORE INSERT OR UPDATE OF content
ON chunks
FOR EACH ROW
EXECUTE FUNCTION chunks_fts_update();

UPDATE chunks c
SET fts =
	setweight(to_tsvector('russian', unaccent(coalesce(c.content, ''))), 'A') ||
	setweight(to_tsvector('simple', unaccent(coalesce(c.content, ''))), 'B');
