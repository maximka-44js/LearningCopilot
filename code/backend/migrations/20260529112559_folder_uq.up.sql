-- Удаляем старый индекс, который не работал для корневых папок
DROP INDEX IF EXISTS idx_unique_active_folder_name;

-- Создаём индекс для вложенных папок
CREATE UNIQUE INDEX idx_unique_active_folder_name 
    ON folders(library_id, parent_id, name) 
    WHERE deleted_at IS NULL AND parent_id IS NOT NULL;

-- Создаём индекс для корневых папок
CREATE UNIQUE INDEX idx_unique_active_root_folder_name 
    ON folders(library_id, name) 
    WHERE deleted_at IS NULL AND parent_id IS NULL;

