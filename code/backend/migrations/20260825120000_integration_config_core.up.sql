CREATE TABLE IF NOT EXISTS integration_config_schemas_cache (
	owner_id UUID NOT NULL,
	integration_kind VARCHAR(16) NOT NULL CHECK (integration_kind IN ('addon', 'agent', 'mcp')),
	integration_id UUID NOT NULL,
	schema_json JSONB NOT NULL,
	schema_version INT NOT NULL CHECK (schema_version > 0),
	source VARCHAR(16) NOT NULL CHECK (source IN ('fetched', 'cached', 'unavailable')),
	last_fetched_at TIMESTAMPTZ NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (owner_id, integration_kind, integration_id)
);

CREATE TABLE IF NOT EXISTS integration_configs_public (
	owner_id UUID NOT NULL,
	integration_kind VARCHAR(16) NOT NULL CHECK (integration_kind IN ('addon', 'agent', 'mcp')),
	integration_id UUID NOT NULL,
	config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	version INT NOT NULL DEFAULT 1 CHECK (version > 0),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (owner_id, integration_kind, integration_id)
);

CREATE TABLE IF NOT EXISTS integration_configs_secret (
	owner_id UUID NOT NULL,
	integration_kind VARCHAR(16) NOT NULL CHECK (integration_kind IN ('addon', 'agent', 'mcp')),
	integration_id UUID NOT NULL,
	secret_key VARCHAR(160) NOT NULL,
	cipher BYTEA NOT NULL,
	nonce BYTEA NOT NULL,
	key_version VARCHAR(32) NOT NULL DEFAULT 'v1',
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (owner_id, integration_kind, integration_id, secret_key)
);

CREATE TABLE IF NOT EXISTS integration_config_audit (
	id UUID PRIMARY KEY,
	owner_id UUID NOT NULL,
	integration_kind VARCHAR(16) NOT NULL CHECK (integration_kind IN ('addon', 'agent', 'mcp')),
	integration_id UUID NOT NULL,
	actor_id UUID NOT NULL,
	action VARCHAR(64) NOT NULL,
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_integration_config_schemas_cache_expires
	ON integration_config_schemas_cache (expires_at);

CREATE INDEX IF NOT EXISTS idx_integration_configs_public_updated
	ON integration_configs_public (updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_integration_configs_secret_updated
	ON integration_configs_secret (updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_integration_config_audit_owner_created
	ON integration_config_audit (owner_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_integration_config_audit_action_created
	ON integration_config_audit (action, created_at DESC);
