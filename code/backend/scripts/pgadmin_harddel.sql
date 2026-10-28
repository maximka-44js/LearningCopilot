BEGIN;

DELETE FROM files
WHERE deleted_at IS NOT NULL;

DELETE FROM folders
WHERE deleted_at IS NOT NULL;

DELETE FROM libraries
WHERE deleted_at IS NOT NULL;

COMMIT;