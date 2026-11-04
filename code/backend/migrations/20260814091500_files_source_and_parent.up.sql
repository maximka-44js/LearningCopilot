ALTER TABLE files
ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'user_upload'
    CHECK (source IN ('user_upload', 'parse_result')),
ADD COLUMN IF NOT EXISTS derived_from_file_id UUID NULL,
ADD COLUMN IF NOT EXISTS parse_provider VARCHAR(64) NOT NULL DEFAULT '';

ALTER TABLE files
ADD CONSTRAINT fk_files_derived_from_file
FOREIGN KEY (derived_from_file_id)
REFERENCES files(id)
ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_files_source_live
ON files (source)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_files_derived_from_live
ON files (derived_from_file_id)
WHERE deleted_at IS NULL AND derived_from_file_id IS NOT NULL;
