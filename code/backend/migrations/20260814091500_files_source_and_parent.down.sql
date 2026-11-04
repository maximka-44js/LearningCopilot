DROP INDEX IF EXISTS idx_files_derived_from_live;
DROP INDEX IF EXISTS idx_files_source_live;

ALTER TABLE files
DROP CONSTRAINT IF EXISTS fk_files_derived_from_file;

ALTER TABLE files
DROP COLUMN IF EXISTS derived_from_file_id,
DROP COLUMN IF EXISTS source,
DROP COLUMN IF EXISTS parse_provider;
