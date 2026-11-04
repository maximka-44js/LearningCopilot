CREATE TABLE IF NOT EXISTS interaction_tutor_states (
	interaction_id UUID PRIMARY KEY,
	tutor_plan JSONB NOT NULL DEFAULT '{"goal":"","items":[]}'::jsonb,
	tutor_summary TEXT NOT NULL DEFAULT '',
	tutor_validation_error_code TEXT,
	tutor_validation_error_field TEXT,
	tutor_validation_error_message TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_interaction_tutor_states_interaction
		FOREIGN KEY (interaction_id)
		REFERENCES interactions(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_interaction_tutor_states_updated_at
ON interaction_tutor_states (updated_at DESC);

DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_name = 'chats' AND column_name = 'tutor_plan'
	) THEN
		INSERT INTO interaction_tutor_states (
			interaction_id,
			tutor_plan,
			tutor_summary,
			tutor_validation_error_code,
			tutor_validation_error_field,
			tutor_validation_error_message,
			created_at,
			updated_at
		)
		SELECT
			i.id,
			COALESCE(c.tutor_plan, '{"goal":"","items":[]}'::jsonb),
			COALESCE(c.tutor_summary, ''),
			c.tutor_validation_error_code,
			c.tutor_validation_error_field,
			c.tutor_validation_error_message,
			now(),
			now()
		FROM chats c
		JOIN LATERAL (
			SELECT id
			FROM interactions
			WHERE chat_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) i ON TRUE
		ON CONFLICT (interaction_id) DO UPDATE
		SET tutor_plan = EXCLUDED.tutor_plan,
			tutor_summary = EXCLUDED.tutor_summary,
			tutor_validation_error_code = EXCLUDED.tutor_validation_error_code,
			tutor_validation_error_field = EXCLUDED.tutor_validation_error_field,
			tutor_validation_error_message = EXCLUDED.tutor_validation_error_message,
			updated_at = now();
	ELSIF EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_name = 'chats' AND column_name = 'tutor_state'
	) THEN
		INSERT INTO interaction_tutor_states (
			interaction_id,
			tutor_plan,
			tutor_summary,
			tutor_validation_error_code,
			tutor_validation_error_field,
			tutor_validation_error_message,
			created_at,
			updated_at
		)
		SELECT
			i.id,
			COALESCE(c.tutor_state->'plan', '{"goal":"","items":[]}'::jsonb),
			COALESCE(c.tutor_state->>'summary', ''),
			NULLIF(c.tutor_state#>>'{validation_error,code}', ''),
			NULLIF(c.tutor_state#>>'{validation_error,field}', ''),
			NULLIF(c.tutor_state#>>'{validation_error,message}', ''),
			now(),
			now()
		FROM chats c
		JOIN LATERAL (
			SELECT id
			FROM interactions
			WHERE chat_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) i ON TRUE
		ON CONFLICT (interaction_id) DO UPDATE
		SET tutor_plan = EXCLUDED.tutor_plan,
			tutor_summary = EXCLUDED.tutor_summary,
			tutor_validation_error_code = EXCLUDED.tutor_validation_error_code,
			tutor_validation_error_field = EXCLUDED.tutor_validation_error_field,
			tutor_validation_error_message = EXCLUDED.tutor_validation_error_message,
			updated_at = now();
	END IF;
END $$;

ALTER TABLE chats
	DROP COLUMN IF EXISTS tutor_validation_error_message,
	DROP COLUMN IF EXISTS tutor_validation_error_field,
	DROP COLUMN IF EXISTS tutor_validation_error_code,
	DROP COLUMN IF EXISTS tutor_summary,
	DROP COLUMN IF EXISTS tutor_plan,
	DROP COLUMN IF EXISTS tutor_state;