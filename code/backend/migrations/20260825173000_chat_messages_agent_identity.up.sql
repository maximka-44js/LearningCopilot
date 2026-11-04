ALTER TABLE chat_messages
	ADD COLUMN IF NOT EXISTS agent_id UUID,
	ADD COLUMN IF NOT EXISTS agent_name TEXT;

CREATE INDEX IF NOT EXISTS idx_chat_messages_agent_id
ON chat_messages (agent_id);
