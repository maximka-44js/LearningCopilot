ALTER TABLE user_agents
	ADD COLUMN IF NOT EXISTS supports_custom_mcp BOOLEAN NOT NULL DEFAULT false;
