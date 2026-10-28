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
	projectPolicyColumnAggregation    = "aggregation_policy"
	projectPolicyColumnRisk           = "risk_policy"
	projectPolicyColumnRecommendation = "recommendation_policy"
	projectPolicyColumnCalibration    = "calibration_policy"
)

type ProjectPolicyRepo struct {
	*postgres.Postgres
}

func NewProjectPolicyRepo(pg *postgres.Postgres) *ProjectPolicyRepo {
	return &ProjectPolicyRepo{pg}
}

func (r *ProjectPolicyRepo) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeUndefined, "Project policy not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return d.NewConflict(entity.CodeUndefined, "Project policy conflict")
		case "23503", "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeUndefined, "Project policy data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func (r *ProjectPolicyRepo) SetAggregationPolicy(ctx context.Context, ownerID, projectID string, policy entity.AggregationPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, projectID, projectPolicyColumnAggregation, policy)
}

func (r *ProjectPolicyRepo) SetRiskPolicy(ctx context.Context, ownerID, projectID string, policy entity.RiskPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, projectID, projectPolicyColumnRisk, policy)
}

func (r *ProjectPolicyRepo) SetRecommendationPolicy(ctx context.Context, ownerID, projectID string, policy entity.RecommendationPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, projectID, projectPolicyColumnRecommendation, policy)
}

func (r *ProjectPolicyRepo) SetCalibrationPolicy(ctx context.Context, ownerID, projectID string, policy entity.CalibrationPolicy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	return r.setPolicy(ctx, ownerID, projectID, projectPolicyColumnCalibration, policy)
}

func (r *ProjectPolicyRepo) GetAggregationPolicy(ctx context.Context, ownerID, projectID string) (entity.AggregationPolicy, bool, error) {
	defaultPolicy := entity.DefaultAggregationPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, projectID, projectPolicyColumnAggregation)
	if err != nil {
		return entity.AggregationPolicy{}, false, err
	}
	if !found {
		return defaultPolicy, false, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.AggregationPolicy{}, false, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, false, nil
	}

	return value, true, nil
}

func (r *ProjectPolicyRepo) GetRiskPolicy(ctx context.Context, ownerID, projectID string) (entity.RiskPolicy, bool, error) {
	defaultPolicy := entity.DefaultRiskPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, projectID, projectPolicyColumnRisk)
	if err != nil {
		return entity.RiskPolicy{}, false, err
	}
	if !found {
		return defaultPolicy, false, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.RiskPolicy{}, false, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, false, nil
	}

	return value, true, nil
}

func (r *ProjectPolicyRepo) GetRecommendationPolicy(ctx context.Context, ownerID, projectID string) (entity.RecommendationPolicy, bool, error) {
	defaultPolicy := entity.DefaultRecommendationPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, projectID, projectPolicyColumnRecommendation)
	if err != nil {
		return entity.RecommendationPolicy{}, false, err
	}
	if !found {
		return defaultPolicy, false, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.RecommendationPolicy{}, false, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, false, nil
	}

	return value, true, nil
}

func (r *ProjectPolicyRepo) GetCalibrationPolicy(ctx context.Context, ownerID, projectID string) (entity.CalibrationPolicy, bool, error) {
	defaultPolicy := entity.DefaultCalibrationPolicy()
	raw, found, err := r.getPolicyRaw(ctx, ownerID, projectID, projectPolicyColumnCalibration)
	if err != nil {
		return entity.CalibrationPolicy{}, false, err
	}
	if !found {
		return defaultPolicy, false, nil
	}

	value := defaultPolicy
	if err := json.Unmarshal(raw, &value); err != nil {
		return entity.CalibrationPolicy{}, false, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	if err := value.Validate(); err != nil {
		return defaultPolicy, false, nil
	}

	return value, true, nil
}

func (r *ProjectPolicyRepo) setPolicy(ctx context.Context, ownerID, projectID, column string, payload any) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	if !isAllowedProjectPolicyColumn(column) {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	policyJSON, err := json.Marshal(payload)
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	sql, args, err := r.Builder.
		Insert("project_policies").
		Columns("project_id", "owner_id", column).
		Values(projectID, ownerID, sq.Expr("?::jsonb", string(policyJSON))).
		Suffix("ON CONFLICT (project_id) DO UPDATE SET " + column + " = EXCLUDED." + column + ", owner_id = EXCLUDED.owner_id, updated_at = now()").
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, sql, args...); err != nil {
		return r.handleErr(err)
	}

	return nil
}

func (r *ProjectPolicyRepo) getPolicyRaw(ctx context.Context, ownerID, projectID, column string) ([]byte, bool, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	if !isAllowedProjectPolicyColumn(column) {
		return nil, false, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	sql, args, err := r.Builder.
		Select(column).
		From("project_policies").
		Where(sq.Eq{"project_id": projectID}).
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

	return raw, true, nil
}

func isAllowedProjectPolicyColumn(column string) bool {
	switch column {
	case projectPolicyColumnAggregation, projectPolicyColumnRisk, projectPolicyColumnRecommendation, projectPolicyColumnCalibration:
		return true
	default:
		return false
	}
}
