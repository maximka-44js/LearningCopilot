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

type TaskRepo struct {
	*postgres.Postgres
}

func NewTaskRepo(pg *postgres.Postgres) *TaskRepo {
	return &TaskRepo{pg}
}

func (r *TaskRepo) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeTaskNotFound, "Task not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return d.NewConflict(entity.CodeTaskInvalid, "Task already exists")
		case "23503":
			return d.NewValidationError(entity.CodeTaskInvalid, "Referenced resource not found")
		case "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeTaskInvalid, "Task data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func (r *TaskRepo) Store(ctx context.Context, task *entity.Task) error {
	if task == nil {
		return d.NewValidationError(entity.CodeTaskInvalid, "Task payload is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}

	return r.WithinTx(ctx, func(ctx context.Context) error {
		exec := postgres.ExecutorFromContext(ctx, r.Pool)

		sql, args, err := r.Builder.
			Insert("tasks").
			Columns(
				"id",
				"stage_id",
				"outcome",
				"difficulty",
				"priority",
				"exam_ready_min_mastery",
				"require_non_critical_block_risk",
				"deadline_at",
				"status",
				"coverage",
				"readiness",
				"risk",
				"created_at",
				"updated_at",
				"archived_at",
			).
			Values(
				task.ID,
				task.StageID,
				task.Outcome,
				task.Difficulty,
				task.Priority,
				task.Policy.ExamReadyMinMastery,
				task.Policy.RequireNonCriticalBlockRisk,
				task.DeadlineAt,
				task.Status,
				task.Coverage,
				task.Readiness,
				task.Risk,
				task.CreatedAt,
				task.UpdatedAt,
				task.ArchivedAt,
			).
			ToSql()
		if err != nil {
			return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
		}

		if _, err := exec.Exec(ctx, sql, args...); err != nil {
			return r.handleErr(err)
		}

		if err := r.replaceBlocks(ctx, exec, task.ID, task.Blocks, false); err != nil {
			return err
		}

		return nil
	})
}

func (r *TaskRepo) GetByID(ctx context.Context, stageID, taskID string) (entity.Task, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Select(
			"id",
			"stage_id",
			"outcome",
			"difficulty",
			"priority",
			"exam_ready_min_mastery",
			"require_non_critical_block_risk",
			"deadline_at",
			"status",
			"coverage",
			"readiness",
			"risk",
			"created_at",
			"updated_at",
			"archived_at",
		).
		From("tasks").
		Where(sq.Eq{"id": taskID}).
		Where(sq.Eq{"stage_id": stageID}).
		ToSql()
	if err != nil {
		return entity.Task{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var task entity.Task
	row := exec.QueryRow(ctx, sql, args...)
	if err := row.Scan(
		&task.ID,
		&task.StageID,
		&task.Outcome,
		&task.Difficulty,
		&task.Priority,
		&task.Policy.ExamReadyMinMastery,
		&task.Policy.RequireNonCriticalBlockRisk,
		&task.DeadlineAt,
		&task.Status,
		&task.Coverage,
		&task.Readiness,
		&task.Risk,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.ArchivedAt,
	); err != nil {
		return entity.Task{}, r.handleErr(err)
	}

	blocksByTaskID, err := r.listBlocks(ctx, exec, []string{task.ID})
	if err != nil {
		return entity.Task{}, err
	}
	task.Blocks = blocksByTaskID[task.ID]

	return task, nil
}

func (r *TaskRepo) GetByTaskID(ctx context.Context, taskID string) (entity.Task, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Select(
			"id",
			"stage_id",
			"outcome",
			"difficulty",
			"priority",
			"exam_ready_min_mastery",
			"require_non_critical_block_risk",
			"deadline_at",
			"status",
			"coverage",
			"readiness",
			"risk",
			"created_at",
			"updated_at",
			"archived_at",
		).
		From("tasks").
		Where(sq.Eq{"id": taskID}).
		Limit(1).
		ToSql()
	if err != nil {
		return entity.Task{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var task entity.Task

	row := exec.QueryRow(ctx, sql, args...)
	if err := row.Scan(
		&task.ID,
		&task.StageID,
		&task.Outcome,
		&task.Difficulty,
		&task.Priority,
		&task.Policy.ExamReadyMinMastery,
		&task.Policy.RequireNonCriticalBlockRisk,
		&task.DeadlineAt,
		&task.Status,
		&task.Coverage,
		&task.Readiness,
		&task.Risk,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.ArchivedAt,
	); err != nil {
		return entity.Task{}, r.handleErr(err)
	}

	blocksByTaskID, err := r.listBlocks(ctx, exec, []string{task.ID})
	if err != nil {
		return entity.Task{}, err
	}
	if blocksByTaskID != nil {
		task.Blocks = blocksByTaskID[task.ID]
	}

	return task, nil
}

func (r *TaskRepo) GetScopeByTaskID(ctx context.Context, taskID string) (string, string, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Select(
			"s.project_id",
			"t.stage_id",
		).
		From("tasks t").
		Join("stages s ON s.id = t.stage_id").
		Where(sq.Eq{"t.id": taskID}).
		Limit(1).
		ToSql()
	if err != nil {
		return "", "", d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	row := exec.QueryRow(ctx, sql, args...)
	var projectID string
	var stageID string
	if err := row.Scan(&projectID, &stageID); err != nil {
		return "", "", r.handleErr(err)
	}

	return projectID, stageID, nil
}

func (r *TaskRepo) ListByStage(ctx context.Context, stageID string) ([]entity.Task, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Select(
			"id",
			"stage_id",
			"outcome",
			"difficulty",
			"priority",
			"exam_ready_min_mastery",
			"require_non_critical_block_risk",
			"deadline_at",
			"status",
			"coverage",
			"readiness",
			"risk",
			"created_at",
			"updated_at",
			"archived_at",
		).
		From("tasks").
		Where(sq.Eq{"stage_id": stageID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	tasks := make([]entity.Task, 0)
	taskIDs := make([]string, 0)
	for rows.Next() {
		var task entity.Task
		if err := rows.Scan(
			&task.ID,
			&task.StageID,
			&task.Outcome,
			&task.Difficulty,
			&task.Priority,
			&task.Policy.ExamReadyMinMastery,
			&task.Policy.RequireNonCriticalBlockRisk,
			&task.DeadlineAt,
			&task.Status,
			&task.Coverage,
			&task.Readiness,
			&task.Risk,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.ArchivedAt,
		); err != nil {
			return nil, r.handleErr(err)
		}
		tasks = append(tasks, task)
		taskIDs = append(taskIDs, task.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	if len(taskIDs) == 0 {
		return tasks, nil
	}

	blocksByTaskID, err := r.listBlocks(ctx, exec, taskIDs)
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		tasks[i].Blocks = blocksByTaskID[tasks[i].ID]
	}

	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, task *entity.Task) error {
	if task == nil {
		return d.NewValidationError(entity.CodeTaskInvalid, "Task payload is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}

	return r.WithinTx(ctx, func(ctx context.Context) error {
		exec := postgres.ExecutorFromContext(ctx, r.Pool)

		sql, args, err := r.Builder.
			Update("tasks").
			Set("outcome", task.Outcome).
			Set("difficulty", task.Difficulty).
			Set("priority", task.Priority).
			Set("exam_ready_min_mastery", task.Policy.ExamReadyMinMastery).
			Set("require_non_critical_block_risk", task.Policy.RequireNonCriticalBlockRisk).
			Set("deadline_at", task.DeadlineAt).
			Set("status", task.Status).
			Set("coverage", task.Coverage).
			Set("readiness", task.Readiness).
			Set("risk", task.Risk).
			Set("updated_at", task.UpdatedAt).
			Set("archived_at", task.ArchivedAt).
			Where(sq.Eq{"id": task.ID}).
			Where(sq.Eq{"stage_id": task.StageID}).
			ToSql()
		if err != nil {
			return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
		}

		res, err := exec.Exec(ctx, sql, args...)
		if err != nil {
			return r.handleErr(err)
		}
		if res.RowsAffected() == 0 {
			return d.NewNotFound(entity.CodeTaskNotFound, "Task not found")
		}

		if err := r.replaceBlocks(ctx, exec, task.ID, task.Blocks, true); err != nil {
			return err
		}

		return nil
	})
}

func (r *TaskRepo) Delete(ctx context.Context, stageID, taskID string) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Delete("tasks").
		Where(sq.Eq{"id": taskID}).
		Where(sq.Eq{"stage_id": stageID}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewNotFound(entity.CodeTaskNotFound, "Task not found")
	}

	return nil
}

func (r *TaskRepo) replaceBlocks(
	ctx context.Context,
	exec postgres.Executor,
	taskID string,
	blocks []entity.CompetencyBlock,
	deleteExisting bool,
) error {
	if deleteExisting {
		deleteSQL, deleteArgs, err := r.Builder.
			Delete("task_competency_blocks").
			Where(sq.Eq{"task_id": taskID}).
			ToSql()
		if err != nil {
			return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
		}

		if _, err := exec.Exec(ctx, deleteSQL, deleteArgs...); err != nil {
			return r.handleErr(err)
		}
	}

	insertBuilder := r.Builder.
		Insert("task_competency_blocks").
		Columns(
			"task_id",
			"block_type",
			"weight",
			"mastery",
			"last_assessed_at",
			"decay_profile",
			"risk_state",
			"created_at",
			"updated_at",
		)

	for _, block := range blocks {
		insertBuilder = insertBuilder.Values(
			taskID,
			block.Type,
			block.Weight,
			block.Mastery,
			block.LastAssessedAt,
			block.DecayProfile,
			block.RiskState,
			sq.Expr("now()"),
			sq.Expr("now()"),
		)
	}

	insertSQL, insertArgs, err := insertBuilder.ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, insertSQL, insertArgs...); err != nil {
		return r.handleErr(err)
	}

	return nil
}

func (r *TaskRepo) listBlocks(
	ctx context.Context,
	exec postgres.Executor,
	taskIDs []string,
) (map[string][]entity.CompetencyBlock, error) {
	result := make(map[string][]entity.CompetencyBlock, len(taskIDs))
	if len(taskIDs) == 0 {
		return result, nil
	}

	sql, args, err := r.Builder.
		Select(
			"task_id",
			"block_type",
			"weight",
			"mastery",
			"last_assessed_at",
			"decay_profile",
			"risk_state",
		).
		From("task_competency_blocks").
		Where(sq.Eq{"task_id": taskIDs}).
		OrderBy("task_id ASC", "block_type ASC").
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskID string
		var block entity.CompetencyBlock
		var decayProfile *string
		if err := rows.Scan(
			&taskID,
			&block.Type,
			&block.Weight,
			&block.Mastery,
			&block.LastAssessedAt,
			&decayProfile,
			&block.RiskState,
		); err != nil {
			return nil, r.handleErr(err)
		}
		if decayProfile != nil {
			block.DecayProfile = *decayProfile
		}
		result[taskID] = append(result[taskID], block)
	}

	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return result, nil
}
