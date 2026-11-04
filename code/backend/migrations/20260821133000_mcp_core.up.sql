CREATE TABLE IF NOT EXISTS mcp_connections (
	id UUID PRIMARY KEY,
	owner_id UUID NOT NULL,
	name VARCHAR(120) NOT NULL CHECK (length(trim(name)) > 0),
	mode VARCHAR(16) NOT NULL CHECK (mode IN ('url', 'command')),
	enabled BOOLEAN NOT NULL DEFAULT TRUE,
	pass_when_unknown BOOLEAN NOT NULL DEFAULT TRUE,
	version INT NOT NULL DEFAULT 1 CHECK (version > 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mcp_url_configs (
	connection_id UUID PRIMARY KEY REFERENCES mcp_connections(id) ON DELETE CASCADE,
	endpoint_url TEXT NOT NULL CHECK (length(trim(endpoint_url)) > 0),
	auth_type VARCHAR(16) NOT NULL DEFAULT 'none' CHECK (auth_type IN ('none', 'bearer', 'hmac', 'basic')),
	auth_secret_ref TEXT,
	tls_verify BOOLEAN NOT NULL DEFAULT TRUE,
	headers_json JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS mcp_command_configs (
	connection_id UUID PRIMARY KEY REFERENCES mcp_connections(id) ON DELETE CASCADE,
	command TEXT NOT NULL CHECK (length(trim(command)) > 0),
	args_json JSONB NOT NULL DEFAULT '[]'::jsonb,
	working_dir TEXT,
	env_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	env_policy VARCHAR(16) NOT NULL DEFAULT 'allowlist' CHECK (env_policy IN ('allowlist', 'inherit_none', 'inherit_all')),
	resource_profile_id VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS mcp_gateway_bindings (
	connection_id UUID PRIMARY KEY REFERENCES mcp_connections(id) ON DELETE CASCADE,
	gateway_id VARCHAR(120) NOT NULL,
	telemetry_enabled BOOLEAN NOT NULL DEFAULT FALSE,
	heartbeat_interval_sec INT NOT NULL DEFAULT 30 CHECK (heartbeat_interval_sec > 0),
	heartbeat_ttl_sec INT NOT NULL DEFAULT 90 CHECK (heartbeat_ttl_sec > 0),
	last_seen_at TIMESTAMPTZ,
	active_secret_ciphertext BYTEA NOT NULL DEFAULT ''::bytea,
	active_secret_nonce BYTEA NOT NULL DEFAULT ''::bytea,
	next_secret_ciphertext BYTEA NOT NULL DEFAULT ''::bytea,
	next_secret_nonce BYTEA NOT NULL DEFAULT ''::bytea,
	secret_key_version VARCHAR(32) NOT NULL DEFAULT 'v1',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mcp_runtime_state (
	connection_id UUID PRIMARY KEY REFERENCES mcp_connections(id) ON DELETE CASCADE,
	lifecycle_state VARCHAR(16) NOT NULL CHECK (lifecycle_state IN ('draft', 'active', 'launching', 'running', 'stopping', 'stopped', 'degraded', 'failed', 'unknown')),
	status_reason TEXT,
	status_details_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	last_transition_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	last_heartbeat_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS mcp_dispatch_commands (
	id UUID PRIMARY KEY,
	connection_id UUID NOT NULL REFERENCES mcp_connections(id) ON DELETE CASCADE,
	gateway_id VARCHAR(120) NOT NULL,
	command_type VARCHAR(16) NOT NULL CHECK (command_type IN ('launch', 'stop', 'restart', 'probe')),
	payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	idempotency_key VARCHAR(160) NOT NULL,
	state VARCHAR(16) NOT NULL CHECK (state IN ('pending', 'dispatched', 'acked', 'completed', 'failed', 'canceled', 'timed_out')),
	attempts INT NOT NULL DEFAULT 0 CHECK (attempts >= 0),
	next_retry_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT uq_mcp_dispatch_commands_connection_idempotency
		UNIQUE (connection_id, idempotency_key)
);

CREATE TABLE IF NOT EXISTS mcp_events_audit (
	id UUID PRIMARY KEY,
	connection_id UUID REFERENCES mcp_connections(id) ON DELETE SET NULL,
	event_type VARCHAR(64) NOT NULL,
	source VARCHAR(32) NOT NULL,
	trace_id VARCHAR(128),
	payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mcp_webhook_nonces (
	nonce VARCHAR(128) NOT NULL,
	gateway_id VARCHAR(120) NOT NULL,
	seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	expires_at TIMESTAMPTZ NOT NULL,

	PRIMARY KEY (gateway_id, nonce)
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_mcp_connections_owner_name
	ON mcp_connections (owner_id, lower(name));

CREATE INDEX IF NOT EXISTS idx_mcp_connections_owner_mode_enabled
	ON mcp_connections (owner_id, mode, enabled, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS ux_mcp_gateway_bindings_gateway_connection
	ON mcp_gateway_bindings (gateway_id, connection_id);

CREATE INDEX IF NOT EXISTS idx_mcp_runtime_state_lifecycle
	ON mcp_runtime_state (lifecycle_state, last_transition_at DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_dispatch_commands_state_retry
	ON mcp_dispatch_commands (state, next_retry_at, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_dispatch_commands_connection_created
	ON mcp_dispatch_commands (connection_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_events_audit_connection_created
	ON mcp_events_audit (connection_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_events_audit_event_type_created
	ON mcp_events_audit (event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_mcp_webhook_nonces_expires
	ON mcp_webhook_nonces (expires_at);
