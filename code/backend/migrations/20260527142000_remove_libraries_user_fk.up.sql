-- Remove foreign key constraints linking libraries and library_accesses to users
ALTER TABLE IF EXISTS library_accesses DROP CONSTRAINT IF EXISTS library_accesses_user_id_fkey;
ALTER TABLE IF EXISTS libraries DROP CONSTRAINT IF EXISTS libraries_user_id_fkey;
