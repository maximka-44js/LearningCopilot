DROP INDEX IF EXISTS idx_interactions_chat_mode_created_at;
DROP INDEX IF EXISTS idx_chat_messages_chat_mode_created_at;
DROP INDEX IF EXISTS idx_chats_user_active_mode_created_at;

ALTER TABLE chats
	DROP COLUMN IF EXISTS active_mode;

ALTER TABLE interactions
	DROP COLUMN IF EXISTS mode;

ALTER TABLE chat_messages
	DROP COLUMN IF EXISTS mode;
