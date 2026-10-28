CREATE TABLE IF NOT EXISTS project_policies (
	project_id UUID PRIMARY KEY,
	owner_id UUID NOT NULL,
	aggregation_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	risk_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	recommendation_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	calibration_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT fk_project_policies_project
		FOREIGN KEY (project_id)
		REFERENCES projects(id)
		ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_project_policies_owner_id
ON project_policies (owner_id);
