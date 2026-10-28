CREATE TABLE IF NOT EXISTS projects (
	id UUID PRIMARY KEY,
	owner_id UUID NOT NULL,
	name VARCHAR(200) NOT NULL
		CHECK (length(trim(name)) BETWEEN 1 AND 200),
	description TEXT NOT NULL DEFAULT ''
		CHECK (length(trim(description)) <= 4000),
	deadline_at TIMESTAMPTZ,
	priority VARCHAR(20) NOT NULL DEFAULT 'normal'
		CHECK (priority IN ('low', 'normal', 'high')),
	status VARCHAR(20) NOT NULL DEFAULT 'active'
		CHECK (status IN ('active', 'paused', 'archived')),
	coverage NUMERIC(5,2) NOT NULL DEFAULT 0
		CHECK (coverage >= 0 AND coverage <= 100),
	readiness NUMERIC(5,2) NOT NULL DEFAULT 0
		CHECK (readiness >= 0 AND readiness <= 100),
	risk VARCHAR(20) NOT NULL DEFAULT 'medium'
		CHECK (risk IN ('low', 'medium', 'high', 'critical')),
	metadata JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	archived_at TIMESTAMPTZ,

	CONSTRAINT chk_projects_deadline_not_before_created
		CHECK (deadline_at IS NULL OR deadline_at >= created_at),
	CONSTRAINT chk_projects_archived_at_state
		CHECK (
			(status = 'archived' AND archived_at IS NOT NULL)
			OR (status <> 'archived' AND archived_at IS NULL)
		)
);

CREATE INDEX IF NOT EXISTS idx_projects_owner_id
ON projects (owner_id);

CREATE INDEX IF NOT EXISTS idx_projects_owner_status
ON projects (owner_id, status);

