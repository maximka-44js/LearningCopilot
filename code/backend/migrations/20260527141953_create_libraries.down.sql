DROP TRIGGER IF EXISTS set_libraries_updated_at ON libraries;

DROP FUNCTION IF EXISTS trigger_set_timestamp();

DROP TABLE IF EXISTS library_accesses;

DROP TABLE IF EXISTS libraries;

DROP TABLE IF EXISTS users;