-- Удаляем оба индекса
DROP INDEX IF EXISTS idx_unique_active_folder_name;
DROP INDEX IF EXISTS idx_unique_active_root_folder_name;

-- Возвращаем старый индекс
CREATE UNIQUE INDEX idx_unique_active_folder_name 
    ON folders(library_id, parent_id, name) 
    WHERE deleted_at IS NULL;