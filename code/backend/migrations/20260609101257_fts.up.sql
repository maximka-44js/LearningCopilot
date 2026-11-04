CREATE EXTENSION IF NOT EXISTS unaccent;

ALTER TABLE chunks
	ADD COLUMN IF NOT EXISTS fts tsvector;

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

DO $$
DECLARE
	v_rows_updated INT;
BEGIN
	LOOP
		WITH batch AS (
			SELECT ctid
			FROM chunks
			WHERE fts IS NULL
			LIMIT 5000
		)
		UPDATE chunks c
		SET fts =
			setweight(to_tsvector('russian', unaccent(coalesce(c.content, ''))), 'A') ||
			setweight(to_tsvector('simple', unaccent(coalesce(c.content, ''))), 'B')
		FROM batch
		WHERE c.ctid = batch.ctid;

		GET DIAGNOSTICS v_rows_updated = ROW_COUNT;
		EXIT WHEN v_rows_updated = 0;
	END LOOP;
END;
$$;

CREATE INDEX IF NOT EXISTS idx_chunks_fts
ON chunks
USING gin (fts);

CREATE INDEX IF NOT EXISTS idx_files_search_scope
ON files (user_id, library_id, id)
WHERE deleted_at IS NULL AND search_status = 'searchable';

CREATE INDEX IF NOT EXISTS idx_files_user_file_searchable
ON files (user_id, id)
WHERE deleted_at IS NULL AND search_status = 'searchable';
