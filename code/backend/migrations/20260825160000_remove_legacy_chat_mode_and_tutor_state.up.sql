DROP INDEX IF EXISTS idx_interactions_chat_mode_created_at;
DROP INDEX IF EXISTS idx_chat_messages_chat_mode_created_at;
DROP INDEX IF EXISTS idx_chats_user_active_mode_created_at;

ALTER TABLE IF EXISTS interactions
	DROP COLUMN IF EXISTS mode;

ALTER TABLE IF EXISTS chat_messages
	DROP COLUMN IF EXISTS mode;

ALTER TABLE IF EXISTS chats
	DROP COLUMN IF EXISTS active_mode;

DROP TABLE IF EXISTS interaction_tutor_states;
