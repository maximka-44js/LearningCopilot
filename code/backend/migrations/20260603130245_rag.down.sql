DROP INDEX IF EXISTS idx_chunks_embedding;
DROP INDEX IF EXISTS idx_chunks_vectorization_id;
DROP INDEX IF EXISTS idx_chunks_file_id;
DROP TABLE IF EXISTS chunks;

DROP INDEX IF EXISTS idx_vectorizations_status;
DROP INDEX IF EXISTS idx_vectorizations_file_id;
DROP TABLE IF EXISTS vectorizations;
