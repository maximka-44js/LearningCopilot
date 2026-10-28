DROP INDEX IF EXISTS idx_files_object_key_live;
DROP INDEX IF EXISTS idx_files_checksum_live;

ALTER TABLE files
DROP COLUMN IF EXISTS object_key;
