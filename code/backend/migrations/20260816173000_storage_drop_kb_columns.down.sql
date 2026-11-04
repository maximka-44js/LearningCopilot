ALTER TABLE files
ADD COLUMN IF NOT EXISTS source VARCHAR(32) NOT NULL DEFAULT 'user_upload'
    CHECK (source IN ('user_upload', 'parse_result')),
ADD COLUMN IF NOT EXISTS derived_from_file_id UUID NULL,
ADD COLUMN IF NOT EXISTS search_status VARCHAR(20) NOT NULL DEFAULT 'not_indexed'
    CHECK (search_status IN ('not_indexed', 'indexing', 'searchable', 'unusable')),
ADD COLUMN IF NOT EXISTS index_provider VARCHAR(64) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS index_profile VARCHAR(64) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS index_tool_hints TEXT[] NOT NULL DEFAULT ARRAY[]::text[],
ADD COLUMN IF NOT EXISTS parse_status TEXT,
ADD COLUMN IF NOT EXISTS parse_provider VARCHAR(64) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS parse_markdown_ref TEXT,
ADD COLUMN IF NOT EXISTS parse_error_message TEXT,
ADD COLUMN IF NOT EXISTS parse_updated_at TIMESTAMPTZ;

ALTER TABLE folders
ADD COLUMN IF NOT EXISTS kind VARCHAR(32) NOT NULL DEFAULT 'regular'
    CHECK (kind IN ('regular', 'parse_result_root')),
ADD COLUMN IF NOT EXISTS parse_source_file_id UUID NULL,
ADD COLUMN IF NOT EXISTS parse_provider VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_files_source_live
ON files (source)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_files_derived_from_live
ON files (derived_from_file_id)
WHERE deleted_at IS NULL AND derived_from_file_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_files_index_provider_live
ON files (index_provider)
WHERE deleted_at IS NULL AND index_provider <> '';

CREATE INDEX IF NOT EXISTS idx_folders_parse_source_live
ON folders (parse_source_file_id)
WHERE deleted_at IS NULL AND parse_source_file_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS vectorizations (
	id UUID PRIMARY KEY,
	file_id UUID NOT NULL,
	model VARCHAR(128) NOT NULL,
	chunk_size INT NOT NULL,
	chunk_overlap INT NOT NULL,
	status VARCHAR(32) NOT NULL,
	total_chunks INT NOT NULL DEFAULT 0,
	processed_chunks INT NOT NULL DEFAULT 0,
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
	chunk_index INT NOT NULL CHECK (chunk_index >= 0),
	content TEXT NOT NULL,
	search_text TEXT,
	embedding vector(1536) NOT NULL,
	page_number INT,
	start_pos INT,
	end_pos INT,
	language VARCHAR(32),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	CONSTRAINT fk_chunks_file
		FOREIGN KEY (file_id)
		REFERENCES files(id)
		ON DELETE CASCADE,
	CONSTRAINT fk_chunks_vectorization
		FOREIGN KEY (vectorization_id)
		REFERENCES vectorizations(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chunks_file_id
ON chunks (file_id);

CREATE INDEX IF NOT EXISTS idx_chunks_vectorization_id
ON chunks (vectorization_id);

CREATE INDEX IF NOT EXISTS idx_chunks_embedding
ON chunks USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

CREATE TABLE IF NOT EXISTS vectorization_batch_checkpoints (
	id UUID PRIMARY KEY,
	vectorization_id UUID NOT NULL,
	from_index INT NOT NULL CHECK (from_index >= 0),
	to_index INT NOT NULL CHECK (to_index > from_index),
	status VARCHAR(32) NOT NULL,
	error_message TEXT,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	CONSTRAINT fk_vectorization_batch_checkpoints_vectorization
		FOREIGN KEY (vectorization_id)
		REFERENCES vectorizations(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vectorization_batch_checkpoints_vectorization_id
ON vectorization_batch_checkpoints (vectorization_id);

CREATE INDEX IF NOT EXISTS idx_vectorization_batch_checkpoints_status
ON vectorization_batch_checkpoints (status);
