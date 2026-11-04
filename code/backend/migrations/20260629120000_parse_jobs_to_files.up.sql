ALTER TABLE files
ADD COLUMN parse_markdown_ref TEXT,
ADD COLUMN parse_error_message TEXT,
ADD COLUMN parse_updated_at TIMESTAMPTZ;

UPDATE files AS f
SET parse_status = p.status,
    parse_markdown_ref = p.markdown_ref,
    parse_error_message = p.error_message,
    parse_updated_at = p.updated_at
FROM (
    SELECT DISTINCT ON (file_id)
        file_id,
        status,
        markdown_ref,
        error_message,
        updated_at,
        created_at
    FROM parse_jobs
    ORDER BY file_id, updated_at DESC, created_at DESC
) AS p
WHERE p.file_id = f.id;

DROP TABLE IF EXISTS parse_jobs;
