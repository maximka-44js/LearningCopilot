ALTER TABLE chats
	ADD COLUMN IF NOT EXISTS tutor_plan JSONB NOT NULL DEFAULT '{"goal":"","items":[]}'::jsonb,
	ADD COLUMN IF NOT EXISTS tutor_summary TEXT NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS tutor_validation_error_code TEXT,
	ADD COLUMN IF NOT EXISTS tutor_validation_error_field TEXT,
	ADD COLUMN IF NOT EXISTS tutor_validation_error_message TEXT;

UPDATE chats c
SET
	tutor_plan = s.tutor_plan,
	tutor_summary = s.tutor_summary,
	tutor_validation_error_code = s.tutor_validation_error_code,
	tutor_validation_error_field = s.tutor_validation_error_field,
	tutor_validation_error_message = s.tutor_validation_error_message
FROM (
	SELECT DISTINCT ON (i.chat_id)
		i.chat_id,
		s.tutor_plan,
		s.tutor_summary,
		s.tutor_validation_error_code,
		s.tutor_validation_error_field,
		s.tutor_validation_error_message
	FROM interaction_tutor_states s
	JOIN interactions i ON i.id = s.interaction_id
	ORDER BY i.chat_id, s.updated_at DESC
) s
WHERE c.id = s.chat_id;

DROP INDEX IF EXISTS idx_interaction_tutor_states_updated_at;
DROP TABLE IF EXISTS interaction_tutor_states;