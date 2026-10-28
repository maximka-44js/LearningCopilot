package usecase

import (
	"context"

	"github.com/maximka-44js/lc/internal/projects/entity"
)

type (
	Project interface {
		Create(ctx context.Context, in CreateProjectInput) (entity.Project, error)
		Update(ctx context.Context, in UpdateProjectInput) (entity.Project, error)
		GetByID(ctx context.Context, ownerID, projectID string) (entity.Project, error)
		ListByOwner(ctx context.Context, ownerID string) ([]entity.Project, error)
		Delete(ctx context.Context, ownerID, projectID string) error
		Pause(ctx context.Context, ownerID, projectID string) (entity.Project, error)
		Activate(ctx context.Context, ownerID, projectID string) (entity.Project, error)
		Archive(ctx context.Context, ownerID, projectID string) (entity.Project, error)
		RecalculateReadiness(ctx context.Context, in RecalculateProjectReadinessInput) (entity.Project, error)
	}

	Stage interface {
		Add(ctx context.Context, in AddStageInput) (entity.Stage, error)
		Update(ctx context.Context, in UpdateStageInput) (entity.Stage, error)
		GetByID(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error)
		ListByProject(ctx context.Context, ownerID, projectID string) ([]entity.Stage, error)
		Delete(ctx context.Context, ownerID, projectID, stageID string) error
		Pause(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error)
		Activate(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error)
		Archive(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error)
		RecalculateReadiness(ctx context.Context, in RecalculateStageReadinessInput) (entity.Stage, error)
	}

	Task interface {
		Add(ctx context.Context, in AddTaskInput) (entity.Task, error)
		Update(ctx context.Context, in UpdateTaskInput) (entity.Task, error)
		GetByID(ctx context.Context, ownerID, projectID, stageID, taskID string) (entity.Task, error)
		GetOutcomeByTaskID(ctx context.Context, ownerID, taskID string) (string, error)
		GetScopeByTaskID(ctx context.Context, ownerID, taskID string) (projectID, stageID string, err error)
		ListByStage(ctx context.Context, ownerID, projectID, stageID string) ([]entity.Task, error)
		Delete(ctx context.Context, ownerID, projectID, stageID, taskID string) error
		ConfigureWeights(ctx context.Context, in ConfigureTaskWeightsInput) (entity.Task, error)
		StartSession(ctx context.Context, in StartTaskSessionInput) (entity.Task, error)
		RecordAssessmentSnapshot(ctx context.Context, in RecordAssessmentSnapshotInput) (RecordAssessmentSnapshotOutput, error)
		RecalculateReadiness(ctx context.Context, in RecalculateTaskReadinessInput) (entity.Task, error)
		GenerateRecommendations(ctx context.Context, in GenerateRecommendationsInput) ([]entity.Recommendation, error)
		MarkExamReady(ctx context.Context, in MarkTaskExamReadyInput) (entity.Task, error)
		TransitionStatus(ctx context.Context, in TransitionTaskStatusInput) (entity.Task, error)
	}

	Policy interface {
		SetTaskPolicy(ctx context.Context, ownerID, projectID, stageID, taskID string, policy entity.TaskPolicy) (entity.Task, error)
		SetAggregationPolicy(ctx context.Context, ownerID string, policy entity.AggregationPolicy) error
		SetRiskPolicy(ctx context.Context, ownerID string, policy entity.RiskPolicy) error
		SetRecommendationPolicy(ctx context.Context, ownerID string, policy entity.RecommendationPolicy) error
		SetCalibrationPolicy(ctx context.Context, ownerID string, policy entity.CalibrationPolicy) error
		SetProjectAggregationPolicy(ctx context.Context, ownerID, projectID string, policy entity.AggregationPolicy) error
		SetProjectRiskPolicy(ctx context.Context, ownerID, projectID string, policy entity.RiskPolicy) error
		SetProjectRecommendationPolicy(ctx context.Context, ownerID, projectID string, policy entity.RecommendationPolicy) error
		SetProjectCalibrationPolicy(ctx context.Context, ownerID, projectID string, policy entity.CalibrationPolicy) error
		GetUserPolicies(ctx context.Context, ownerID string) (UserPolicies, error)
		GetProjectPolicies(ctx context.Context, ownerID, projectID string) (UserPolicies, error)
	}

	ReadModel interface {
		ProjectDashboard(ctx context.Context, ownerID, projectID string) (ProjectDashboardView, error)
		StageBoard(ctx context.Context, ownerID, projectID, stageID string) (StageBoardView, error)
		TaskReadiness(ctx context.Context, ownerID, projectID, stageID, taskID string) (TaskReadinessView, error)
	}
)
