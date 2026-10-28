package repo

import (
	"context"
	"time"

	"github.com/maximka-44js/lc/internal/projects/entity"
)

type (
	ProjectRepo interface {
		Store(ctx context.Context, project *entity.Project) error
		GetByID(ctx context.Context, ownerID, projectID string) (entity.Project, error)
		ListByOwner(ctx context.Context, ownerID string) ([]entity.Project, error)
		Update(ctx context.Context, project *entity.Project) error
		Delete(ctx context.Context, ownerID, projectID string) error
	}

	StageRepo interface {
		Store(ctx context.Context, stage *entity.Stage) error
		GetByID(ctx context.Context, projectID, stageID string) (entity.Stage, error)
		ListByProject(ctx context.Context, projectID string) ([]entity.Stage, error)
		Update(ctx context.Context, stage *entity.Stage) error
		Delete(ctx context.Context, projectID, stageID string) error
	}

	TaskRepo interface {
		Store(ctx context.Context, task *entity.Task) error
		GetByID(ctx context.Context, stageID, taskID string) (entity.Task, error)
		GetByTaskID(ctx context.Context, taskID string) (entity.Task, error)
		GetScopeByTaskID(ctx context.Context, taskID string) (projectID, stageID string, err error)
		ListByStage(ctx context.Context, stageID string) ([]entity.Task, error)
		Update(ctx context.Context, task *entity.Task) error
		Delete(ctx context.Context, stageID, taskID string) error
	}

	AssessmentSnapshotRepo interface {
		Store(ctx context.Context, snapshot *entity.AssessmentSnapshot) error
		ListByTask(ctx context.Context, taskID string, filter AssessmentSnapshotFilter) ([]entity.AssessmentSnapshot, error)
		GetLatestByTask(ctx context.Context, taskID string) (*entity.AssessmentSnapshot, error)
		GetLatestByTaskAndBlock(ctx context.Context, taskID string, blockType entity.CompetencyBlockType) (*entity.AssessmentSnapshot, error)
	}

	UserPolicyRepo interface {
		SetAggregationPolicy(ctx context.Context, ownerID string, policy entity.AggregationPolicy) error
		SetRiskPolicy(ctx context.Context, ownerID string, policy entity.RiskPolicy) error
		SetRecommendationPolicy(ctx context.Context, ownerID string, policy entity.RecommendationPolicy) error
		SetCalibrationPolicy(ctx context.Context, ownerID string, policy entity.CalibrationPolicy) error

		GetAggregationPolicy(ctx context.Context, ownerID string) (entity.AggregationPolicy, error)
		GetRiskPolicy(ctx context.Context, ownerID string) (entity.RiskPolicy, error)
		GetRecommendationPolicy(ctx context.Context, ownerID string) (entity.RecommendationPolicy, error)
		GetCalibrationPolicy(ctx context.Context, ownerID string) (entity.CalibrationPolicy, error)
	}

	ProjectPolicyRepo interface {
		SetAggregationPolicy(ctx context.Context, ownerID, projectID string, policy entity.AggregationPolicy) error
		SetRiskPolicy(ctx context.Context, ownerID, projectID string, policy entity.RiskPolicy) error
		SetRecommendationPolicy(ctx context.Context, ownerID, projectID string, policy entity.RecommendationPolicy) error
		SetCalibrationPolicy(ctx context.Context, ownerID, projectID string, policy entity.CalibrationPolicy) error

		GetAggregationPolicy(ctx context.Context, ownerID, projectID string) (entity.AggregationPolicy, bool, error)
		GetRiskPolicy(ctx context.Context, ownerID, projectID string) (entity.RiskPolicy, bool, error)
		GetRecommendationPolicy(ctx context.Context, ownerID, projectID string) (entity.RecommendationPolicy, bool, error)
		GetCalibrationPolicy(ctx context.Context, ownerID, projectID string) (entity.CalibrationPolicy, bool, error)
	}

	ReadModelRepo interface {
		ProjectDashboard(ctx context.Context, ownerID, projectID string, at time.Time) (ProjectDashboardData, error)
		StageBoard(ctx context.Context, ownerID, projectID, stageID string, at time.Time) (StageBoardData, error)
		TaskReadiness(ctx context.Context, ownerID, projectID, stageID, taskID string, at time.Time) (TaskReadinessData, error)
	}

	PolicyDefaultsProvider interface {
		TaskPolicy(ctx context.Context, ownerID, projectID, stageID, taskID string) (entity.TaskPolicy, error)
		AggregationPolicy(ctx context.Context, ownerID, projectID string) (entity.AggregationPolicy, error)
		RiskPolicy(ctx context.Context, ownerID, projectID string) (entity.RiskPolicy, error)
		RecommendationPolicy(ctx context.Context, ownerID, projectID string) (entity.RecommendationPolicy, error)
		CalibrationPolicy(ctx context.Context, ownerID, projectID string) (entity.CalibrationPolicy, error)
	}

	TxManager interface {
		WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
	}
)

type AssessmentSnapshotFilter struct {
	BlockTypes []entity.CompetencyBlockType
	Sources    []entity.AssessmentSource
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

type ProjectDashboardData struct {
	Project            entity.Project
	Stages             []entity.Stage
	CriticalToDeadline []CriticalDeadlineItem
}

type StageBoardData struct {
	Project entity.Project
	Stage   entity.Stage
	Tasks   []entity.Task
}

type TaskReadinessData struct {
	Project         entity.Project
	Stage           entity.Stage
	Task            entity.Task
	LatestSnapshots []entity.AssessmentSnapshot
}

type CriticalDeadlineItem struct {
	StageID    string
	TaskID     string
	Reason     string
	DeadlineAt *time.Time
}
