ALTER TABLE user_agents
	ADD COLUMN IF NOT EXISTS response_format VARCHAR(32) NOT NULL DEFAULT 'vercel',
	ADD CONSTRAINT user_agents_response_format_chk CHECK (response_format IN ('vercel'));
