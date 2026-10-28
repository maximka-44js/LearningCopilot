package request

import "github.com/maximka-44js/lc/internal/projects/entity"

// SetTaskPolicyParams -.
type SetTaskPolicyParams struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageID   string `params:"stage_id" validate:"required,uuid4" example:"223e4567-e89b-12d3-a456-426614174111"`
	TaskID    string `params:"task_id" validate:"required,uuid4" example:"323e4567-e89b-12d3-a456-426614174222"`
}

// ProjectPolicyParams -.
type ProjectPolicyParams struct {
	ProjectID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// SetTaskPolicy -.
type SetTaskPolicy struct {
	Policy entity.TaskPolicy `json:"policy" validate:"required"`
} // @name v1.SetTaskPolicy

// SetAggregationPolicy -.
type SetAggregationPolicy struct {
	Policy entity.AggregationPolicy `json:"policy" validate:"required"`
} // @name v1.SetAggregationPolicy

// SetRiskPolicy -.
type SetRiskPolicy struct {
	Policy entity.RiskPolicy `json:"policy" validate:"required"`
} // @name v1.SetRiskPolicy

// SetRecommendationPolicy -.
type SetRecommendationPolicy struct {
	Policy entity.RecommendationPolicy `json:"policy" validate:"required"`
} // @name v1.SetRecommendationPolicy

// SetCalibrationPolicy -.
type SetCalibrationPolicy struct {
	Policy entity.CalibrationPolicy `json:"policy" validate:"required"`
} // @name v1.SetCalibrationPolicy
