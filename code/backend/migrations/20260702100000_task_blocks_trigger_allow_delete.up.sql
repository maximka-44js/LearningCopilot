CREATE OR REPLACE FUNCTION check_task_competency_blocks_weight_sum()
RETURNS trigger
AS $$
DECLARE
	task_id_to_check UUID;
	total_weight INT;
	blocks_count INT;
BEGIN
	task_id_to_check := coalesce(NEW.task_id, OLD.task_id);

	IF task_id_to_check IS NULL THEN
		RETURN NULL;
	END IF;

	IF NOT EXISTS (SELECT 1 FROM tasks WHERE id = task_id_to_check) THEN
		RETURN NULL;
	END IF;

	SELECT coalesce(sum(weight), 0), count(*)
	INTO total_weight, blocks_count
	FROM task_competency_blocks
	WHERE task_id = task_id_to_check;

	IF blocks_count = 0 THEN
		RAISE EXCEPTION 'Task % must contain at least one competency block', task_id_to_check;
	END IF;

	IF total_weight <> 100 THEN
		RAISE EXCEPTION 'Task % competency block weight sum must be 100, actual %', task_id_to_check, total_weight;
	END IF;

	RETURN NULL;
END;
$$ LANGUAGE plpgsql;
