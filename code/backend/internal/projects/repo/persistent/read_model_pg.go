package persistent

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/entity"
	projectrepo "github.com/maximka-44js/lc/internal/projects/repo"
	"github.com/maximka-44js/lc/pkg/postgres"
)

type ReadModelRepo struct {
	*postgres.Postgres
}

func NewReadModelRepo(pg *postgres.Postgres) *ReadModelRepo {
	return &ReadModelRepo{pg}
}

func (r *ReadModelRepo) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeProjectNotFound, "Resource not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503", "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeUndefined, "Read model data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func (r *ReadModelRepo) ProjectDashboard(
	ctx context.Context,
	ownerID, projectID string,
	at time.Time,
) (projectrepo.ProjectDashboardData, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	project, err := r.getProject(ctx, exec, ownerID, projectID)
	if err != nil {
		return projectrepo.ProjectDashboardData{}, err
	}

	stages, err := r.getStagesByProject(ctx, exec, projectID)
	if err != nil {
		return projectrepo.ProjectDashboardData{}, err
	}

	criticalItems, err := r.listCriticalDeadlines(ctx, exec, ownerID, projectID, at)
	if err != nil {
		return projectrepo.ProjectDashboardData{}, err
	}

	return projectrepo.ProjectDashboardData{
		Project:            project,
		Stages:             stages,
		CriticalToDeadline: criticalItems,
	}, nil
}

func (r *ReadModelRepo) StageBoard(
	ctx context.Context,
	ownerID, projectID, stageID string,
	at time.Time,
) (projectrepo.StageBoardData, error) {
	_ = at
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	project, err := r.getProject(ctx, exec, ownerID, projectID)
	if err != nil {
		return projectrepo.StageBoardData{}, err
	}

	stage, err := r.getStage(ctx, exec, projectID, stageID)
	if err != nil {
		return projectrepo.StageBoardData{}, err
	}

	tasks, err := r.getTasksByStage(ctx, exec, stageID)
	if err != nil {
		return projectrepo.StageBoardData{}, err
	}

	return projectrepo.StageBoardData{
		Project: project,
		Stage:   stage,
		Tasks:   tasks,
	}, nil
}

func (r *ReadModelRepo) TaskReadiness(
	ctx context.Context,
	ownerID, projectID, stageID, taskID string,
	at time.Time,
) (projectrepo.TaskReadinessData, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	project, err := r.getProject(ctx, exec, ownerID, projectID)
	if err != nil {
		return projectrepo.TaskReadinessData{}, err
	}

	stage, err := r.getStage(ctx, exec, projectID, stageID)
	if err != nil {
		return projectrepo.TaskReadinessData{}, err
	}

	task, err := r.getTask(ctx, exec, stageID, taskID)
	if err != nil {
		return projectrepo.TaskReadinessData{}, err
	}

	snapshots, err := r.latestSnapshotsByTaskAt(ctx, exec, taskID, at)
	if err != nil {
		return projectrepo.TaskReadinessData{}, err
	}

	return projectrepo.TaskReadinessData{
		Project:         project,
		Stage:           stage,
		Task:            task,
		LatestSnapshots: snapshots,
	}, nil
}

