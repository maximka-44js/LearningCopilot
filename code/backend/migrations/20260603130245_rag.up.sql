CREATE TABLE IF NOT EXISTS vectorizations (
	id UUID PRIMARY KEY,
	file_id UUID NOT NULL,
	model TEXT NOT NULL,
	chunk_size INT NOT NULL
		CHECK (chunk_size >= 100 AND chunk_size <= 10000),
	chunk_overlap INT NOT NULL
		CHECK (chunk_overlap >= 0 AND chunk_overlap <= (chunk_size / 2)),
	status VARCHAR(20) NOT NULL DEFAULT 'created'
		CHECK (status IN (
			'created',
			'queued',
			'processing',
			'completed',
			'failed',
			'canceled'
		)),
	total_chunks INT NOT NULL DEFAULT 0
		CHECK (total_chunks >= 0),
	processed_chunks INT NOT NULL DEFAULT 0
		CHECK (processed_chunks >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	started_at TIMESTAMPTZ,
	completed_at TIMESTAMPTZ,
	failed_at TIMESTAMPTZ,
	canceled_at TIMESTAMPTZ,

	CONSTRAINT fk_vectorizations_file
		FOREIGN KEY (file_id)
		REFERENCES files(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vectorizations_file_id
ON vectorizations (file_id);

CREATE INDEX IF NOT EXISTS idx_vectorizations_status
ON vectorizations (status);

CREATE TABLE IF NOT EXISTS chunks (
	id UUID PRIMARY KEY,
	file_id UUID NOT NULL,
	vectorization_id UUID NOT NULL,
	chunk_index INT NOT NULL
		CHECK (chunk_index >= 0),
	content TEXT NOT NULL
		CHECK (length(trim(content)) > 0),
	embedding vector(384) NOT NULL,
	metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_chunks_file
		FOREIGN KEY (file_id)
		REFERENCES files(id)
		ON DELETE CASCADE,

	CONSTRAINT fk_chunks_vectorization
		FOREIGN KEY (vectorization_id)
		REFERENCES vectorizations(id)
		ON DELETE CASCADE,

	CONSTRAINT ux_chunks_vectorization_index
		UNIQUE (vectorization_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_chunks_file_id
ON chunks (file_id);

CREATE INDEX IF NOT EXISTS idx_chunks_vectorization_id
ON chunks (vectorization_id);

CREATE INDEX IF NOT EXISTS idx_chunks_embedding
ON chunks USING hnsw (embedding vector_cosine_ops);
