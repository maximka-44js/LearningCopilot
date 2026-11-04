DROP INDEX IF EXISTS idx_files_index_provider_live;

ALTER TABLE files
DROP COLUMN IF EXISTS index_tool_hints,
DROP COLUMN IF EXISTS index_profile,
DROP COLUMN IF EXISTS index_provider;
