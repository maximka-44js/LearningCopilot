CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS parse_jobs (
	id UUID PRIMARY KEY,
	file_id UUID NOT NULL,
	user_id UUID NOT NULL,
	provider TEXT NOT NULL,
	external_job_id TEXT,
	status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
		CHECK (status IN (
			'PENDING',
			'RUNNING',
			'COMPLETED',
			'FAILED',
			'CANCELLED'
		)),
	tier TEXT,
	version TEXT,
	expand JSONB,
	vectorize BOOLEAN NOT NULL DEFAULT false,
	attempts INT NOT NULL DEFAULT 0
		CHECK (attempts >= 0),
	error_message TEXT,
	markdown_ref TEXT,
	input_options JSONB,
	processing_options JSONB,
	agentic_options JSONB,
	output_options JSONB,
	crop_box JSONB,
	page_ranges JSONB,
	disable_cache BOOLEAN NOT NULL DEFAULT false,
	processing_control JSONB,
	webhook_configurations JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	started_at TIMESTAMPTZ,
	completed_at TIMESTAMPTZ,
	failed_at TIMESTAMPTZ,
	canceled_at TIMESTAMPTZ,

	CONSTRAINT fk_parse_jobs_file
		FOREIGN KEY (file_id)
		REFERENCES files(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_parse_jobs_file_id
ON parse_jobs (file_id);

CREATE INDEX IF NOT EXISTS idx_parse_jobs_user_id
ON parse_jobs (user_id);

CREATE INDEX IF NOT EXISTS idx_parse_jobs_status
ON parse_jobs (status);

CREATE INDEX IF NOT EXISTS idx_parse_jobs_external_job_id
ON parse_jobs (provider, external_job_id);

ALTER TABLE files
DROP COLUMN parse_markdown_ref,
DROP COLUMN parse_error_message,
DROP COLUMN parse_updated_at;
