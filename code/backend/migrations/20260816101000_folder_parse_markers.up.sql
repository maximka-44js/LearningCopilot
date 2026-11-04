ALTER TABLE folders
ADD COLUMN IF NOT EXISTS kind VARCHAR(32) NOT NULL DEFAULT 'regular'
    CHECK (kind IN ('regular', 'parse_result_root')),
ADD COLUMN IF NOT EXISTS parse_source_file_id UUID NULL,
ADD COLUMN IF NOT EXISTS parse_provider VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_folders_parse_source_live
ON folders (parse_source_file_id)
WHERE deleted_at IS NULL AND parse_source_file_id IS NOT NULL;