func (r *ReadModelRepo) getProject(
	ctx context.Context,
	exec postgres.Executor,
	ownerID, projectID string,
) (entity.Project, error) {
	sql, args, err := r.Builder.
		Select(
			"id",
			"owner_id",
			"name",
			"description",
			"deadline_at",
			"priority",
			"status",
			"coverage",
			"readiness",
			"risk",
			"metadata",
			"created_at",
			"updated_at",
			"archived_at",
		).
		From("projects").
		Where(sq.Eq{"id": projectID}).
		Where(sq.Eq{"owner_id": ownerID}).
		Limit(1).
		ToSql()
	if err != nil {
		return entity.Project{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var project entity.Project
	var metadataRaw []byte
	if err := exec.QueryRow(ctx, sql, args...).Scan(
		&project.ID,
		&project.OwnerID,
		&project.Name,
		&project.Description,
		&project.DeadlineAt,
		&project.Priority,
		&project.Status,
		&project.Coverage,
		&project.Readiness,
		&project.Risk,
		&metadataRaw,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.ArchivedAt,
	); err != nil {
		return entity.Project{}, r.handleErr(err)
	}

	if len(metadataRaw) > 0 {
		if err := jsonUnmarshalMetadata(metadataRaw, &project); err != nil {
			return entity.Project{}, err
		}
	}

	return project, nil
}

func (r *ReadModelRepo) getStage(
	ctx context.Context,
	exec postgres.Executor,
	projectID, stageID string,
) (entity.Stage, error) {
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
		Limit(1).
		ToSql()
	if err != nil {
		return entity.Stage{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var stage entity.Stage
	if err := exec.QueryRow(ctx, sql, args...).Scan(
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

func (r *ReadModelRepo) getStagesByProject(
	ctx context.Context,
	exec postgres.Executor,
	projectID string,
) ([]entity.Stage, error) {
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

	items := make([]entity.Stage, 0)
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
		items = append(items, stage)
	}
	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return items, nil
}

func (r *ReadModelRepo) getTask(
	ctx context.Context,
	exec postgres.Executor,
	stageID, taskID string,
) (entity.Task, error) {
	tasks, err := r.getTasks(ctx, exec, sq.Eq{"t.stage_id": stageID, "t.id": taskID}, 1)
	if err != nil {
		return entity.Task{}, err
	}
	if len(tasks) == 0 {
		return entity.Task{}, d.NewNotFound(entity.CodeTaskNotFound, "Task not found")
	}
	return tasks[0], nil
}

func (r *ReadModelRepo) getTasksByStage(
	ctx context.Context,
	exec postgres.Executor,
	stageID string,
) ([]entity.Task, error) {
	return r.getTasks(ctx, exec, sq.Eq{"t.stage_id": stageID}, 0)
}

func (r *ReadModelRepo) getTasks(
	ctx context.Context,
	exec postgres.Executor,
	where sq.Sqlizer,
	limit uint64,
) ([]entity.Task, error) {
	builder := r.Builder.
		Select(
			"t.id",
			"t.stage_id",
			"t.outcome",
			"t.difficulty",
			"t.priority",
			"t.exam_ready_min_mastery",
			"t.require_non_critical_block_risk",
			"t.deadline_at",
			"t.status",
			"t.coverage",
			"t.readiness",
			"t.risk",
			"t.created_at",
			"t.updated_at",
			"t.archived_at",
		).
		From("tasks t").
		Where(where).
		OrderBy("t.created_at DESC")
	if limit > 0 {
		builder = builder.Limit(limit)
	}

	sql, args, err := builder.ToSql()
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
	if len(tasks) == 0 {
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

func (r *ReadModelRepo) listCriticalDeadlines(
	ctx context.Context,
	exec postgres.Executor,
	ownerID, projectID string,
	at time.Time,
) ([]projectrepo.CriticalDeadlineItem, error) {
	windowEnd := at.Add(72 * time.Hour)

	sql, args, err := r.Builder.
		Select("s.id", "t.id", "t.deadline_at", "t.status", "t.risk").
		From("tasks t").
		Join("stages s ON s.id = t.stage_id").
		Join("projects p ON p.id = s.project_id").
		Where(sq.Eq{"p.owner_id": ownerID}).
		Where(sq.Eq{"p.id": projectID}).
		Where("t.deadline_at IS NOT NULL").
		Where(sq.LtOrEq{"t.deadline_at": windowEnd}).
		Where(sq.NotEq{"t.status": entity.TaskStatusArchived}).
		Where(sq.Or{
			sq.LtOrEq{"t.deadline_at": at},
			sq.Eq{"t.status": entity.TaskStatusAtRisk},
			sq.Eq{"t.risk": entity.RiskLevelCritical},
			sq.Eq{"t.risk": entity.RiskLevelHigh},
		}).
		OrderBy("t.deadline_at ASC").
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	items := make([]projectrepo.CriticalDeadlineItem, 0)
	for rows.Next() {
		var item projectrepo.CriticalDeadlineItem
		var status entity.TaskStatus
		var risk entity.RiskLevel
		if err := rows.Scan(&item.StageID, &item.TaskID, &item.DeadlineAt, &status, &risk); err != nil {
			return nil, r.handleErr(err)
		}
		item.Reason = criticalReason(item.DeadlineAt, at, status, risk)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return items, nil
}

func (r *ReadModelRepo) latestSnapshotsByTaskAt(
	ctx context.Context,
	exec postgres.Executor,
	taskID string,
	at time.Time,
) ([]entity.AssessmentSnapshot, error) {
	sql, args, err := r.Builder.
		Select(
			"id",
			"task_id",
			"block_type",
			"source",
			"raw_score",
			"normalized_score",
			"confidence_before",
			"confidence_after",
			"recorded_at",
			"trace_id",
			"source_reference",
			"source_reference_type",
		).
		From("assessment_snapshots").
		Where(sq.Eq{"task_id": taskID}).
		Where(sq.LtOrEq{"recorded_at": at}).
		OrderBy("block_type ASC", "recorded_at DESC", "id DESC").
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	items := make([]entity.AssessmentSnapshot, 0)
	seen := make(map[entity.CompetencyBlockType]struct{})
	for rows.Next() {
		var item entity.AssessmentSnapshot
		if err := rows.Scan(
			&item.ID,
			&item.TaskID,
			&item.BlockType,
			&item.Source,
			&item.RawScore,
			&item.NormalizedScore,
			&item.ConfidenceBefore,
			&item.ConfidenceAfter,
			&item.RecordedAt,
			&item.TraceID,
			&item.SourceReference,
			&item.SourceReferenceType,
		); err != nil {
			return nil, r.handleErr(err)
		}
		if _, ok := seen[item.BlockType]; ok {
			continue
		}
		seen[item.BlockType] = struct{}{}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return items, nil
}

func criticalReason(deadlineAt *time.Time, at time.Time, status entity.TaskStatus, risk entity.RiskLevel) string {
	if deadlineAt == nil {
		return "critical"
	}
	if !deadlineAt.After(at) {
		return "overdue"
	}
	if status == entity.TaskStatusAtRisk {
		return "at_risk_near_deadline"
	}
	if risk == entity.RiskLevelCritical {
		return "critical_risk_near_deadline"
	}
	if risk == entity.RiskLevelHigh {
		return "high_risk_near_deadline"
	}
	return "near_deadline"
}

func (r *ReadModelRepo) listBlocks(
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

func jsonUnmarshalMetadata(raw []byte, project *entity.Project) error {
	if err := json.Unmarshal(raw, &project.Metadata); err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}
	return nil
}
