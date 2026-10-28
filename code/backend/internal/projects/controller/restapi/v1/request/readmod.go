package request

// ProjectDashboardParams -.
type ProjectDashboardParams struct {
	ProjectID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// StageBoardParams -.
type StageBoardParams struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageID   string `params:"stage_id" validate:"required,uuid4" example:"223e4567-e89b-12d3-a456-426614174111"`
}

// TaskReadinessParams -.
type TaskReadinessParams struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageID   string `params:"stage_id" validate:"required,uuid4" example:"223e4567-e89b-12d3-a456-426614174111"`
	TaskID    string `params:"task_id" validate:"required,uuid4" example:"323e4567-e89b-12d3-a456-426614174222"`
}
