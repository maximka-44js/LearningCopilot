package persistent

import (
	"context"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/pkg/postgres"
)

type FolderRepo struct {
	*postgres.Postgres
}

func NewFolderRepo(pg *postgres.Postgres) *FolderRepo {
	return &FolderRepo{pg}
}

func (r *FolderRepo) handleErr(err error) *d.AppError {
	return handleRepoErr(err, RepoErrMapping{
		DuplicateCode:   entity.CodeFolderNameInvalid,
		DuplicateMsg:    "Folder with this name already exists",
		ReferenceCode:   entity.CodeFolderAccessInvalid,
		ReferenceMsg:    "Referenced resource not found",
		InvalidDataCode: entity.CodeUndefined,
		InvalidDataMsg:  "Invalid data",
		NotFoundCode:    entity.CodeFolderNotFound,
		NotFoundMsg:     "Folder not found",
	})
}

func (r *FolderRepo) GetByID(ctx context.Context, userID, folderID string) (entity.Folder, error) {
	sql, args, err := r.Builder.
		Select("id", "user_id", "library_id", "COALESCE(parent_id::text, '')", "name", "created_at", "updated_at", "deleted_at").
		From("folders").
		Where(sq.Eq{"id": folderID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()

	if err != nil {
		return entity.Folder{}, d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	var folder entity.Folder
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&folder.ID,
		&folder.UserID,
		&folder.LibID,
		&folder.ParentID,
		&folder.Name,
		&folder.CreatedAt,
		&folder.UpdatedAt,
		&folder.DeletedAt,
	)
	if err != nil {
		return entity.Folder{}, r.handleErr(err)
	}
	return folder, nil
}

// Update -.
func (r *FolderRepo) Update(ctx context.Context, fd *entity.Folder) error {
	var parentID interface{}
	if fd.ParentID == "" {
		parentID = nil
	} else {
		parentID = fd.ParentID
	}

	sql, args, err := r.Builder.
		Update("folders").
		Set("parent_id", parentID).
		Set("name", fd.Name).
		Set("updated_at", fd.UpdatedAt).
		Where(sq.Eq{"id": fd.ID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()

	if err != nil {
		return d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}

	return nil
}

// Delete -.
func (r *FolderRepo) Delete(ctx context.Context, folderID string) (int64, error) {
	tx := postgres.TxFromContext(ctx)
	ownsTx := false
	if tx == nil {
		var err error
		tx, err = r.Pool.Begin(ctx)
		if err != nil {
			return 0, d.NewAppError(
				entity.CodeUndefined,
				"Internal error",
				d.KindInternal, nil)
		}
		ownsTx = true
		defer func() {
			_ = tx.Rollback(ctx)
		}()
	}

	var deletedSize int64
	sizeSQL := `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE id = $1 AND deleted_at IS NULL
	UNION ALL
	SELECT f.id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.deleted_at IS NULL
)
SELECT COALESCE(SUM(size), 0)
FROM files
WHERE folder_id IN (SELECT id FROM folder_tree) AND deleted_at IS NULL
`
	if err := tx.QueryRow(ctx, sizeSQL, folderID).Scan(&deletedSize); err != nil {
		return 0, r.handleErr(err)
	}

	filesSQL := `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE id = $1 AND deleted_at IS NULL
	UNION ALL
	SELECT f.id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.deleted_at IS NULL
)
UPDATE files
SET status = $2, deleted_at = $3, updated_at = $3
WHERE folder_id IN (SELECT id FROM folder_tree) AND deleted_at IS NULL
`
	now := time.Now().UTC()
	_, err := tx.Exec(ctx, filesSQL, folderID, entity.FileStatusDeleted, now)
	if err != nil {
		return 0, r.handleErr(err)
	}

	foldersSQL := `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE id = $1 AND deleted_at IS NULL
	UNION ALL
	SELECT f.id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.deleted_at IS NULL
)
UPDATE folders
SET deleted_at = $2, updated_at = $2
WHERE id IN (SELECT id FROM folder_tree) AND deleted_at IS NULL
`
	res, err := tx.Exec(ctx, foldersSQL, folderID, now)
	if err != nil {
		return 0, r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return 0, d.NewAppError(
			entity.CodeFolderNotFound,
			"Folder not found",
			d.KindNotFound, nil)
	}
	if ownsTx {
		if err = tx.Commit(ctx); err != nil {
			return 0, r.handleErr(err)
		}
	}
	return deletedSize, nil
}

func (r *FolderRepo) ListByLibraryAndParent(ctx context.Context, _ string, libID, parentID string) ([]entity.Folder, error) {
	libID = strings.TrimSpace(libID)
	parentID = strings.TrimSpace(parentID)
	if libID == "" {
		return nil, d.NewAppError(entity.CodeLibNotFound, "Library not found", d.KindNotFound, nil)
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	query := r.Builder.
		Select("id", "user_id", "library_id", "COALESCE(parent_id::text, '')", "name", "created_at", "updated_at", "deleted_at").
		From("folders").
		Where(sq.Eq{"library_id": libID}).
		Where(sq.Eq{"deleted_at": nil})

	if parentID == "" {
		query = query.Where(sq.Eq{"parent_id": nil})
	} else {
		query = query.Where(sq.Eq{"parent_id": parentID})
	}

	sql, args, err := query.OrderBy("name ASC", "created_at ASC").ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	res := make([]entity.Folder, 0)
	for rows.Next() {
		var folder entity.Folder
		if err := rows.Scan(
			&folder.ID,
			&folder.UserID,
			&folder.LibID,
			&folder.ParentID,
			&folder.Name,
			&folder.CreatedAt,
			&folder.UpdatedAt,
			&folder.DeletedAt,
		); err != nil {
			return nil, r.handleErr(err)
		}
		res = append(res, folder)
	}
	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return res, nil
}

// Store -.
func (r *FolderRepo) Store(ctx context.Context, fd *entity.Folder) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)

	var parentID interface{}
	if fd.ParentID == "" {
		parentID = nil
	} else {
		parentID = fd.ParentID
	}

	sql, args, err := r.Builder.
		Insert("folders").
		Columns("id", "user_id", "library_id", "parent_id", "name", "created_at", "updated_at").
		Values(fd.ID, fd.UserID, fd.LibID, parentID, fd.Name, fd.CreatedAt, fd.UpdatedAt).
		ToSql()

	if err != nil {
		return d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	_, err = exec.Exec(ctx, sql, args...)

	if err != nil {
		return r.handleErr(err)
	}
	return nil
}
