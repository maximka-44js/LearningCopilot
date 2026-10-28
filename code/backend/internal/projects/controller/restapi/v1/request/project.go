package request

import (
	"time"

	"github.com/maximka-44js/lc/internal/projects/entity"
)

// CreateProject -.
type CreateProject struct {
	Name        string                 `json:"name" validate:"required,max=200" example:"Prepare for microbiology exam"`
	Description string                 `json:"description" validate:"max=4000" example:"Weekly study plan with milestones"`
	DeadlineAt  *time.Time             `json:"deadline_at,omitempty" example:"2026-12-31T23:59:59Z"`
	Priority    entity.ProjectPriority `json:"priority" validate:"required,oneof=low normal high" example:"normal"`
	Metadata    map[string]string      `json:"metadata" example:"source:planner,track:biology"`
} // @name v1.CreateProject

// UpdateProjectParams -.
type UpdateProjectParams struct {
	ProjectID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// GetProjectByID -.
type GetProjectByID struct {
	ProjectID string `params:"id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// UpdateProject -.
type UpdateProject struct {
	Name        string                 `json:"name" validate:"required,max=200" example:"Prepare for microbiology exam"`
	Description string                 `json:"description" validate:"max=4000" example:"Updated plan after diagnostics"`
	DeadlineAt  *time.Time             `json:"deadline_at,omitempty" example:"2026-12-31T23:59:59Z"`
	Priority    entity.ProjectPriority `json:"priority" validate:"required,oneof=low normal high" example:"high"`
	Metadata    map[string]string      `json:"metadata" example:"source:planner,track:biology"`
} // @name v1.UpdateProject

// RecalculateProjectReadiness -.
type RecalculateProjectReadiness struct {
	At                *time.Time                `json:"at,omitempty" example:"2026-06-12T10:00:00Z"`
	AggregationPolicy *entity.AggregationPolicy `json:"aggregation_policy,omitempty"`
} // @name v1.RecalculateProjectReadiness
