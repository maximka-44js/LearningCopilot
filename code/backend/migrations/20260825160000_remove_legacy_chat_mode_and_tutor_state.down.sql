CREATE TABLE IF NOT EXISTS interaction_tutor_states (
	interaction_id UUID PRIMARY KEY,
	tutor_plan JSONB NOT NULL DEFAULT '{"goal":"","items":[]}'::jsonb,
	tutor_summary TEXT NOT NULL DEFAULT '',
	tutor_validation_error_code TEXT,
	tutor_validation_error_field TEXT,
	tutor_validation_error_message TEXT,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT fk_interaction_tutor_states_interaction
		FOREIGN KEY (interaction_id) REFERENCES interactions(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_interaction_tutor_states_updated_at
ON interaction_tutor_states (updated_at DESC);

ALTER TABLE chats
	ADD COLUMN IF NOT EXISTS active_mode VARCHAR(20) NOT NULL DEFAULT 'tutor'
		CHECK (active_mode IN ('tutor', 'examine'));

ALTER TABLE chat_messages
	ADD COLUMN IF NOT EXISTS mode VARCHAR(20) NOT NULL DEFAULT 'tutor'
		CHECK (mode IN ('tutor', 'examine'));

ALTER TABLE interactions
	ADD COLUMN IF NOT EXISTS mode VARCHAR(20) NOT NULL DEFAULT 'tutor'
		CHECK (mode IN ('tutor', 'examine'));

CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_mode_created_at
ON chat_messages (chat_id, mode, created_at);

CREATE INDEX IF NOT EXISTS idx_interactions_chat_mode_created_at
ON interactions (chat_id, mode, created_at);

CREATE INDEX IF NOT EXISTS idx_chats_user_active_mode_created_at
ON chats (user_id, active_mode, created_at DESC);
