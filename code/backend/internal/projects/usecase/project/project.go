package project

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/internal/projects/repo"
	"github.com/maximka-44js/lc/internal/projects/usecase"
)

type UseCase struct {
	projectRepo repo.ProjectRepo
	stageRepo   repo.StageRepo
	tx          repo.TxManager
	readiness   entity.ReadinessService
}

var _ usecase.Project = (*UseCase)(nil)

func NewProjectUseCase(projectRepo repo.ProjectRepo, stageRepo repo.StageRepo, tx repo.TxManager) *UseCase {
	return &UseCase{
		projectRepo: projectRepo,
		stageRepo:   stageRepo,
		tx:          tx,
		readiness:   entity.NewReadinessService(),
	}
}

func (uc *UseCase) Create(ctx context.Context, in usecase.CreateProjectInput) (entity.Project, error) {
	if err := validateProjectIDInput(in.OwnerID); err != nil {
		return entity.Project{}, err
	}

	project, err := entity.NewProject(
		uuid.NewString(),
		in.OwnerID,
		in.Name,
		in.Description,
		in.DeadlineAt,
		in.Priority,
		in.Metadata,
	)
	if err != nil {
		return entity.Project{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectRepo.Store(txCtx, &project)
	})
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) Update(ctx context.Context, in usecase.UpdateProjectInput) (entity.Project, error) {
	if err := validateProjectIDInput(in.OwnerID, in.ProjectID); err != nil {
		return entity.Project{}, err
	}

	project, err := uc.projectRepo.GetByID(ctx, in.OwnerID, in.ProjectID)
	if err != nil {
		return entity.Project{}, err
	}

	if err := project.UpdateDetails(in.Name, in.Description, in.DeadlineAt, in.Priority, in.Metadata); err != nil {
		return entity.Project{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectRepo.Update(txCtx, &project)
	})
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) GetByID(ctx context.Context, ownerID, projectID string) (entity.Project, error) {
	if err := validateProjectIDInput(ownerID, projectID); err != nil {
		return entity.Project{}, err
	}

	project, err := uc.projectRepo.GetByID(ctx, ownerID, projectID)
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) ListByOwner(ctx context.Context, ownerID string) ([]entity.Project, error) {
	if err := validateProjectIDInput(ownerID); err != nil {
		return nil, err
	}

	return uc.projectRepo.ListByOwner(ctx, ownerID)
}

func (uc *UseCase) Delete(ctx context.Context, ownerID, projectID string) error {
	if err := validateProjectIDInput(ownerID, projectID); err != nil {
		return err
	}

	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectRepo.Delete(txCtx, ownerID, projectID)
	})
}

func (uc *UseCase) Pause(ctx context.Context, ownerID, projectID string) (entity.Project, error) {
	project, err := uc.changeStatus(ctx, ownerID, projectID, (*entity.Project).Pause)
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) Activate(ctx context.Context, ownerID, projectID string) (entity.Project, error) {
	project, err := uc.changeStatus(ctx, ownerID, projectID, (*entity.Project).Activate)
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) Archive(ctx context.Context, ownerID, projectID string) (entity.Project, error) {
	project, err := uc.changeStatus(ctx, ownerID, projectID, (*entity.Project).Archive)
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) RecalculateReadiness(ctx context.Context, in usecase.RecalculateProjectReadinessInput) (entity.Project, error) {
	if err := validateProjectIDInput(in.OwnerID, in.ProjectID); err != nil {
		return entity.Project{}, err
	}

	project, err := uc.projectRepo.GetByID(ctx, in.OwnerID, in.ProjectID)
	if err != nil {
		return entity.Project{}, err
	}
	if project.Status != entity.ProjectStatusActive {
		return project, nil
	}

	stages, err := uc.stageRepo.ListByProject(ctx, in.ProjectID)
	if err != nil {
		return entity.Project{}, err
	}
	activeStages := make([]entity.Stage, 0, len(stages))
	for _, stage := range stages {
		if stage.Status == entity.StageStatusActive {
			activeStages = append(activeStages, stage)
		}
	}

	readiness, err := uc.readiness.RecalculateProjectReadinessAt(project, activeStages, in.AggregationPolicy, in.At)
	if err != nil {
		return entity.Project{}, err
	}

	coverage, err := aggregateStageCoverage(activeStages, in.AggregationPolicy, in.At)
	if err != nil {
		return entity.Project{}, err
	}

	risk := deriveProjectRisk(activeStages)
	if err := project.UpdateMetrics(coverage, float64(readiness), risk); err != nil {
		return entity.Project{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectRepo.Update(txCtx, &project)
	})
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) changeStatus(
	ctx context.Context,
	ownerID, projectID string,
	transition func(*entity.Project) error,
) (entity.Project, error) {
	if err := validateProjectIDInput(ownerID, projectID); err != nil {
		return entity.Project{}, err
	}

	project, err := uc.projectRepo.GetByID(ctx, ownerID, projectID)
	if err != nil {
		return entity.Project{}, err
	}

	if err := transition(&project); err != nil {
		return entity.Project{}, err
	}

	err = uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectRepo.Update(txCtx, &project)
	})
	if err != nil {
		return entity.Project{}, err
	}

	return project, nil
}

func (uc *UseCase) withinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if uc.tx == nil {
		return fn(ctx)
	}
	return uc.tx.WithinTx(ctx, fn)
}

func validateProjectIDInput(ids ...string) error {
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

func aggregateStageCoverage(stages []entity.Stage, policy entity.AggregationPolicy, at time.Time) (float64, error) {
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	if at.IsZero() {
		return 0, d.NewValidationError(entity.CodeReadinessInvalid, "Coverage aggregation time must be set")
	}
	at = at.UTC()

	if len(stages) == 0 {
		return 0, nil
	}

	totalWeight := 0.0
	weightedCoverage := 0.0
	for _, stage := range stages {
		if err := stage.Validate(); err != nil {
			return 0, err
		}
		weight := stageCoverageAggregationWeight(stage, policy, at)
		totalWeight += weight
		weightedCoverage += stage.Coverage * weight
	}

	if totalWeight == 0 {
		return 0, nil
	}

	return weightedCoverage / totalWeight, nil
}

func stageCoverageAggregationWeight(stage entity.Stage, policy entity.AggregationPolicy, at time.Time) float64 {
	if stage.DeadlineAt == nil {
		return 1
	}

	hoursToDeadline := stage.DeadlineAt.Sub(at).Hours()
	switch {
	case hoursToDeadline <= 24:
		return policy.Urgency24hWeight
	case hoursToDeadline <= 72:
		return policy.Urgency72hWeight
	case hoursToDeadline <= 168:
		return policy.Urgency7dWeight
	default:
		return 1
	}
}

func deriveProjectRisk(stages []entity.Stage) entity.RiskLevel {
	worst := entity.RiskLevelMedium
	for _, stage := range stages {
		if riskLevelSeverity(stage.Risk) > riskLevelSeverity(worst) {
			worst = stage.Risk
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
