CREATE TABLE IF NOT EXISTS integration_signing_secrets (
	owner_id UUID NOT NULL,
	integration_kind VARCHAR(16) NOT NULL CHECK (integration_kind IN ('addon', 'agent', 'mcp')),
	integration_id UUID NOT NULL,
	key_id VARCHAR(120) NOT NULL DEFAULT 'default',
	cipher BYTEA NOT NULL,
	nonce BYTEA NOT NULL,
	key_version VARCHAR(32) NOT NULL DEFAULT 'v1',
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (owner_id, integration_kind, integration_id)
);

CREATE INDEX IF NOT EXISTS idx_integration_signing_secrets_updated
	ON integration_signing_secrets (updated_at DESC);