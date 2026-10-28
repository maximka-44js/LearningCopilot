package task

import (
	"context"
	"strings"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/internal/projects/repo"
	"github.com/maximka-44js/lc/internal/projects/usecase"
)

type UseCase struct {
	projectRepo            repo.ProjectRepo
	stageRepo              repo.StageRepo
	taskRepo               repo.TaskRepo
	assessmentSnapshotRepo repo.AssessmentSnapshotRepo
	tx                     repo.TxManager
	readiness              entity.ReadinessService
	risk                   entity.RiskService
	recommendation         entity.RecommendationService
	calibration            entity.CalibrationService
}

var _ usecase.Task = (*UseCase)(nil)

func NewTaskUseCase(
	projectRepo repo.ProjectRepo,
	stageRepo repo.StageRepo,
	taskRepo repo.TaskRepo,
	assessmentSnapshotRepo repo.AssessmentSnapshotRepo,
	tx repo.TxManager,
) *UseCase {
	return &UseCase{
		projectRepo:            projectRepo,
		stageRepo:              stageRepo,
		taskRepo:               taskRepo,
		assessmentSnapshotRepo: assessmentSnapshotRepo,
		tx:                     tx,
		readiness:              entity.NewReadinessService(),
		risk:                   entity.NewRiskService(),
		recommendation:         entity.NewRecommendationService(),
		calibration:            entity.NewCalibrationService(),
	}
}

func (uc *UseCase) Add(ctx context.Context, in usecase.AddTaskInput) (entity.Task, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	task, err := entity.NewTask(
		uuid.NewString(),
		in.StageID,
		in.Outcome,
		in.Difficulty,
		in.Priority,
		in.Policy,
		in.DeadlineAt,
		in.Blocks,
	)
	if err != nil {
		return entity.Task{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Store(txCtx, &task)
	})
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) Update(ctx context.Context, in usecase.UpdateTaskInput) (entity.Task, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return entity.Task{}, err
	}

	if err := task.UpdateDetails(in.Outcome, in.Difficulty, in.Priority, in.DeadlineAt); err != nil {
		return entity.Task{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Update(txCtx, &task)
	})
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) GetByID(ctx context.Context, ownerID, projectID, stageID, taskID string) (entity.Task, error) {
	if err := validateTaskIDInput(ownerID, projectID, stageID, taskID); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, ownerID, projectID, stageID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, stageID, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) GetOutcomeByTaskID(ctx context.Context, ownerID, taskID string) (string, error) {
	if strings.TrimSpace(ownerID) == "" {
		return "", d.NewValidationError(entity.CodeInvalidID, "Owner id is required")
	}
	if strings.TrimSpace(taskID) == "" {
		return "", d.NewValidationError(entity.CodeInvalidID, "Task id is required")
	}
	item, err := uc.taskRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(item.Outcome), nil
}

func (uc *UseCase) GetScopeByTaskID(ctx context.Context, ownerID, taskID string) (string, string, error) {
	if strings.TrimSpace(ownerID) == "" {
		return "", "", d.NewValidationError(entity.CodeInvalidID, "Owner id is required")
	}
	if strings.TrimSpace(taskID) == "" {
		return "", "", d.NewValidationError(entity.CodeInvalidID, "Task id is required")
	}
	if _, err := uuid.Parse(strings.TrimSpace(ownerID)); err != nil {
		return "", "", d.NewValidationError(entity.CodeInvalidID, "UUID is invalid")
	}
	if _, err := uuid.Parse(strings.TrimSpace(taskID)); err != nil {
		return "", "", d.NewValidationError(entity.CodeInvalidID, "UUID is invalid")
	}

	projectID, stageID, err := uc.taskRepo.GetScopeByTaskID(ctx, taskID)
	if err != nil {
		return "", "", err
	}
	if err := uc.ensureStageAccess(ctx, ownerID, projectID, stageID); err != nil {
		return "", "", err
	}

	return projectID, stageID, nil
}

func (uc *UseCase) ListByStage(ctx context.Context, ownerID, projectID, stageID string) ([]entity.Task, error) {
	if err := validateTaskIDInput(ownerID, projectID, stageID); err != nil {
		return nil, err
	}

	if err := uc.ensureStageAccess(ctx, ownerID, projectID, stageID); err != nil {
		return nil, err
	}

	return uc.taskRepo.ListByStage(ctx, stageID)
}

