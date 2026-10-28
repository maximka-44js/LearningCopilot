package persistent

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/projects/entity"
	projectrepo "github.com/maximka-44js/lc/internal/projects/repo"
	"github.com/maximka-44js/lc/pkg/postgres"
)

type AssessmentSnapshotRepo struct {
	*postgres.Postgres
}

func NewAssessmentSnapshotRepo(pg *postgres.Postgres) *AssessmentSnapshotRepo {
	return &AssessmentSnapshotRepo{pg}
}

func (r *AssessmentSnapshotRepo) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeAssessmentSnapshotInvalid, "Assessment snapshot not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return d.NewConflict(entity.CodeAssessmentSnapshotInvalid, "Assessment snapshot already exists")
		case "23503":
			return d.NewValidationError(entity.CodeAssessmentSnapshotInvalid, "Referenced resource not found")
		case "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeAssessmentSnapshotInvalid, "Assessment snapshot data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func (r *AssessmentSnapshotRepo) Store(ctx context.Context, snapshot *entity.AssessmentSnapshot) error {
	if snapshot == nil {
		return d.NewValidationError(entity.CodeAssessmentSnapshotInvalid, "Assessment snapshot payload is required")
	}
	if err := snapshot.Validate(); err != nil {
		return err
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Insert("assessment_snapshots").
		Columns(
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
		Values(
			snapshot.ID,
			snapshot.TaskID,
			snapshot.BlockType,
			snapshot.Source,
			snapshot.RawScore,
			snapshot.NormalizedScore,
			snapshot.ConfidenceBefore,
			snapshot.ConfidenceAfter,
			snapshot.RecordedAt,
			snapshot.TraceID,
			snapshot.SourceReference,
			snapshot.SourceReferenceType,
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

func (r *AssessmentSnapshotRepo) ListByTask(
	ctx context.Context,
	taskID string,
	filter projectrepo.AssessmentSnapshotFilter,
) ([]entity.AssessmentSnapshot, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	query := r.Builder.
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
		Where(sq.Eq{"task_id": taskID})

	if len(filter.BlockTypes) > 0 {
		blockTypes := make([]string, 0, len(filter.BlockTypes))
		for _, blockType := range filter.BlockTypes {
			blockTypes = append(blockTypes, string(blockType))
		}
		query = query.Where(sq.Eq{"block_type": blockTypes})
	}

	if len(filter.Sources) > 0 {
		sources := make([]string, 0, len(filter.Sources))
		for _, source := range filter.Sources {
			sources = append(sources, string(source))
		}
		query = query.Where(sq.Eq{"source": sources})
	}

	if filter.From != nil {
		query = query.Where(sq.GtOrEq{"recorded_at": *filter.From})
	}
	if filter.To != nil {
		query = query.Where(sq.LtOrEq{"recorded_at": *filter.To})
	}

	query = query.OrderBy("recorded_at DESC", "id DESC")
	if filter.Limit > 0 {
		query = query.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		query = query.Offset(uint64(filter.Offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	items := make([]entity.AssessmentSnapshot, 0)
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
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return items, nil
}

func (r *AssessmentSnapshotRepo) GetLatestByTask(ctx context.Context, taskID string) (*entity.AssessmentSnapshot, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

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
		OrderBy("recorded_at DESC", "id DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var item entity.AssessmentSnapshot
	if err := exec.QueryRow(ctx, sql, args...).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, r.handleErr(err)
	}

	return &item, nil
}

func (r *AssessmentSnapshotRepo) GetLatestByTaskAndBlock(
	ctx context.Context,
	taskID string,
	blockType entity.CompetencyBlockType,
) (*entity.AssessmentSnapshot, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

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
		Where(sq.Eq{"block_type": blockType}).
		OrderBy("recorded_at DESC", "id DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var item entity.AssessmentSnapshot
	if err := exec.QueryRow(ctx, sql, args...).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, r.handleErr(err)
	}

	return &item, nil
}
