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

type ProjectRepo struct {
	*postgres.Postgres
}

func NewProjectRepo(pg *postgres.Postgres) *ProjectRepo {
	return &ProjectRepo{pg}
}

func (r *ProjectRepo) handleErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeProjectNotFound, "Project not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return d.NewConflict(entity.CodeProjectNameInvalid, "Project with this name already exists")
		case "23503":
			return d.NewValidationError(entity.CodeProjectInvalid, "Referenced resource not found")
		case "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeProjectInvalid, "Project data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func (r *ProjectRepo) Store(ctx context.Context, project *entity.Project) error {
	if project == nil {
		return d.NewValidationError(entity.CodeProjectInvalid, "Project payload is required")
	}
	if err := project.Validate(); err != nil {
		return err
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	metadataJSON, err := json.Marshal(project.Metadata)
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	sql, args, err := r.Builder.
		Insert("projects").
		Columns(
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
		Values(
			project.ID,
			project.OwnerID,
			project.Name,
			project.Description,
			project.DeadlineAt,
			project.Priority,
			project.Status,
			project.Coverage,
			project.Readiness,
			project.Risk,
			sq.Expr("?::jsonb", string(metadataJSON)),
			project.CreatedAt,
			project.UpdatedAt,
			project.ArchivedAt,
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

func (r *ProjectRepo) GetByID(ctx context.Context, ownerID, projectID string) (entity.Project, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

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
		ToSql()
	if err != nil {
		return entity.Project{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var project entity.Project
	var metadataRaw []byte

	row := exec.QueryRow(ctx, sql, args...)
	if err := row.Scan(
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
		if err := json.Unmarshal(metadataRaw, &project.Metadata); err != nil {
			return entity.Project{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
		}
	}

	return project, nil
}

func (r *ProjectRepo) ListByOwner(ctx context.Context, ownerID string) ([]entity.Project, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

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
		Where(sq.Eq{"owner_id": ownerID}).
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

	projects := make([]entity.Project, 0)
	for rows.Next() {
		var project entity.Project
		var metadataRaw []byte

		if err := rows.Scan(
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
			return nil, r.handleErr(err)
		}

		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &project.Metadata); err != nil {
				return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
			}
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return projects, nil
}

func (r *ProjectRepo) Update(ctx context.Context, project *entity.Project) error {
	if project == nil {
		return d.NewValidationError(entity.CodeProjectInvalid, "Project payload is required")
	}
	if err := project.Validate(); err != nil {
		return err
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	metadataJSON, err := json.Marshal(project.Metadata)
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	sql, args, err := r.Builder.
		Update("projects").
		Set("name", project.Name).
		Set("description", project.Description).
		Set("deadline_at", project.DeadlineAt).
		Set("priority", project.Priority).
		Set("status", project.Status).
		Set("coverage", project.Coverage).
		Set("readiness", project.Readiness).
		Set("risk", project.Risk).
		Set("metadata", sq.Expr("?::jsonb", string(metadataJSON))).
		Set("updated_at", project.UpdatedAt).
		Set("archived_at", project.ArchivedAt).
		Where(sq.Eq{"id": project.ID}).
		Where(sq.Eq{"owner_id": project.OwnerID}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewNotFound(entity.CodeProjectNotFound, "Project not found")
	}

	return nil
}

func (r *ProjectRepo) Delete(ctx context.Context, ownerID, projectID string) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	sql, args, err := r.Builder.
		Delete("projects").
		Where(sq.Eq{"id": projectID}).
		Where(sq.Eq{"owner_id": ownerID}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewNotFound(entity.CodeProjectNotFound, "Project not found")
	}

	return nil
}
