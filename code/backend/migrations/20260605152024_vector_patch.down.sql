DROP INDEX IF EXISTS idx_chunks_embedding;

ALTER TABLE chunks
	ALTER COLUMN embedding TYPE vector(384);

CREATE INDEX IF NOT EXISTS idx_chunks_embedding
ON chunks USING hnsw (embedding vector_cosine_ops);
