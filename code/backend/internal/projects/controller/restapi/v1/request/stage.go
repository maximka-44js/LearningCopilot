package request

import (
	"time"

	"github.com/maximka-44js/lc/internal/projects/entity"
)

// AddStageParams -.
type AddStageParams struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// AddStage -.
type AddStage struct {
	Name       string     `json:"name" validate:"required,max=160" example:"Cell biology fundamentals"`
	Goal       string     `json:"goal" validate:"max=2000" example:"Build exam-ready understanding of cell structures and processes"`
	Order      int        `json:"order" validate:"required,min=1" example:"1"`
	DeadlineAt *time.Time `json:"deadline_at,omitempty" example:"2026-09-30T23:59:59Z"`
} // @name v1.AddStage

// UpdateStageParams -.
type UpdateStageParams struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageID   string `params:"stage_id" validate:"required,uuid4" example:"223e4567-e89b-12d3-a456-426614174111"`
}

// GetStageByID -.
type GetStageByID struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageID   string `params:"stage_id" validate:"required,uuid4" example:"223e4567-e89b-12d3-a456-426614174111"`
}

// UpdateStage -.
type UpdateStage struct {
	Name       string     `json:"name" validate:"required,max=160" example:"Cell biology and signaling"`
	Goal       string     `json:"goal" validate:"max=2000" example:"Update milestones based on diagnostic results"`
	Order      int        `json:"order" validate:"required,min=1" example:"2"`
	DeadlineAt *time.Time `json:"deadline_at,omitempty" example:"2026-10-15T23:59:59Z"`
} // @name v1.UpdateStage

// RecalculateStageReadiness -.
type RecalculateStageReadiness struct {
	At                *time.Time                `json:"at,omitempty" example:"2026-06-12T10:00:00Z"`
	AggregationPolicy *entity.AggregationPolicy `json:"aggregation_policy,omitempty"`
} // @name v1.RecalculateStageReadiness
