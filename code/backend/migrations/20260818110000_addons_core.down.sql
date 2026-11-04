DROP TABLE IF EXISTS addon_audit_events;
DROP TABLE IF EXISTS addon_runtime_sessions;
DROP TABLE IF EXISTS addon_api_key_bindings;
DROP TABLE IF EXISTS addon_installations;
ALTER TABLE addons_catalog DROP CONSTRAINT IF EXISTS fk_addons_catalog_current_version;
DROP TABLE IF EXISTS addon_versions;
DROP TABLE IF EXISTS addons_catalog;
