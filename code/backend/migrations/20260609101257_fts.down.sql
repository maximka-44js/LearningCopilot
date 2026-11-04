DROP INDEX IF EXISTS idx_files_user_file_searchable;
DROP INDEX IF EXISTS idx_files_search_scope;
DROP INDEX IF EXISTS idx_chunks_fts;

DROP TRIGGER IF EXISTS trg_chunks_fts_update ON chunks;
DROP FUNCTION IF EXISTS chunks_fts_update();

ALTER TABLE chunks
	DROP COLUMN IF EXISTS fts;
