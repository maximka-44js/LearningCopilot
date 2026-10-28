package persistent

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/entity"
	"github.com/maximka-44js/lc/pkg/postgres"
)

type StageRepo struct {
	*postgres.Postgres
}

func NewStageRepo(pg *postgres.Postgres) *StageRepo {
	return &StageRepo{pg}
}

func (r *StageRepo) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeStageNotFound, "Stage not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return d.NewConflict(entity.CodeStageOrderInvalid, "Stage with this order already exists")
		case "23503":
			return d.NewValidationError(entity.CodeStageInvalid, "Referenced resource not found")
		case "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeStageInvalid, "Stage data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func (r *StageRepo) Store(ctx context.Context, stage *entity.Stage) error {
	if stage == nil {
		return d.NewValidationError(entity.CodeStageInvalid, "Stage payload is required")
	}
	if err := stage.Validate(); err != nil {
		return err
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	if err := r.lockProject(ctx, exec, stage.ProjectID); err != nil {
		return err
	}

	maxOrder, err := r.maxStageOrder(ctx, exec, stage.ProjectID)
	if err != nil {
		return err
	}

	order := normalizeInsertStageOrder(stage.Order, maxOrder)
	if order <= maxOrder {
		if err := r.shiftStageOrders(ctx, exec, stage.ProjectID, order, maxOrder, 1); err != nil {
			return err
		}
	}
	stage.Order = order

	sql, args, err := r.Builder.
		Insert("stages").
		Columns(
			"id",
			"project_id",
			"name",
			"goal",
			"stage_order",
			"status",
			"deadline_at",
			"coverage",
			"readiness",
			"risk",
			"created_at",
			"updated_at",
			"archived_at",
		).
		Values(
			stage.ID,
			stage.ProjectID,
			stage.Name,
			stage.Goal,
			stage.Order,
			stage.Status,
			stage.DeadlineAt,
			stage.Coverage,
			stage.Readiness,
			stage.Risk,
			stage.CreatedAt,
			stage.UpdatedAt,
			stage.ArchivedAt,
		).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, sql, args...); err != nil {
		return r.handleErr(err)
	}

	return nil
}

