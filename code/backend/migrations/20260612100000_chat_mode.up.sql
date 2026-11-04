ALTER TABLE chat_messages
	ADD COLUMN IF NOT EXISTS mode VARCHAR(20) NOT NULL DEFAULT 'tutor'
		CHECK (mode IN ('tutor', 'examine'));

ALTER TABLE interactions
	ADD COLUMN IF NOT EXISTS mode VARCHAR(20) NOT NULL DEFAULT 'tutor'
		CHECK (mode IN ('tutor', 'examine'));

ALTER TABLE chats
	ADD COLUMN IF NOT EXISTS active_mode VARCHAR(20) NOT NULL DEFAULT 'tutor'
		CHECK (active_mode IN ('tutor', 'examine'));

CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_mode_created_at
ON chat_messages (chat_id, mode, created_at);

CREATE INDEX IF NOT EXISTS idx_interactions_chat_mode_created_at
ON interactions (chat_id, mode, created_at);

CREATE INDEX IF NOT EXISTS idx_chats_user_active_mode_created_at
ON chats (user_id, active_mode, created_at DESC);
