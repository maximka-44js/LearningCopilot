CREATE TABLE IF NOT EXISTS addons_catalog (
	id UUID PRIMARY KEY,
	owner_id UUID NOT NULL,
	manifest_url TEXT NOT NULL CHECK (length(trim(manifest_url)) > 0),
	status VARCHAR(24) NOT NULL CHECK (status IN ('draft', 'verified', 'active', 'disabled', 'archived')),
	display_name VARCHAR(160) NOT NULL DEFAULT '',
	vendor VARCHAR(160) NOT NULL DEFAULT '',
	current_version_id UUID,
	provenance JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	archived_at TIMESTAMPTZ,

	CONSTRAINT uq_addons_catalog_owner_manifest UNIQUE (owner_id, manifest_url)
);

CREATE TABLE IF NOT EXISTS addon_versions (
	id UUID PRIMARY KEY,
	addon_id UUID NOT NULL REFERENCES addons_catalog(id) ON DELETE CASCADE,
	manifest_version VARCHAR(64) NOT NULL,
	manifest_snapshot JSONB NOT NULL,
	manifest_hash CHAR(64) NOT NULL,
	signature TEXT,
	fetched_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT uq_addon_versions_addon_hash UNIQUE (addon_id, manifest_hash)
);

ALTER TABLE addons_catalog
	ADD CONSTRAINT fk_addons_catalog_current_version
	FOREIGN KEY (current_version_id)
	REFERENCES addon_versions(id)
	ON DELETE SET NULL;

CREATE TABLE IF NOT EXISTS addon_installations (
	id UUID PRIMARY KEY,
	addon_id UUID NOT NULL REFERENCES addons_catalog(id) ON DELETE CASCADE,
	owner_id UUID NOT NULL,
	status VARCHAR(24) NOT NULL CHECK (status IN ('created', 'configured', 'connected', 'healthy', 'unhealthy', 'running', 'stopped', 'blocked')),
	config JSONB NOT NULL DEFAULT '{}'::jsonb,
	config_version INT NOT NULL DEFAULT 1 CHECK (config_version > 0),
	pinned_version_id UUID REFERENCES addon_versions(id) ON DELETE SET NULL,
	health_status VARCHAR(24) NOT NULL DEFAULT 'unknown',
	last_launch_at TIMESTAMPTZ,
	blocked_reason TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT uq_addon_installations_owner_addon UNIQUE (owner_id, addon_id)
);

CREATE TABLE IF NOT EXISTS addon_api_key_bindings (
	id UUID PRIMARY KEY,
	installation_id UUID NOT NULL UNIQUE REFERENCES addon_installations(id) ON DELETE CASCADE,
	mode VARCHAR(16) NOT NULL CHECK (mode IN ('auto', 'manual')),
	api_key_id UUID REFERENCES api_keys(id) ON DELETE SET NULL,
	scope_profile JSONB NOT NULL DEFAULT '[]'::jsonb,
	scope_profile_version INT NOT NULL DEFAULT 1 CHECK (scope_profile_version > 0),
	last_rotated_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS addon_runtime_sessions (
	id UUID PRIMARY KEY,
	installation_id UUID NOT NULL REFERENCES addon_installations(id) ON DELETE CASCADE,
	status VARCHAR(16) NOT NULL CHECK (status IN ('running', 'stopped')),
	session_nonce VARCHAR(128) NOT NULL CHECK (length(trim(session_nonce)) > 0),
	iframe_url TEXT NOT NULL CHECK (length(trim(iframe_url)) > 0),
	allowed_origin TEXT NOT NULL CHECK (length(trim(allowed_origin)) > 0),
	bridge_protocol_version VARCHAR(32) NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	terminated_reason VARCHAR(240),
	started_at TIMESTAMPTZ NOT NULL,
	stopped_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS addon_audit_events (
	id UUID PRIMARY KEY,
	owner_id UUID NOT NULL,
	addon_id UUID REFERENCES addons_catalog(id) ON DELETE SET NULL,
	installation_id UUID REFERENCES addon_installations(id) ON DELETE SET NULL,
	event_type VARCHAR(64) NOT NULL,
	severity VARCHAR(16) NOT NULL CHECK (severity IN ('info', 'warning', 'error')),
	payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_addons_catalog_owner_status
ON addons_catalog (owner_id, status);

CREATE INDEX IF NOT EXISTS idx_addons_versions_addon_created
ON addon_versions (addon_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_addon_installations_owner_status
ON addon_installations (owner_id, status);

CREATE INDEX IF NOT EXISTS idx_addon_runtime_sessions_installation_status_exp
ON addon_runtime_sessions (installation_id, status, expires_at DESC);

CREATE INDEX IF NOT EXISTS idx_addon_audit_events_owner_created
ON addon_audit_events (owner_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_addon_audit_events_type_created
ON addon_audit_events (event_type, created_at DESC);