func (r *StageRepo) GetByID(ctx context.Context, projectID, stageID string) (entity.Stage, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Select(
			"id",
			"project_id",
			"name",
			"goal",
			"stage_order",
			"status",
			"deadline_at",
			"coverage",
			"readiness",
			"risk",
			"created_at",
			"updated_at",
			"archived_at",
		).
		From("stages").
		Where(sq.Eq{"id": stageID}).
		Where(sq.Eq{"project_id": projectID}).
		ToSql()
	if err != nil {
		return entity.Stage{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var stage entity.Stage

	row := exec.QueryRow(ctx, sql, args...)
	if err := row.Scan(
		&stage.ID,
		&stage.ProjectID,
		&stage.Name,
		&stage.Goal,
		&stage.Order,
		&stage.Status,
		&stage.DeadlineAt,
		&stage.Coverage,
		&stage.Readiness,
		&stage.Risk,
		&stage.CreatedAt,
		&stage.UpdatedAt,
		&stage.ArchivedAt,
	); err != nil {
		return entity.Stage{}, r.handleErr(err)
	}

	return stage, nil
}

func (r *StageRepo) ListByProject(ctx context.Context, projectID string) ([]entity.Stage, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Select(
			"id",
			"project_id",
			"name",
			"goal",
			"stage_order",
			"status",
			"deadline_at",
			"coverage",
			"readiness",
			"risk",
			"created_at",
			"updated_at",
			"archived_at",
		).
		From("stages").
		Where(sq.Eq{"project_id": projectID}).
		OrderBy("stage_order ASC", "created_at ASC").
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	stages := make([]entity.Stage, 0)
	for rows.Next() {
		var stage entity.Stage

		if err := rows.Scan(
			&stage.ID,
			&stage.ProjectID,
			&stage.Name,
			&stage.Goal,
			&stage.Order,
			&stage.Status,
			&stage.DeadlineAt,
			&stage.Coverage,
			&stage.Readiness,
			&stage.Risk,
			&stage.CreatedAt,
			&stage.UpdatedAt,
			&stage.ArchivedAt,
		); err != nil {
			return nil, r.handleErr(err)
		}

		stages = append(stages, stage)
	}

	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return stages, nil
}

func (r *StageRepo) Update(ctx context.Context, stage *entity.Stage) error {
	if stage == nil {
		return d.NewValidationError(entity.CodeStageInvalid, "Stage payload is required")
	}
	if err := stage.Validate(); err != nil {
		return err
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	if err := r.lockProject(ctx, exec, stage.ProjectID); err != nil {
		return err
	}

	var currentOrder int
	row := exec.QueryRow(ctx, "SELECT stage_order FROM stages WHERE id = $1 AND project_id = $2 FOR UPDATE", stage.ID, stage.ProjectID)
	if err := row.Scan(&currentOrder); err != nil {
		return r.handleErr(err)
	}

	maxOrder, err := r.maxStageOrder(ctx, exec, stage.ProjectID)
	if err != nil {
		return err
	}

	targetOrder := normalizeUpdateStageOrder(stage.Order, maxOrder)
	if err := r.reorderStageOrders(ctx, exec, stage.ProjectID, currentOrder, targetOrder); err != nil {
		return err
	}
	stage.Order = targetOrder

	sql, args, err := r.Builder.
		Update("stages").
		Set("name", stage.Name).
		Set("goal", stage.Goal).
		Set("stage_order", stage.Order).
		Set("status", stage.Status).
		Set("deadline_at", stage.DeadlineAt).
		Set("coverage", stage.Coverage).
		Set("readiness", stage.Readiness).
		Set("risk", stage.Risk).
		Set("updated_at", stage.UpdatedAt).
		Set("archived_at", stage.ArchivedAt).
		Where(sq.Eq{"id": stage.ID}).
		Where(sq.Eq{"project_id": stage.ProjectID}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewNotFound(entity.CodeStageNotFound, "Stage not found")
	}

	return nil
}

func (r *StageRepo) lockProject(ctx context.Context, exec postgres.Executor, projectID string) error {
	var locked int
	if err := exec.QueryRow(ctx, "SELECT 1 FROM projects WHERE id = $1 FOR UPDATE", projectID).Scan(&locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.NewNotFound(entity.CodeProjectNotFound, "Project not found")
		}
		return r.handleErr(err)
	}

	return nil
}

func (r *StageRepo) maxStageOrder(ctx context.Context, exec postgres.Executor, projectID string) (int, error) {
	var maxOrder int
	if err := exec.QueryRow(ctx, "SELECT COALESCE(MAX(stage_order), 0) FROM stages WHERE project_id = $1", projectID).Scan(&maxOrder); err != nil {
		return 0, r.handleErr(err)
	}

	return maxOrder, nil
}

func (r *StageRepo) reorderStageOrders(ctx context.Context, exec postgres.Executor, projectID string, currentOrder, targetOrder int) error {
	if currentOrder == targetOrder {
		return nil
	}

	fromOrder, toOrder, delta, ok := planStageOrderShift(currentOrder, targetOrder)
	if !ok {
		return nil
	}

	return r.shiftStageOrders(ctx, exec, projectID, fromOrder, toOrder, delta)
}

func (r *StageRepo) shiftStageOrders(ctx context.Context, exec postgres.Executor, projectID string, fromOrder, toOrder, delta int) error {
	if fromOrder > toOrder {
		return nil
	}

	maxOrder, err := r.maxStageOrder(ctx, exec, projectID)
	if err != nil {
		return err
	}
	temporaryOffset := maxOrder + 1

	updateToTemporary, args, err := r.Builder.
		Update("stages").
		Set("stage_order", sq.Expr("stage_order + ?", temporaryOffset)).
		Where(sq.Eq{"project_id": projectID}).
		Where(sq.GtOrEq{"stage_order": fromOrder}).
		Where(sq.LtOrEq{"stage_order": toOrder}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, updateToTemporary, args...); err != nil {
		return r.handleErr(err)
	}

	updateToFinal, args, err := r.Builder.
		Update("stages").
		Set("stage_order", sq.Expr("stage_order - ? + ?", temporaryOffset, delta)).
		Where(sq.Eq{"project_id": projectID}).
		Where(sq.GtOrEq{"stage_order": fromOrder + temporaryOffset}).
		Where(sq.LtOrEq{"stage_order": toOrder + temporaryOffset}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, updateToFinal, args...); err != nil {
		return r.handleErr(err)
	}

	return nil
}

func normalizeInsertStageOrder(requestedOrder, maxOrder int) int {
	if requestedOrder < 1 {
		return 1
	}
	if requestedOrder > maxOrder+1 {
		return maxOrder + 1
	}
	return requestedOrder
}

func normalizeUpdateStageOrder(requestedOrder, maxOrder int) int {
	if requestedOrder < 1 {
		return 1
	}
	if requestedOrder > maxOrder {
		return maxOrder
	}
	return requestedOrder
}

func planStageOrderCompaction(deletedOrder, maxOrder int) (fromOrder, toOrder int, ok bool) {
	if deletedOrder >= maxOrder {
		return 0, 0, false
	}
	return deletedOrder + 1, maxOrder, true
}

func planStageOrderShift(currentOrder, targetOrder int) (fromOrder, toOrder, delta int, ok bool) {
	if currentOrder == targetOrder {
		return 0, 0, 0, false
	}
	if targetOrder < currentOrder {
		return targetOrder, currentOrder - 1, 1, true
	}
	return currentOrder + 1, targetOrder, -1, true
}

func (r *StageRepo) Delete(ctx context.Context, projectID, stageID string) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	if err := r.lockProject(ctx, exec, projectID); err != nil {
		return err
	}

	var deletedOrder int
	row := exec.QueryRow(ctx, "SELECT stage_order FROM stages WHERE id = $1 AND project_id = $2 FOR UPDATE", stageID, projectID)
	if err := row.Scan(&deletedOrder); err != nil {
		return r.handleErr(err)
	}

	sql, args, err := r.Builder.
		Delete("stages").
		Where(sq.Eq{"id": stageID}).
		Where(sq.Eq{"project_id": projectID}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewNotFound(entity.CodeStageNotFound, "Stage not found")
	}

	maxOrder, err := r.maxStageOrder(ctx, exec, projectID)
	if err != nil {
		return err
	}

	fromOrder, toOrder, ok := planStageOrderCompaction(deletedOrder, maxOrder)
	if !ok {
		return nil
	}

	if err := r.compactStageOrdersAfterDelete(ctx, exec, projectID, fromOrder, toOrder); err != nil {
		return err
	}

	return nil
}

func (r *StageRepo) compactStageOrdersAfterDelete(ctx context.Context, exec postgres.Executor, projectID string, fromOrder, toOrder int) error {
	if fromOrder > toOrder {
		return nil
	}

	maxOrder, err := r.maxStageOrder(ctx, exec, projectID)
	if err != nil {
		return err
	}
	temporaryOffset := maxOrder + 1

	updateToTemporary, args, err := r.Builder.
		Update("stages").
		Set("stage_order", sq.Expr("stage_order + ?", temporaryOffset)).
		Where(sq.Eq{"project_id": projectID}).
		Where(sq.GtOrEq{"stage_order": fromOrder}).
		Where(sq.LtOrEq{"stage_order": toOrder}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, updateToTemporary, args...); err != nil {
		return r.handleErr(err)
	}

	updateToFinal, args, err := r.Builder.
		Update("stages").
		Set("stage_order", sq.Expr("stage_order - ? - 1", temporaryOffset)).
		Where(sq.Eq{"project_id": projectID}).
		Where(sq.GtOrEq{"stage_order": fromOrder + temporaryOffset}).
		Where(sq.LtOrEq{"stage_order": toOrder + temporaryOffset}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, updateToFinal, args...); err != nil {
		return r.handleErr(err)
	}

	return nil
}
