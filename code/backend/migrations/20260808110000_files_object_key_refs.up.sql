ALTER TABLE files
ADD COLUMN IF NOT EXISTS object_key TEXT;

UPDATE files
SET object_key = 'files/' || user_id::text || '/' || library_id::text || '/' || COALESCE(folder_id::text, '') || '/' || name
WHERE object_key IS NULL;

CREATE INDEX IF NOT EXISTS idx_files_checksum_live
ON files (checksum)
WHERE deleted_at IS NULL AND checksum IS NOT NULL AND checksum <> '';

CREATE INDEX IF NOT EXISTS idx_files_object_key_live
ON files (object_key)
WHERE deleted_at IS NULL AND object_key IS NOT NULL AND object_key <> '';
