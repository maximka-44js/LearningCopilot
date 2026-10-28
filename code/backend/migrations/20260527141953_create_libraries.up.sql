CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS libraries (
    id UUID PRIMARY KEY,

    user_id UUID NOT NULL,

    title VARCHAR(100) NOT NULL
        CHECK (length(trim(title)) > 0),

    description TEXT,

    public BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_libraries_user_id_title
ON libraries (user_id, lower(title))
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_libraries_user_id
ON libraries (user_id);

CREATE INDEX IF NOT EXISTS idx_libraries_public
ON libraries(public)
WHERE public = TRUE;

CREATE TABLE IF NOT EXISTS library_accesses (
    library_id UUID NOT NULL
        REFERENCES libraries(id) ON DELETE CASCADE,

    user_id UUID NOT NULL,

    role VARCHAR(20) NOT NULL
        CHECK (role IN ('reader', 'writer')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (library_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_library_accesses_user_id
ON library_accesses (user_id);

CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_libraries_updated_at ON libraries;

CREATE TRIGGER set_libraries_updated_at
BEFORE UPDATE ON libraries
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();