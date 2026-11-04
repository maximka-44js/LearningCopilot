CREATE TABLE IF NOT EXISTS vectorization_batch_checkpoints (
	id UUID PRIMARY KEY,
	vectorization_id UUID NOT NULL,
	from_index INT NOT NULL CHECK (from_index >= 0),
	to_index INT NOT NULL CHECK (to_index > from_index),
	status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'done')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	done_at TIMESTAMPTZ,

	CONSTRAINT fk_vectorization_batch_checkpoints_vectorization
		FOREIGN KEY (vectorization_id)
		REFERENCES vectorizations(id)
		ON DELETE CASCADE,

	CONSTRAINT ux_vectorization_batch_checkpoint_window
		UNIQUE (vectorization_id, from_index, to_index)
);

CREATE INDEX IF NOT EXISTS idx_vectorization_batch_checkpoints_vectorization
ON vectorization_batch_checkpoints (vectorization_id);

CREATE INDEX IF NOT EXISTS idx_vectorization_batch_checkpoints_status
ON vectorization_batch_checkpoints (vectorization_id, status, from_index);
