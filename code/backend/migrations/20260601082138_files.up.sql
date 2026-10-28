CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS files (
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL,
	library_id UUID NOT NULL,
	folder_id UUID NULL,
	name VARCHAR(100) NOT NULL
		CHECK (length(trim(name)) > 0),
	size BIGINT NOT NULL
		CHECK (size >= 0 AND size <= 524288000),
	mime_type TEXT NOT NULL,
	checksum_declared TEXT,
	checksum TEXT,
	status VARCHAR(20) NOT NULL DEFAULT 'created'
		CHECK (status IN (
			'created',
			'uploaded',
			'scan_pending',
			'scan_ok',
			'scan_failed',
			'rejected',
			'deleted'
		)),
	search_status VARCHAR(20) NOT NULL DEFAULT 'not_indexed'
		CHECK (search_status IN (
			'not_indexed',
			'indexing',
			'searchable',
			'unusable'
		)),
	scan_attempts INT NOT NULL DEFAULT 0
		CHECK (scan_attempts >= 0),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	deleted_at TIMESTAMPTZ,

	CONSTRAINT fk_files_library
		FOREIGN KEY (library_id)
		REFERENCES libraries(id)
		ON DELETE CASCADE,

	CONSTRAINT fk_files_folder
		FOREIGN KEY (folder_id)
		REFERENCES folders(id)
		ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_files_library_folder_name
ON files (library_id, folder_id, lower(name))
WHERE deleted_at IS NULL AND folder_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_files_library_root_name
ON files (library_id, lower(name))
WHERE deleted_at IS NULL AND folder_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_files_library_id
ON files (library_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_files_folder_id
ON files (folder_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_files_user_id
ON files (user_id);

CREATE INDEX IF NOT EXISTS idx_files_status
ON files (status)
WHERE deleted_at IS NULL;
