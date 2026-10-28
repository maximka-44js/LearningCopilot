package persistent

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/pkg/postgres"
)

// LibRepo
type LibRepo struct {
	*postgres.Postgres
}

// NewLibRepo
func NewLibRepo(pg *postgres.Postgres) *LibRepo {
	return &LibRepo{pg}
}

func (r *LibRepo) handleErr(err error) *d.AppError {
	return handleRepoErr(err, RepoErrMapping{
		DuplicateCode:   entity.CodeLibraryNameInvalid,
		DuplicateMsg:    "Library with this name already exists",
		ReferenceCode:   entity.CodeLibraryAccessInvalid,
		ReferenceMsg:    "Referenced resource not found",
		InvalidDataCode: entity.CodeUndefined,
		InvalidDataMsg:  "Invalid data",
		NotFoundCode:    entity.CodeLibNotFound,
		NotFoundMsg:     "Library not found",
	})
}

// Store -.
func (r *LibRepo) Store(ctx context.Context, lib *entity.Library) error {
	tx := postgres.TxFromContext(ctx)
	ownsTx := false
	if tx == nil {
		var err error
		tx, err = r.Pool.Begin(ctx)
		if err != nil {
			return d.NewAppError(
				entity.CodeUndefined,
				"Internal error",
				d.KindInternal, nil)
		}
		ownsTx = true
		defer func() {
			_ = tx.Rollback(ctx)
		}()
	}

	// Вставка библиотеки
	sql, args, err := r.Builder.
		Insert("libraries").
		Columns("id", "user_id", "title", "description", "public", "created_at", "updated_at").
		Values(lib.ID, lib.UserID, lib.Title, lib.Description, lib.Public, lib.CreatedAt, lib.UpdatedAt).
		ToSql()

	if err != nil {
		return d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}

	// Вставка доступов, если они есть
	if len(lib.Accesses) > 0 {
		insertBuilder := r.Builder.
			Insert("library_accesses").
			Columns("library_id", "user_id", "role")

		for _, access := range lib.Accesses {
			insertBuilder = insertBuilder.Values(lib.ID, access.UserID, access.Role)
		}

		sql, args, err = insertBuilder.ToSql()
		if err != nil {
			return d.NewAppError(
				entity.CodeUndefined,
				"Internal error",
				d.KindInternal, nil)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return r.handleErr(err)
		}
	}

	if ownsTx {
		if err = tx.Commit(ctx); err != nil {
			return r.handleErr(err)
		}
	}

	return nil
}