CREATE TABLE IF NOT EXISTS stages (
	id UUID PRIMARY KEY,
	project_id UUID NOT NULL,
	name VARCHAR(160) NOT NULL
		CHECK (length(trim(name)) BETWEEN 1 AND 160),
	goal TEXT NOT NULL DEFAULT ''
		CHECK (length(trim(goal)) <= 2000),
	stage_order INT NOT NULL
		CHECK (stage_order >= 1),
	status VARCHAR(20) NOT NULL DEFAULT 'active'
		CHECK (status IN ('active', 'paused', 'archived')),
	deadline_at TIMESTAMPTZ,
	coverage NUMERIC(5,2) NOT NULL DEFAULT 0
		CHECK (coverage >= 0 AND coverage <= 100),
	readiness NUMERIC(5,2) NOT NULL DEFAULT 0
		CHECK (readiness >= 0 AND readiness <= 100),
	risk VARCHAR(20) NOT NULL DEFAULT 'medium'
		CHECK (risk IN ('low', 'medium', 'high', 'critical')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	archived_at TIMESTAMPTZ,

	CONSTRAINT fk_stages_project
		FOREIGN KEY (project_id)
		REFERENCES projects(id)
		ON DELETE CASCADE,
	CONSTRAINT ux_stages_project_order
		UNIQUE (project_id, stage_order),
	CONSTRAINT chk_stages_deadline_not_before_created
		CHECK (deadline_at IS NULL OR deadline_at >= created_at),
	CONSTRAINT chk_stages_archived_at_state
		CHECK (
			(status = 'archived' AND archived_at IS NOT NULL)
			OR (status <> 'archived' AND archived_at IS NULL)
		)
);

CREATE INDEX IF NOT EXISTS idx_stages_project_id
ON stages (project_id);

CREATE INDEX IF NOT EXISTS idx_stages_project_status
ON stages (project_id, status);

CREATE TABLE IF NOT EXISTS tasks (
	id UUID PRIMARY KEY,
	stage_id UUID NOT NULL,
	outcome TEXT NOT NULL
		CHECK (length(trim(outcome)) BETWEEN 1 AND 2000),
	difficulty INT NOT NULL
		CHECK (difficulty >= 1 AND difficulty <= 10),
	priority VARCHAR(20) NOT NULL DEFAULT 'normal'
		CHECK (priority IN ('low', 'normal', 'high')),
	exam_ready_min_mastery NUMERIC(5,2) NOT NULL DEFAULT 70
		CHECK (exam_ready_min_mastery > 0 AND exam_ready_min_mastery <= 100),
	require_non_critical_block_risk BOOLEAN NOT NULL DEFAULT TRUE,
	deadline_at TIMESTAMPTZ,
	status VARCHAR(30) NOT NULL DEFAULT 'backlog'
		CHECK (
			status IN (
				'backlog',
				'learning',
				'baseline_passed',
				'at_risk',
				'stable',
				'exam_ready',
				'archived'
			)
		),
	coverage NUMERIC(5,2) NOT NULL DEFAULT 0
		CHECK (coverage >= 0 AND coverage <= 100),
	readiness NUMERIC(5,2) NOT NULL DEFAULT 0
		CHECK (readiness >= 0 AND readiness <= 100),
	risk VARCHAR(20) NOT NULL DEFAULT 'medium'
		CHECK (risk IN ('low', 'medium', 'high', 'critical')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	archived_at TIMESTAMPTZ,

	CONSTRAINT fk_tasks_stage
		FOREIGN KEY (stage_id)
		REFERENCES stages(id)
		ON DELETE CASCADE,
	CONSTRAINT chk_tasks_deadline_not_before_created
		CHECK (deadline_at IS NULL OR deadline_at >= created_at),
	CONSTRAINT chk_tasks_archived_at_state
		CHECK (
			(status = 'archived' AND archived_at IS NOT NULL)
			OR (status <> 'archived' AND archived_at IS NULL)
		)
);

CREATE INDEX IF NOT EXISTS idx_tasks_stage_id
ON tasks (stage_id);

CREATE INDEX IF NOT EXISTS idx_tasks_stage_status
ON tasks (stage_id, status);

CREATE TABLE IF NOT EXISTS task_competency_blocks (
	task_id UUID NOT NULL,
	block_type VARCHAR(20) NOT NULL
		CHECK (block_type IN ('comprehension', 'memorization', 'practice')),
	weight INT NOT NULL
		CHECK (weight >= 1 AND weight <= 100),
	mastery NUMERIC(5,2) NOT NULL DEFAULT 0
		CHECK (mastery >= 0 AND mastery <= 100),
	last_assessed_at TIMESTAMPTZ,
	decay_profile TEXT,
	risk_state VARCHAR(20) NOT NULL DEFAULT 'medium'
		CHECK (risk_state IN ('low', 'medium', 'high', 'critical')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT pk_task_competency_blocks
		PRIMARY KEY (task_id, block_type),
	CONSTRAINT fk_task_competency_blocks_task
		FOREIGN KEY (task_id)
		REFERENCES tasks(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_competency_blocks_task_id
ON task_competency_blocks (task_id);

CREATE TABLE IF NOT EXISTS assessment_snapshots (
	id UUID PRIMARY KEY,
	task_id UUID NOT NULL,
	block_type VARCHAR(20) NOT NULL
		CHECK (block_type IN ('comprehension', 'memorization', 'practice')),
	source VARCHAR(30) NOT NULL
		CHECK (source IN ('dialog', 'mini_case', 'external_system', 'manual_confirmation')),
	raw_score NUMERIC(5,2) NOT NULL
		CHECK (raw_score >= 0 AND raw_score <= 100),
	normalized_score NUMERIC(5,2) NOT NULL
		CHECK (normalized_score >= 0 AND normalized_score <= 100),
	confidence_before NUMERIC(5,2) NOT NULL
		CHECK (confidence_before >= 0 AND confidence_before <= 100),
	confidence_after NUMERIC(5,2) NOT NULL
		CHECK (confidence_after >= 0 AND confidence_after <= 100),
	recorded_at TIMESTAMPTZ NOT NULL,
	trace_id VARCHAR(128),
	source_reference TEXT,
	source_reference_type VARCHAR(64),
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_assessment_snapshots_task
		FOREIGN KEY (task_id)
		REFERENCES tasks(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_assessment_snapshots_task_recorded
ON assessment_snapshots (task_id, recorded_at DESC);

CREATE INDEX IF NOT EXISTS idx_assessment_snapshots_task_block_recorded
ON assessment_snapshots (task_id, block_type, recorded_at DESC);

CREATE TABLE IF NOT EXISTS user_project_policies (
	owner_id UUID PRIMARY KEY,
	aggregation_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	risk_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	recommendation_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	calibration_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION check_task_competency_blocks_weight_sum()
RETURNS trigger
AS $$
DECLARE
	task_id_to_check UUID;
	total_weight INT;
	blocks_count INT;
BEGIN
	task_id_to_check := coalesce(NEW.task_id, OLD.task_id);

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

DROP TRIGGER IF EXISTS trg_task_competency_blocks_weight_sum ON task_competency_blocks;

CREATE CONSTRAINT TRIGGER trg_task_competency_blocks_weight_sum
AFTER INSERT OR UPDATE OR DELETE
ON task_competency_blocks
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION check_task_competency_blocks_weight_sum();
