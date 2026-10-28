CREATE TABLE IF NOT EXISTS api_keys (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	name VARCHAR(120) NOT NULL
		CHECK (length(trim(name)) BETWEEN 1 AND 120),
	key_prefix VARCHAR(64) NOT NULL UNIQUE
		CHECK (length(trim(key_prefix)) BETWEEN 8 AND 64),
	secret_hash CHAR(64) NOT NULL
		CHECK (length(secret_hash) = 64),
	scopes JSONB NOT NULL
		CHECK (jsonb_typeof(scopes) = 'array'),
	expires_at TIMESTAMPTZ,
	last_used_at TIMESTAMPTZ,
	revoked_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id
ON api_keys (user_id);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_revoked
ON api_keys (user_id, revoked_at);

CREATE TABLE IF NOT EXISTS api_key_audit_logs (
	id UUID PRIMARY KEY,
	api_key_id UUID,
	actor_user_id UUID,
	event_type VARCHAR(20) NOT NULL
		CHECK (event_type IN ('created', 'revoked', 'rotated', 'used', 'denied')),
	meta JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_api_key_audit_logs_api_key
		FOREIGN KEY (api_key_id)
		REFERENCES api_keys(id)
		ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_api_key_audit_logs_key
ON api_key_audit_logs (api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_api_key_audit_logs_event
ON api_key_audit_logs (event_type, created_at DESC);