func (uc *UseCase) Delete(ctx context.Context, ownerID, projectID, stageID, taskID string) error {
	if err := validateTaskIDInput(ownerID, projectID, stageID, taskID); err != nil {
		return err
	}

	if err := uc.ensureStageAccess(ctx, ownerID, projectID, stageID); err != nil {
		return err
	}

	if _, err := uc.taskRepo.GetByID(ctx, stageID, taskID); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Delete(txCtx, stageID, taskID)
	})
}

func (uc *UseCase) ConfigureWeights(ctx context.Context, in usecase.ConfigureTaskWeightsInput) (entity.Task, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return entity.Task{}, err
	}
	if err := in.Weights.Validate(); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return entity.Task{}, err
	}

	updatedBlocks, err := applyWeightDistribution(task.Blocks, in.Weights)
	if err != nil {
		return entity.Task{}, err
	}

	if err := task.ConfigureBlocks(updatedBlocks); err != nil {
		return entity.Task{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Update(txCtx, &task)
	})
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) StartSession(ctx context.Context, in usecase.StartTaskSessionInput) (entity.Task, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return entity.Task{}, err
	}

	if err := task.TransitionStatus(entity.TaskStatusLearning); err != nil {
		return entity.Task{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Update(txCtx, &task)
	})
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) RecordAssessmentSnapshot(
	ctx context.Context,
	in usecase.RecordAssessmentSnapshotInput,
) (usecase.RecordAssessmentSnapshotOutput, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return usecase.RecordAssessmentSnapshotOutput{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return usecase.RecordAssessmentSnapshotOutput{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return usecase.RecordAssessmentSnapshotOutput{}, err
	}

	snapshot, err := entity.NewAssessmentSnapshot(
		uuid.NewString(),
		in.TaskID,
		in.BlockType,
		in.Source,
		in.RawScore,
		in.NormalizedScore,
		in.ConfidenceBefore,
		in.ConfidenceAfter,
		in.RecordedAt,
		in.TraceID,
		in.SourceReference,
		in.SourceReferenceType,
	)
	if err != nil {
		return usecase.RecordAssessmentSnapshotOutput{}, err
	}

	block, err := findBlock(task.Blocks, in.BlockType)
	if err != nil {
		return usecase.RecordAssessmentSnapshotOutput{}, err
	}
	blockRisk := uc.risk.CalculateBlockRiskWithPolicy(block, in.NormalizedScore, entity.DefaultRiskPolicy())

	if err := task.RecordAssessmentSnapshot(in.BlockType, in.NormalizedScore, in.RecordedAt, blockRisk); err != nil {
		return usecase.RecordAssessmentSnapshotOutput{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		if err := uc.assessmentSnapshotRepo.Store(txCtx, &snapshot); err != nil {
			return err
		}
		if err := uc.taskRepo.Update(txCtx, &task); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return usecase.RecordAssessmentSnapshotOutput{}, err
	}

	return usecase.RecordAssessmentSnapshotOutput{Task: task, Snapshot: snapshot}, nil
}

func (uc *UseCase) RecalculateReadiness(ctx context.Context, in usecase.RecalculateTaskReadinessInput) (entity.Task, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return entity.Task{}, err
	}
	if err := in.RiskPolicy.Validate(); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return entity.Task{}, err
	}
	if task.Status == entity.TaskStatusArchived {
		return task, nil
	}

	for i := range task.Blocks {
		profile, ok := in.Profiles[task.Blocks[i].Type]
		if !ok {
			return entity.Task{}, d.NewValidationError(entity.CodeDecayProfileInvalid, "Decay profile is missing for competency block type")
		}
		if err := profile.Validate(); err != nil {
			return entity.Task{}, err
		}

		nowReadiness := 0.0
		if task.Blocks[i].LastAssessedAt != nil {
			freshness := profile.Freshness(*task.Blocks[i].LastAssessedAt, in.At)
			nowReadiness = task.Blocks[i].Mastery * freshness
		}
		task.Blocks[i].RiskState = uc.risk.CalculateBlockRiskWithPolicy(task.Blocks[i], nowReadiness, in.RiskPolicy)
	}

	readiness, err := uc.readiness.RecalculateTaskReadiness(task, in.Profiles, in.At)
	if err != nil {
		return entity.Task{}, err
	}

	coverage, err := uc.readiness.RecalculateTaskCoverage(task)
	if err != nil {
		return entity.Task{}, err
	}

	taskRisk := taskRiskFromReadiness(float64(readiness), in.RiskPolicy)
	if err := task.UpdateMetrics(coverage, float64(readiness), taskRisk); err != nil {
		return entity.Task{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Update(txCtx, &task)
	})
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) GenerateRecommendations(
	ctx context.Context,
	in usecase.GenerateRecommendationsInput,
) ([]entity.Recommendation, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return nil, err
	}
	if err := in.RecommendationPolicy.Validate(); err != nil {
		return nil, err
	}
	if err := in.CalibrationPolicy.Validate(); err != nil {
		return nil, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return nil, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return nil, err
	}

	latestSnapshot, err := uc.assessmentSnapshotRepo.GetLatestByTask(ctx, in.TaskID)
	if err != nil {
		return nil, err
	}

	policy := in.RecommendationPolicy
	policy.CalibrationGapThreshold = in.CalibrationPolicy.GapThreshold

	if latestSnapshot != nil {
		if _, err := uc.calibration.NeedsCalibrationWithPolicy(*latestSnapshot, in.CalibrationPolicy); err != nil {
			return nil, err
		}
	}

	return uc.recommendation.GenerateForTaskWithPolicy(task, latestSnapshot, policy)
}

