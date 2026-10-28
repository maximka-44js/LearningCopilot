package stage

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
	projectRepo repo.ProjectRepo
	stageRepo   repo.StageRepo
	taskRepo    repo.TaskRepo
	tx          repo.TxManager
	readiness   entity.ReadinessService
}

var _ usecase.Stage = (*UseCase)(nil)

func NewStageUseCase(projectRepo repo.ProjectRepo, stageRepo repo.StageRepo, taskRepo repo.TaskRepo, tx repo.TxManager) *UseCase {
	return &UseCase{
		projectRepo: projectRepo,
		stageRepo:   stageRepo,
		taskRepo:    taskRepo,
		tx:          tx,
		readiness:   entity.NewReadinessService(),
	}
}

func (uc *UseCase) Add(ctx context.Context, in usecase.AddStageInput) (entity.Stage, error) {
	if err := validateStageIDInput(in.OwnerID, in.ProjectID); err != nil {
		return entity.Stage{}, err
	}

	if _, err := uc.projectRepo.GetByID(ctx, in.OwnerID, in.ProjectID); err != nil {
		return entity.Stage{}, err
	}

	stage, err := entity.NewStage(
		uuid.NewString(),
		in.ProjectID,
		in.Name,
		in.Goal,
		in.Order,
		in.DeadlineAt,
	)
	if err != nil {
		return entity.Stage{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.stageRepo.Store(txCtx, &stage)
	})
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) Update(ctx context.Context, in usecase.UpdateStageInput) (entity.Stage, error) {
	if err := validateStageIDInput(in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Stage{}, err
	}

	if _, err := uc.projectRepo.GetByID(ctx, in.OwnerID, in.ProjectID); err != nil {
		return entity.Stage{}, err
	}

	stage, err := uc.stageRepo.GetByID(ctx, in.ProjectID, in.StageID)
	if err != nil {
		return entity.Stage{}, err
	}

	if err := stage.UpdateDetails(in.Name, in.Goal, in.Order, in.DeadlineAt); err != nil {
		return entity.Stage{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.stageRepo.Update(txCtx, &stage)
	})
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) GetByID(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error) {
	if err := validateStageIDInput(ownerID, projectID, stageID); err != nil {
		return entity.Stage{}, err
	}

	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return entity.Stage{}, err
	}

	stage, err := uc.stageRepo.GetByID(ctx, projectID, stageID)
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) ListByProject(ctx context.Context, ownerID, projectID string) ([]entity.Stage, error) {
	if err := validateStageIDInput(ownerID, projectID); err != nil {
		return nil, err
	}

	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return nil, err
	}

	return uc.stageRepo.ListByProject(ctx, projectID)
}

func (uc *UseCase) Delete(ctx context.Context, ownerID, projectID, stageID string) error {
	if err := validateStageIDInput(ownerID, projectID, stageID); err != nil {
		return err
	}

	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	if _, err := uc.stageRepo.GetByID(ctx, projectID, stageID); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.stageRepo.Delete(txCtx, projectID, stageID)
	})
}

func (uc *UseCase) Pause(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error) {
	stage, err := uc.changeStatus(ctx, ownerID, projectID, stageID, (*entity.Stage).Pause)
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) Activate(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error) {
	stage, err := uc.changeStatus(ctx, ownerID, projectID, stageID, (*entity.Stage).Activate)
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) Archive(ctx context.Context, ownerID, projectID, stageID string) (entity.Stage, error) {
	stage, err := uc.changeStatus(ctx, ownerID, projectID, stageID, (*entity.Stage).Archive)
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) RecalculateReadiness(ctx context.Context, in usecase.RecalculateStageReadinessInput) (entity.Stage, error) {
	if err := validateStageIDInput(in.OwnerID, in.ProjectID, in.StageID); err != nil {
		return entity.Stage{}, err
	}

	if _, err := uc.projectRepo.GetByID(ctx, in.OwnerID, in.ProjectID); err != nil {
		return entity.Stage{}, err
	}

	stage, err := uc.stageRepo.GetByID(ctx, in.ProjectID, in.StageID)
	if err != nil {
		return entity.Stage{}, err
	}
	if stage.Status != entity.StageStatusActive {
		return stage, nil
	}

	tasks, err := uc.taskRepo.ListByStage(ctx, in.StageID)
	if err != nil {
		return entity.Stage{}, err
	}
	activeTasks := make([]entity.Task, 0, len(tasks))
	for _, task := range tasks {
		if task.Status != entity.TaskStatusArchived {
			activeTasks = append(activeTasks, task)
		}
	}

	readiness, err := uc.readiness.RecalculateStageReadinessAt(stage, activeTasks, in.AggregationPolicy, in.At)
	if err != nil {
		return entity.Stage{}, err
	}

	coverage, err := uc.readiness.RecalculateCoverageFromTasksAt(activeTasks, in.AggregationPolicy, in.At)
	if err != nil {
		return entity.Stage{}, err
	}

	risk := deriveStageRisk(activeTasks)
	if err := stage.UpdateMetrics(coverage, float64(readiness), risk); err != nil {
		return entity.Stage{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.stageRepo.Update(txCtx, &stage)
	})
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) changeStatus(
	ctx context.Context,
	ownerID, projectID, stageID string,
	transition func(*entity.Stage) error,
) (entity.Stage, error) {
	if err := validateStageIDInput(ownerID, projectID, stageID); err != nil {
		return entity.Stage{}, err
	}

	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return entity.Stage{}, err
	}

	stage, err := uc.stageRepo.GetByID(ctx, projectID, stageID)
	if err != nil {
		return entity.Stage{}, err
	}

	if err := transition(&stage); err != nil {
		return entity.Stage{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.stageRepo.Update(txCtx, &stage)
	})
	if err != nil {
		return entity.Stage{}, err
	}

	return stage, nil
}

func (uc *UseCase) withinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if uc.tx == nil {
		return fn(ctx)
	}
	return uc.tx.WithinTx(ctx, fn)
}

func validateStageIDInput(ids ...string) error {
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

func deriveStageRisk(tasks []entity.Task) entity.RiskLevel {
	worst := entity.RiskLevelMedium
	for _, task := range tasks {
		if riskLevelSeverity(task.Risk) > riskLevelSeverity(worst) {
			worst = task.Risk
		}
	}
	return worst
}

func riskLevelSeverity(level entity.RiskLevel) int {
	switch level {
	case entity.RiskLevelLow:
		return 1
	case entity.RiskLevelMedium:
		return 2
	case entity.RiskLevelHigh:
		return 3
	case entity.RiskLevelCritical:
		return 4
	default:
		return 0
	}
}