// GetByID -.
func (r *LibRepo) GetByID(ctx context.Context, userID, libID string) (entity.Library, error) {
	sql, args, err := r.Builder.
		Select("l.id, l.user_id, l.title, l.description, l.public, l.created_at, l.updated_at").
		From("libraries l").
		LeftJoin("library_accesses la ON la.library_id = l.id AND la.user_id = ?", userID).
		Where(sq.Eq{"l.id": libID}).
		Where(sq.Eq{"l.deleted_at": nil}).
		Where(sq.Or{
			sq.Eq{"l.user_id": userID},
			sq.Expr("l.public = true"),
			sq.Expr("la.user_id IS NOT NULL"),
		}).
		ToSql()

	if err != nil {
		return entity.Library{}, d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	var lib entity.Library

	err = r.Pool.QueryRow(ctx, sql, args...).
		Scan(&lib.ID, &lib.UserID, &lib.Title, &lib.Description, &lib.Public, &lib.CreatedAt, &lib.UpdatedAt)

	if err != nil {
		return entity.Library{}, r.handleErr(err)
	}

	accessSQL, accessArgs, err := r.Builder.
		Select("user_id, role").
		From("library_accesses").
		Where(sq.Eq{"library_id": lib.ID}).
		ToSql()

	if err != nil {
		return entity.Library{}, d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	rows, err := r.Pool.Query(ctx, accessSQL, accessArgs...)
	if err != nil {
		return entity.Library{}, r.handleErr(err)
	}
	defer rows.Close()

	for rows.Next() {
		var access entity.LibraryAccess

		err = rows.Scan(&access.UserID, &access.Role)
		if err != nil {
			return entity.Library{}, r.handleErr(err)
		}

		lib.Accesses = append(lib.Accesses, access)
	}

	if err = rows.Err(); err != nil {
		return entity.Library{}, r.handleErr(err)
	}

	return lib, nil
}

// Update -.
func (r *LibRepo) Update(ctx context.Context, lib *entity.Library) error {
	tx := postgres.TxFromContext(ctx)
	ownsTx := false
	if tx == nil {
		var err error
		tx, err = r.Pool.Begin(ctx)
		if err != nil {
			return d.NewAppError(
				entity.CodeUndefined,
				"Internal error",
				d.KindInternal, nil)
		}
		ownsTx = true
		defer func() {
			_ = tx.Rollback(ctx)
		}()
	}

	sql, args, err := r.Builder.
		Update("libraries").
		Set("title", lib.Title).
		Set("description", lib.Description).
		Set("public", lib.Public).
		Set("deleted_at", lib.DeletedAt).
		Set("updated_at", lib.UpdatedAt).
		Where(sq.Eq{"id": lib.ID}).
		ToSql()
	if err != nil {
		return d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	res, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewAppError(
			entity.CodeLibNotFound,
			"Library not found",
			d.KindNotFound, nil)
	}

	deleteSQL, deleteArgs, err := r.Builder.
		Delete("library_accesses").
		Where(sq.Eq{"library_id": lib.ID}).
		ToSql()
	if err != nil {
		return d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	_, err = tx.Exec(ctx, deleteSQL, deleteArgs...)
	if err != nil {
		return r.handleErr(err)
	}

	if len(lib.Accesses) > 0 {
		insertBuilder := r.Builder.
			Insert("library_accesses").
			Columns("library_id", "user_id", "role")

		for _, access := range lib.Accesses {
			insertBuilder = insertBuilder.Values(lib.ID, access.UserID, access.Role)
		}

		insertSQL, insertArgs, err := insertBuilder.ToSql()
		if err != nil {
			return d.NewAppError(
				entity.CodeUndefined,
				"Internal error",
				d.KindInternal, nil)
		}

		_, err = tx.Exec(ctx, insertSQL, insertArgs...)
		if err != nil {
			return r.handleErr(err)
		}
	}

	if ownsTx {
		if err = tx.Commit(ctx); err != nil {
			return r.handleErr(err)
		}
	}

	return nil
}

func (r *LibRepo) Delete(ctx context.Context, userID, libID string) (int64, error) {
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

	now := time.Now().UTC()
	var deletedSize int64
	sizeSQL := `
SELECT COALESCE(SUM(size), 0)
FROM files
WHERE library_id = $1 AND deleted_at IS NULL
`
	if err := tx.QueryRow(ctx, sizeSQL, libID).Scan(&deletedSize); err != nil {
		return 0, r.handleErr(err)
	}

	foldersSQL := `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE library_id = $1 AND deleted_at IS NULL
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

	_, err := tx.Exec(ctx, foldersSQL, libID, now)
	if err != nil {
		return 0, r.handleErr(err)
	}

	filesSQL := `
UPDATE files
SET status = $2, deleted_at = $3, updated_at = $3
WHERE library_id = $1 AND deleted_at IS NULL
`
	_, err = tx.Exec(ctx, filesSQL, libID, entity.FileStatusDeleted, now)
	if err != nil {
		return 0, r.handleErr(err)
	}

	sql, args, err := r.Builder.Update("libraries").
		Set("deleted_at", now).
		Where(sq.Eq{"id": libID}).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return 0, d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}
	res, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return 0, r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return 0, d.NewAppError(
			entity.CodeLibNotFound,
			"Library not found",
			d.KindNotFound, nil)
	}
	if ownsTx {
		if err = tx.Commit(ctx); err != nil {
			return 0, r.handleErr(err)
		}
	}
	return deletedSize, nil
}

func (r *LibRepo) List(ctx context.Context, userID string) ([]entity.Library, error) {
	sql, args, err := r.Builder.
		Select("l.id, l.user_id, l.title, l.description, l.public, l.created_at, l.updated_at, l.deleted_at").
		From("libraries l").
		Where(sq.Eq{"l.deleted_at": nil}).
		Where(sq.Or{
			sq.Eq{"l.user_id": userID},
			sq.Expr("EXISTS (SELECT 1 FROM library_accesses la WHERE la.library_id = l.id AND la.user_id = ?)", userID),
		}).
		OrderBy("l.created_at DESC").
		ToSql()
	if err != nil {
		return nil, d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	libraries := make([]entity.Library, 0)
	libIDs := make([]string, 0)

	for rows.Next() {
		var lib entity.Library

		err = rows.Scan(
			&lib.ID,
			&lib.UserID,
			&lib.Title,
			&lib.Description,
			&lib.Public,
			&lib.CreatedAt,
			&lib.UpdatedAt,
			&lib.DeletedAt,
		)
		if err != nil {
			return nil, r.handleErr(err)
		}

		libIDs = append(libIDs, lib.ID)

		libraries = append(libraries, lib)
	}

	if err = rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	if len(libIDs) == 0 {
		return libraries, nil
	}

	accessSQL, accessArgs, err := r.Builder.
		Select("library_id, user_id, role").
		From("library_accesses").
		Where(sq.Eq{"library_id": libIDs}).
		ToSql()
	if err != nil {
		return nil, d.NewAppError(
			entity.CodeUndefined,
			"Internal error",
			d.KindInternal, nil)
	}

	accessRows, err := r.Pool.Query(ctx, accessSQL, accessArgs...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer accessRows.Close()

	libIndex := make(map[string]int, len(libIDs))
	for i, lib := range libraries {
		libIndex[lib.ID] = i
	}

	for accessRows.Next() {
		var libID string
		var access entity.LibraryAccess

		err = accessRows.Scan(&libID, &access.UserID, &access.Role)
		if err != nil {
			return nil, r.handleErr(err)
		}

		if idx, ok := libIndex[libID]; ok {
			libraries[idx].Accesses = append(libraries[idx].Accesses, access)
		}
	}

	if err = accessRows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return libraries, nil
}
