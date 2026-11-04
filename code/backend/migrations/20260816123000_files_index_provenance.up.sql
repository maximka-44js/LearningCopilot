ALTER TABLE files
ADD COLUMN IF NOT EXISTS index_provider VARCHAR(64) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS index_profile VARCHAR(64) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS index_tool_hints TEXT[] NOT NULL DEFAULT ARRAY[]::text[];

CREATE INDEX IF NOT EXISTS idx_files_index_provider_live
ON files (index_provider)
WHERE deleted_at IS NULL AND index_provider <> '';
