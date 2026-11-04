CREATE TABLE IF NOT EXISTS user_agents (
	id UUID PRIMARY KEY,
	owner_id UUID NOT NULL,
	name VARCHAR(120) NOT NULL CHECK (length(trim(name)) > 0),
	runtime_url TEXT NOT NULL CHECK (length(trim(runtime_url)) > 0),
	auth_mode VARCHAR(16) NOT NULL CHECK (auth_mode IN ('none', 'hmac')),
	secret_ciphertext BYTEA NOT NULL DEFAULT ''::bytea,
	secret_nonce BYTEA NOT NULL DEFAULT ''::bytea,
	secret_key_version VARCHAR(32) NOT NULL DEFAULT 'v1',
	timeout_ms INT NOT NULL CHECK (timeout_ms > 0),
	retry_max INT NOT NULL CHECK (retry_max >= 0 AND retry_max <= 10),
	status VARCHAR(16) NOT NULL CHECK (status IN ('active', 'disabled')),
	disabled_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_user_agents_owner_name
	ON user_agents (owner_id, lower(name));

CREATE INDEX IF NOT EXISTS idx_user_agents_owner_status
	ON user_agents (owner_id, status, created_at DESC);
