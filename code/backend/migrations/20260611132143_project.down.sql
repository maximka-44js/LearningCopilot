DROP TRIGGER IF EXISTS trg_task_competency_blocks_weight_sum ON task_competency_blocks;
DROP FUNCTION IF EXISTS check_task_competency_blocks_weight_sum();

DROP TABLE IF EXISTS user_project_policies;
DROP TABLE IF EXISTS assessment_snapshots;
DROP TABLE IF EXISTS task_competency_blocks;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS stages;
DROP TABLE IF EXISTS projects;
