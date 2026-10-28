package persistent

import (
	"context"
	"encoding/json"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/pkg/postgres"
)

const (
	policyColumnAggregation    = "aggregation_policy"
	policyColumnRisk           = "risk_policy"
	policyColumnRecommendation = "recommendation_policy"
	policyColumnCalibration    = "calibration_policy"
)

type UserPolicyRepo struct {
	*postgres.Postgres
}

func NewUserPolicyRepo(pg *postgres.Postgres) *UserPolicyRepo {
	return &UserPolicyRepo{pg}
}

func (r *UserPolicyRepo) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeUndefined, "User policy not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return d.NewConflict(entity.CodeUndefined, "User policy conflict")
		case "23503", "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeUndefined, "User policy data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func (r *UserPolicyRepo) SetAggregationPolicy(ctx context.Context, ownerID string, policy entity.AggregationPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, policyColumnAggregation, policy)
}

func (r *UserPolicyRepo) SetRiskPolicy(ctx context.Context, ownerID string, policy entity.RiskPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, policyColumnRisk, policy)
}

func (r *UserPolicyRepo) SetRecommendationPolicy(ctx context.Context, ownerID string, policy entity.RecommendationPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, policyColumnRecommendation, policy)
}

func (r *UserPolicyRepo) SetCalibrationPolicy(ctx context.Context, ownerID string, policy entity.CalibrationPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, policyColumnCalibration, policy)
}

func (r *UserPolicyRepo) GetAggregationPolicy(ctx context.Context, ownerID string) (entity.AggregationPolicy, error) {
	defaultPolicy := entity.DefaultAggregationPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, policyColumnAggregation)
	if err != nil {
		return entity.AggregationPolicy{}, err
	}
	if !found {
		return defaultPolicy, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.AggregationPolicy{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, nil
	}

	return value, nil
}

func (r *UserPolicyRepo) GetRiskPolicy(ctx context.Context, ownerID string) (entity.RiskPolicy, error) {
	defaultPolicy := entity.DefaultRiskPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, policyColumnRisk)
	if err != nil {
		return entity.RiskPolicy{}, err
	}
	if !found {
		return defaultPolicy, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.RiskPolicy{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, nil
	}

	return value, nil
}

func (r *UserPolicyRepo) GetRecommendationPolicy(ctx context.Context, ownerID string) (entity.RecommendationPolicy, error) {
	defaultPolicy := entity.DefaultRecommendationPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, policyColumnRecommendation)
	if err != nil {
		return entity.RecommendationPolicy{}, err
	}
	if !found {
		return defaultPolicy, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.RecommendationPolicy{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, nil
	}

	return value, nil
}

func (r *UserPolicyRepo) GetCalibrationPolicy(ctx context.Context, ownerID string) (entity.CalibrationPolicy, error) {
	defaultPolicy := entity.DefaultCalibrationPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, policyColumnCalibration)
	if err != nil {
		return entity.CalibrationPolicy{}, err
	}
	if !found {
		return defaultPolicy, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.CalibrationPolicy{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, nil
	}

	return value, nil
}

func (r *UserPolicyRepo) TaskPolicy(
	ctx context.Context,
	ownerID, projectID, stageID, taskID string,
) (entity.TaskPolicy, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Select("t.exam_ready_min_mastery", "t.require_non_critical_block_risk").
		From("tasks t").
		Join("stages s ON s.id = t.stage_id").
		Join("projects p ON p.id = s.project_id").
		Where(sq.Eq{"p.owner_id": ownerID}).
		Where(sq.Eq{"p.id": projectID}).
		Where(sq.Eq{"s.id": stageID}).
		Where(sq.Eq{"t.id": taskID}).
		Limit(1).
		ToSql()
	if err != nil {
		return entity.TaskPolicy{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var policy entity.TaskPolicy
	if err := exec.QueryRow(ctx, sql, args...).Scan(
		&policy.ExamReadyMinMastery,
		&policy.RequireNonCriticalBlockRisk,
	); err != nil {
		return entity.TaskPolicy{}, r.handleErr(err)
	}

	if err := policy.Validate(); err != nil {
		return entity.TaskPolicy{}, err
	}

	return policy, nil
}

func (r *UserPolicyRepo) AggregationPolicy(ctx context.Context, ownerID string) (entity.AggregationPolicy, error) {
	return r.GetAggregationPolicy(ctx, ownerID)
}

func (r *UserPolicyRepo) RiskPolicy(ctx context.Context, ownerID string) (entity.RiskPolicy, error) {
	return r.GetRiskPolicy(ctx, ownerID)
}

func (r *UserPolicyRepo) RecommendationPolicy(ctx context.Context, ownerID string) (entity.RecommendationPolicy, error) {
	return r.GetRecommendationPolicy(ctx, ownerID)
}

func (r *UserPolicyRepo) CalibrationPolicy(ctx context.Context, ownerID string) (entity.CalibrationPolicy, error) {
	return r.GetCalibrationPolicy(ctx, ownerID)
}

func (r *UserPolicyRepo) setPolicy(ctx context.Context, ownerID, column string, payload any) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	if !isAllowedPolicyColumn(column) {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	policyJSON, err := json.Marshal(payload)
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	sql, args, err := r.Builder.
		Insert("user_project_policies").
		Columns("owner_id", column).
		Values(ownerID, sq.Expr("?::jsonb", string(policyJSON))).
		Suffix("ON CONFLICT (owner_id) DO UPDATE SET " + column + " = EXCLUDED." + column + ", updated_at = now()").
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, sql, args...); err != nil {
		return r.handleErr(err)
	}

	return nil
}

func (r *UserPolicyRepo) getPolicyRaw(ctx context.Context, ownerID, column string) ([]byte, bool, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	if !isAllowedPolicyColumn(column) {
		return nil, false, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	sql, args, err := r.Builder.
		Select(column).
		From("user_project_policies").
		Where(sq.Eq{"owner_id": ownerID}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, false, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var raw []byte
	if err := exec.QueryRow(ctx, sql, args...).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, r.handleErr(err)
	}

	if len(raw) == 0 || string(raw) == "null" {
		return nil, false, nil
	}

	return raw, true, nil
}

func isAllowedPolicyColumn(column string) bool {
	switch column {
	case policyColumnAggregation, policyColumnRisk, policyColumnRecommendation, policyColumnCalibration:
		return true
	default:
		return false
	}
}
