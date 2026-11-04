DROP INDEX IF EXISTS idx_files_index_provider_live;
DROP INDEX IF EXISTS idx_files_source_live;
DROP INDEX IF EXISTS idx_files_derived_from_live;
DROP INDEX IF EXISTS idx_folders_parse_source_live;

ALTER TABLE files
DROP COLUMN IF EXISTS source,
DROP COLUMN IF EXISTS derived_from_file_id,
DROP COLUMN IF EXISTS search_status,
DROP COLUMN IF EXISTS index_provider,
DROP COLUMN IF EXISTS index_profile,
DROP COLUMN IF EXISTS index_tool_hints,
DROP COLUMN IF EXISTS parse_status,
DROP COLUMN IF EXISTS parse_provider,
DROP COLUMN IF EXISTS parse_markdown_ref,
DROP COLUMN IF EXISTS parse_error_message,
DROP COLUMN IF EXISTS parse_updated_at;

ALTER TABLE folders
DROP COLUMN IF EXISTS kind,
DROP COLUMN IF EXISTS parse_source_file_id,
DROP COLUMN IF EXISTS parse_provider;

DROP TABLE IF EXISTS vectorization_batch_checkpoints;
DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS vectorizations;
