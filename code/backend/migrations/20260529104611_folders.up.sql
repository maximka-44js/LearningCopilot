CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS folders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    library_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id UUID NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT fk_folders_library 
        FOREIGN KEY (library_id) 
        REFERENCES libraries(id) 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_folders_parent 
        FOREIGN KEY (parent_id) 
        REFERENCES folders(id) 
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_unique_active_folder_name 
    ON folders(library_id, parent_id, name) 
    WHERE deleted_at IS NULL;

CREATE INDEX idx_folders_user_id ON folders(user_id);
CREATE INDEX idx_folders_library_id ON folders(library_id);
CREATE INDEX idx_folders_parent_id ON folders(parent_id);