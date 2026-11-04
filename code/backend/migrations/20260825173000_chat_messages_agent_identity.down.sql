DROP INDEX IF EXISTS idx_chat_messages_agent_id;

ALTER TABLE chat_messages
	DROP COLUMN IF EXISTS agent_name,
	DROP COLUMN IF EXISTS agent_id;
