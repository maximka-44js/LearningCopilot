ALTER TABLE user_agents
	DROP CONSTRAINT IF EXISTS user_agents_response_format_chk,
	DROP COLUMN IF EXISTS response_format;
