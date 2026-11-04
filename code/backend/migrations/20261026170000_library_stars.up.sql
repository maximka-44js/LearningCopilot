CREATE TABLE IF NOT EXISTS library_stars (
    library_id UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (library_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_library_stars_library_id ON library_stars (library_id);
