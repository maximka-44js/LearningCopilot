package request

import (
	"time"

	"github.com/maximka-44js/lc/internal/projects/entity"
)

// AddTaskParams -.
type AddTaskParams struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageID   string `params:"stage_id" validate:"required,uuid4" example:"223e4567-e89b-12d3-a456-426614174111"`
}

// TaskIDParams -.
type TaskIDParams struct {
	ProjectID string `params:"project_id" validate:"required,uuid4" example:"123e4567-e89b-12d3-a456-426614174000"`
	StageID   string `params:"stage_id" validate:"required,uuid4" example:"223e4567-e89b-12d3-a456-426614174111"`
	TaskID    string `params:"task_id" validate:"required,uuid4" example:"323e4567-e89b-12d3-a456-426614174222"`
}

// TaskCompetencyBlock -.
type TaskCompetencyBlock struct {
	Type   entity.CompetencyBlockType `json:"type" validate:"required,oneof=comprehension memorization practice" example:"comprehension"`
	Weight int                        `json:"weight" validate:"required,min=1,max=100" example:"40"`
}

// AddTask -.
type AddTask struct {
	Outcome    string                 `json:"outcome" validate:"required,max=2000" example:"Explain immune response mechanisms and solve case questions"`
	Difficulty int                    `json:"difficulty" validate:"required,min=1,max=10" example:"6"`
	Priority   entity.ProjectPriority `json:"priority" validate:"required,oneof=low normal high" example:"normal"`
	Policy     *entity.TaskPolicy     `json:"policy,omitempty"`
	DeadlineAt *time.Time             `json:"deadline_at,omitempty" example:"2026-10-20T23:59:59Z"`
	Blocks     []TaskCompetencyBlock  `json:"blocks" validate:"required,min=1,dive"`
} // @name v1.AddTask

// UpdateTask -.
type UpdateTask struct {
	Outcome    string                 `json:"outcome" validate:"required,max=2000" example:"Explain immune response mechanisms using mixed exam formats"`
	Difficulty int                    `json:"difficulty" validate:"required,min=1,max=10" example:"7"`
	Priority   entity.ProjectPriority `json:"priority" validate:"required,oneof=low normal high" example:"high"`
	DeadlineAt *time.Time             `json:"deadline_at,omitempty" example:"2026-11-01T23:59:59Z"`
} // @name v1.UpdateTask

// TaskWeightDistribution -.
type TaskWeightDistribution struct {
	Comprehension int `json:"comprehension" validate:"required,min=1,max=100" example:"40"`
	Memorization  int `json:"memorization" validate:"required,min=1,max=100" example:"30"`
	Practice      int `json:"practice" validate:"required,min=1,max=100" example:"30"`
} // @name v1.TaskWeightDistribution

// ConfigureTaskWeights -.
type ConfigureTaskWeights struct {
	Weights TaskWeightDistribution `json:"weights" validate:"required"`
} // @name v1.ConfigureTaskWeights

// RecordAssessmentSnapshot -.
type RecordAssessmentSnapshot struct {
	BlockType           entity.CompetencyBlockType `json:"block_type" validate:"required,oneof=comprehension memorization practice" example:"practice"`
	Source              entity.AssessmentSource    `json:"source" validate:"required,oneof=dialog mini_case external_system manual_confirmation" example:"dialog"`
	RawScore            float64                    `json:"raw_score" validate:"gte=0,lte=100" example:"78.5"`
	NormalizedScore     float64                    `json:"normalized_score" validate:"gte=0,lte=100" example:"75.2"`
	ConfidenceBefore    float64                    `json:"confidence_before" validate:"gte=0,lte=100" example:"62"`
	ConfidenceAfter     float64                    `json:"confidence_after" validate:"gte=0,lte=100" example:"71"`
	RecordedAt          *time.Time                 `json:"recorded_at,omitempty" example:"2026-06-12T12:00:00Z"`
	TraceID             string                     `json:"trace_id,omitempty" validate:"max=128" example:"trace-9f2a"`
	SourceReference     string                     `json:"source_reference,omitempty" validate:"max=1024" example:"dialog:session-13"`
	SourceReferenceType string                     `json:"source_reference_type,omitempty" validate:"max=64" example:"manual_review"`
} // @name v1.RecordAssessmentSnapshot

// RecalculateTaskReadiness -.
type RecalculateTaskReadiness struct {
	At         *time.Time                                         `json:"at,omitempty" example:"2026-06-12T13:00:00Z"`
	Profiles   map[entity.CompetencyBlockType]entity.DecayProfile `json:"profiles,omitempty"`
	RiskPolicy *entity.RiskPolicy                                 `json:"risk_policy,omitempty"`
} // @name v1.RecalculateTaskReadiness

// GenerateRecommendations -.
type GenerateRecommendations struct {
	RecommendationPolicy *entity.RecommendationPolicy `json:"recommendation_policy,omitempty"`
	CalibrationPolicy    *entity.CalibrationPolicy    `json:"calibration_policy,omitempty"`
} // @name v1.GenerateRecommendations

// TransitionTaskStatus -.
type TransitionTaskStatus struct {
	NextStatus entity.TaskStatus `json:"next_status" validate:"required,oneof=backlog learning baseline_passed at_risk stable exam_ready archived" example:"learning"`
} // @name v1.TransitionTaskStatus