func (uc *UseCase) MarkExamReady(ctx context.Context, in usecase.MarkTaskExamReadyInput) (entity.Task, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return entity.Task{}, err
	}

	if err := task.TransitionStatus(entity.TaskStatusExamReady); err != nil {
		return entity.Task{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Update(txCtx, &task)
	})
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) TransitionStatus(ctx context.Context, in usecase.TransitionTaskStatusInput) (entity.Task, error) {
	if err := validateTaskIDInput(in.OwnerID, in.ProjectID, in.StageID, in.TaskID); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureStageAccess(ctx, in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, in.StageID, in.TaskID)
	if err != nil {
		return entity.Task{}, err
	}

	if err := task.TransitionStatus(in.NextStatus); err != nil {
		return entity.Task{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.taskRepo.Update(txCtx, &task)
	})
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (uc *UseCase) ensureStageAccess(ctx context.Context, ownerID, projectID, stageID string) error {
	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	if _, err := uc.stageRepo.GetByID(ctx, projectID, stageID); err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) withinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if uc.tx == nil {
		return fn(ctx)
	}
	return uc.tx.WithinTx(ctx, fn)
}

func validateTaskIDInput(ids ...string) error {
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			return d.NewValidationError(entity.CodeInvalidID, "UUID is required")
		}
		if _, err := uuid.Parse(trimmed); err != nil {
			return d.NewValidationError(entity.CodeInvalidID, "UUID is invalid")
		}
	}
	return nil
}

func applyWeightDistribution(
	blocks []entity.CompetencyBlock,
	weights entity.WeightDistribution,
) ([]entity.CompetencyBlock, error) {
	updated := make([]entity.CompetencyBlock, 0, len(blocks))
	for _, block := range blocks {
		weight, ok := weights[block.Type]
		if !ok {
			return nil, d.NewValidationError(entity.CodeWeightDistributionInvalid, "Weight distribution must include all task block types")
		}
		copyBlock := block
		copyBlock.Weight = weight
		updated = append(updated, copyBlock)
	}

	if len(updated) != len(weights) {
		return nil, d.NewValidationError(entity.CodeWeightDistributionInvalid, "Weight distribution contains unsupported task block types")
	}

	return updated, nil
}

func findBlock(blocks []entity.CompetencyBlock, blockType entity.CompetencyBlockType) (entity.CompetencyBlock, error) {
	for _, block := range blocks {
		if block.Type == blockType {
			return block, nil
		}
	}
	return entity.CompetencyBlock{}, d.NewValidationError(entity.CodeTaskBlocksInvalid, "Task does not contain requested competency block")
}

func taskRiskFromReadiness(readiness float64, policy entity.RiskPolicy) entity.RiskLevel {
	switch {
	case readiness < policy.TaskCriticalReadinessBelow:
		return entity.RiskLevelCritical
	case readiness < policy.TaskHighReadinessBelow:
		return entity.RiskLevelHigh
	case readiness < policy.TaskMediumReadinessBelow:
		return entity.RiskLevelMedium
	default:
		return entity.RiskLevelLow
	}
}
