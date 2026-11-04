DROP INDEX IF EXISTS idx_chunks_embedding;

ALTER TABLE chunks
	ALTER COLUMN embedding TYPE vector(1536);

CREATE INDEX IF NOT EXISTS idx_chunks_embedding
ON chunks USING hnsw (embedding vector_cosine_ops);
