CREATE TABLE IF NOT EXISTS chats (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	task_id UUID NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'active'
		CHECK (status IN ('active', 'archived', 'deleted')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	archived_at TIMESTAMPTZ,
	deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_chats_user_id
ON chats (user_id);

CREATE INDEX IF NOT EXISTS idx_chats_task_id
ON chats (task_id);

CREATE INDEX IF NOT EXISTS idx_chats_status
ON chats (status);

CREATE TABLE IF NOT EXISTS chat_messages (
	id UUID PRIMARY KEY,
	chat_id UUID NOT NULL,
	user_id UUID NOT NULL,
	interaction_id UUID NULL,
	role VARCHAR(20) NOT NULL
		CHECK (role IN ('user', 'assistant', 'system')),
	content TEXT NOT NULL
		CHECK (length(trim(content)) > 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_chat_messages_chat
		FOREIGN KEY (chat_id)
		REFERENCES chats(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_id
ON chat_messages (chat_id, created_at);

CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id
ON chat_messages (user_id);

CREATE INDEX IF NOT EXISTS idx_chat_messages_interaction_id
ON chat_messages (interaction_id);

CREATE TABLE IF NOT EXISTS interactions (
	id UUID PRIMARY KEY,
	chat_id UUID NOT NULL,
	user_message_id UUID NOT NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'created'
		CHECK (status IN (
			'created',
			'processing',
			'completed',
			'failed',
			'canceled'
		)),
	active_generation_id UUID NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	started_at TIMESTAMPTZ,
	completed_at TIMESTAMPTZ,
	failed_at TIMESTAMPTZ,
	canceled_at TIMESTAMPTZ,

	CONSTRAINT fk_interactions_chat
		FOREIGN KEY (chat_id)
		REFERENCES chats(id)
		ON DELETE CASCADE,

	CONSTRAINT fk_interactions_user_message
		FOREIGN KEY (user_message_id)
		REFERENCES chat_messages(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_interactions_chat_id
ON interactions (chat_id);

CREATE INDEX IF NOT EXISTS idx_interactions_status
ON interactions (status);

CREATE INDEX IF NOT EXISTS idx_interactions_user_message_id
ON interactions (user_message_id);

CREATE TABLE IF NOT EXISTS interaction_context_sources (
	interaction_id UUID NOT NULL,
	source_type VARCHAR(20) NOT NULL
		CHECK (source_type IN ('library', 'attached_file', 'web_search')),
	ref_id TEXT NOT NULL
		CHECK (length(trim(ref_id)) > 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_interaction_context_sources_interaction
		FOREIGN KEY (interaction_id)
		REFERENCES interactions(id)
		ON DELETE CASCADE,

	CONSTRAINT ux_interaction_context_source
		UNIQUE (interaction_id, source_type, ref_id)
);

CREATE INDEX IF NOT EXISTS idx_interaction_context_sources_interaction
ON interaction_context_sources (interaction_id);

CREATE TABLE IF NOT EXISTS generations (
	id UUID PRIMARY KEY,
	interaction_id UUID NOT NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'created'
		CHECK (status IN (
			'created',
			'streaming',
			'completed',
			'failed',
			'canceled'
		)),
	model TEXT NOT NULL,
	prompt_tokens INT NOT NULL DEFAULT 0
		CHECK (prompt_tokens >= 0),
	completion_tokens INT NOT NULL DEFAULT 0
		CHECK (completion_tokens >= 0),
	total_tokens INT NOT NULL DEFAULT 0
		CHECK (total_tokens >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	started_at TIMESTAMPTZ,
	completed_at TIMESTAMPTZ,
	failed_at TIMESTAMPTZ,
	canceled_at TIMESTAMPTZ,

	CONSTRAINT fk_generations_interaction
		FOREIGN KEY (interaction_id)
		REFERENCES interactions(id)
		ON DELETE CASCADE,

	CONSTRAINT chk_generations_total_tokens
		CHECK (total_tokens = prompt_tokens + completion_tokens)
);

CREATE INDEX IF NOT EXISTS idx_generations_interaction_id
ON generations (interaction_id);

CREATE INDEX IF NOT EXISTS idx_generations_status
ON generations (status);

ALTER TABLE interactions
	ADD CONSTRAINT fk_interactions_active_generation
	FOREIGN KEY (active_generation_id)
	REFERENCES generations(id)
	ON DELETE SET NULL;

ALTER TABLE chat_messages
	ADD CONSTRAINT fk_chat_messages_interaction
	FOREIGN KEY (interaction_id)
	REFERENCES interactions(id)
	ON DELETE SET NULL;
