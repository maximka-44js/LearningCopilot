package policy

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
	projectRepo   repo.ProjectRepo
	stageRepo     repo.StageRepo
	taskRepo      repo.TaskRepo
	userPolicy    repo.UserPolicyRepo
	projectPolicy repo.ProjectPolicyRepo
	tx            repo.TxManager
}

var _ usecase.Policy = (*UseCase)(nil)

func NewPolicyUseCase(
	projectRepo repo.ProjectRepo,
	stageRepo repo.StageRepo,
	taskRepo repo.TaskRepo,
	userPolicy repo.UserPolicyRepo,
	projectPolicy repo.ProjectPolicyRepo,
	tx repo.TxManager,
) *UseCase {
	return &UseCase{
		projectRepo:   projectRepo,
		stageRepo:     stageRepo,
		taskRepo:      taskRepo,
		userPolicy:    userPolicy,
		projectPolicy: projectPolicy,
		tx:            tx,
	}
}

func (uc *UseCase) SetTaskPolicy(
	ctx context.Context,
	ownerID, projectID, stageID, taskID string,
	policy entity.TaskPolicy,
) (entity.Task, error) {
	if err := validatePolicyIDInput(ownerID, projectID, stageID, taskID); err != nil {
		return entity.Task{}, err
	}
	if err := policy.Validate(); err != nil {
		return entity.Task{}, err
	}

	if err := uc.ensureTaskAccess(ctx, ownerID, projectID, stageID, taskID); err != nil {
		return entity.Task{}, err
	}

	task, err := uc.taskRepo.GetByID(ctx, stageID, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	if err := task.UpdatePolicy(policy); err != nil {
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

func (uc *UseCase) SetAggregationPolicy(ctx context.Context, ownerID string, policy entity.AggregationPolicy) error {
	if err := validatePolicyIDInput(ownerID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.userPolicy.SetAggregationPolicy(txCtx, ownerID, policy)
	})
}

func (uc *UseCase) SetRiskPolicy(ctx context.Context, ownerID string, policy entity.RiskPolicy) error {
	if err := validatePolicyIDInput(ownerID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.userPolicy.SetRiskPolicy(txCtx, ownerID, policy)
	})
}

func (uc *UseCase) SetRecommendationPolicy(ctx context.Context, ownerID string, policy entity.RecommendationPolicy) error {
	if err := validatePolicyIDInput(ownerID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.userPolicy.SetRecommendationPolicy(txCtx, ownerID, policy)
	})
}

func (uc *UseCase) SetCalibrationPolicy(ctx context.Context, ownerID string, policy entity.CalibrationPolicy) error {
	if err := validatePolicyIDInput(ownerID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.userPolicy.SetCalibrationPolicy(txCtx, ownerID, policy)
	})
}

func (uc *UseCase) SetProjectAggregationPolicy(ctx context.Context, ownerID, projectID string, policy entity.AggregationPolicy) error {
	if err := validatePolicyIDInput(ownerID, projectID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectPolicy.SetAggregationPolicy(txCtx, ownerID, projectID, policy)
	})
}

func (uc *UseCase) SetProjectRiskPolicy(ctx context.Context, ownerID, projectID string, policy entity.RiskPolicy) error {
	if err := validatePolicyIDInput(ownerID, projectID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectPolicy.SetRiskPolicy(txCtx, ownerID, projectID, policy)
	})
}

func (uc *UseCase) SetProjectRecommendationPolicy(ctx context.Context, ownerID, projectID string, policy entity.RecommendationPolicy) error {
	if err := validatePolicyIDInput(ownerID, projectID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectPolicy.SetRecommendationPolicy(txCtx, ownerID, projectID, policy)
	})
}

func (uc *UseCase) SetProjectCalibrationPolicy(ctx context.Context, ownerID, projectID string, policy entity.CalibrationPolicy) error {
	if err := validatePolicyIDInput(ownerID, projectID); err != nil {
		return err
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	return uc.withinTx(ctx, func(txCtx context.Context) error {
		return uc.projectPolicy.SetCalibrationPolicy(txCtx, ownerID, projectID, policy)
	})
}

func (uc *UseCase) GetUserPolicies(ctx context.Context, ownerID string) (usecase.UserPolicies, error) {
	if err := validatePolicyIDInput(ownerID); err != nil {
		return usecase.UserPolicies{}, err
	}

	aggregation, err := uc.userPolicy.GetAggregationPolicy(ctx, ownerID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}

	risk, err := uc.userPolicy.GetRiskPolicy(ctx, ownerID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}

	recommendation, err := uc.userPolicy.GetRecommendationPolicy(ctx, ownerID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}

	calibration, err := uc.userPolicy.GetCalibrationPolicy(ctx, ownerID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}

	return usecase.UserPolicies{
		Aggregation:    aggregation,
		Risk:           risk,
		Recommendation: recommendation,
		Calibration:    calibration,
	}, nil
}

func (uc *UseCase) GetProjectPolicies(ctx context.Context, ownerID, projectID string) (usecase.UserPolicies, error) {
	if err := validatePolicyIDInput(ownerID, projectID); err != nil {
		return usecase.UserPolicies{}, err
	}
	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return usecase.UserPolicies{}, err
	}

	userPolicies, err := uc.GetUserPolicies(ctx, ownerID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}

	aggregation, aggregationSet, err := uc.projectPolicy.GetAggregationPolicy(ctx, ownerID, projectID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}
	risk, riskSet, err := uc.projectPolicy.GetRiskPolicy(ctx, ownerID, projectID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}
	recommendation, recommendationSet, err := uc.projectPolicy.GetRecommendationPolicy(ctx, ownerID, projectID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}
	calibration, calibrationSet, err := uc.projectPolicy.GetCalibrationPolicy(ctx, ownerID, projectID)
	if err != nil {
		return usecase.UserPolicies{}, err
	}

	if !aggregationSet {
		aggregation = userPolicies.Aggregation
	}
	if !riskSet {
		risk = userPolicies.Risk
	}
	if !recommendationSet {
		recommendation = userPolicies.Recommendation
	}
	if !calibrationSet {
		calibration = userPolicies.Calibration
	}

	return usecase.UserPolicies{
		Aggregation:    aggregation,
		Risk:           risk,
		Recommendation: recommendation,
		Calibration:    calibration,
	}, nil
}

func (uc *UseCase) ensureTaskAccess(ctx context.Context, ownerID, projectID, stageID, taskID string) error {
	if _, err := uc.projectRepo.GetByID(ctx, ownerID, projectID); err != nil {
		return err
	}

	if _, err := uc.stageRepo.GetByID(ctx, projectID, stageID); err != nil {
		return err
	}

	if _, err := uc.taskRepo.GetByID(ctx, stageID, taskID); err != nil {
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

func validatePolicyIDInput(ids ...string) error {
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
