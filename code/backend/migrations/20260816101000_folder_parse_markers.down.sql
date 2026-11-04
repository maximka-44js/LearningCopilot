DROP INDEX IF EXISTS idx_folders_parse_source_live;

ALTER TABLE folders
DROP COLUMN IF EXISTS parse_source_file_id,
DROP COLUMN IF EXISTS parse_provider,
DROP COLUMN IF EXISTS kind;
