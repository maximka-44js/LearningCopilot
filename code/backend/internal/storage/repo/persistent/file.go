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

type FileRepo struct {
	*postgres.Postgres
}

func NewFileRepo(pg *postgres.Postgres) *FileRepo {
	return &FileRepo{pg}
}

func (r *FileRepo) handleErr(err error) *d.AppError {
	return handleRepoErr(err, RepoErrMapping{
		DuplicateCode:   entity.CodeUndefined,
		DuplicateMsg:    "File already exists",
		ReferenceCode:   entity.CodeUndefined,
		ReferenceMsg:    "Referenced resource not found",
		InvalidDataCode: entity.CodeUndefined,
		InvalidDataMsg:  "Invalid data",
		NotFoundCode:    entity.CodeFileNotFound,
		NotFoundMsg:     "File not found",
	})
}

// GetByID -.
func (r *FileRepo) GetByID(ctx context.Context, userID, fileID string) (entity.File, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Select("id", "user_id", "library_id", "COALESCE(folder_id::text, '')", "name", "size", "mime_type", "COALESCE(object_key, '')", "checksum_declared", "checksum", "status", "scan_attempts", "created_at", "updated_at", "deleted_at").
		From("files").
		Where(sq.Eq{"id": fileID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()

	if err != nil {
		return entity.File{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var f entity.File
	err = exec.QueryRow(ctx, sql, args...).Scan(
		&f.ID,
		&f.UserID,
		&f.LibID,
		&f.FolderID,
		&f.Name,
		&f.Size,
		&f.MimeType,
		&f.ObjectKey,
		&f.ChecksumDeclared,
		&f.Checksum,
		&f.Status,
		&f.ScanAttempts,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.DeletedAt,
	)
	if err != nil {
		return entity.File{}, r.handleErr(err)
	}
	return f, nil
}

func (r *FileRepo) FindActiveByChecksum(ctx context.Context, checksum, excludeFileID string) (entity.File, error) {
	checksum = strings.TrimSpace(checksum)
	if checksum == "" {
		return entity.File{}, d.NewAppError(entity.CodeFileChecksumInvalid, "File checksum is required", d.KindValidation, nil)
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	query := r.Builder.
		Select("id", "user_id", "library_id", "COALESCE(folder_id::text, '')", "name", "size", "mime_type", "COALESCE(object_key, '')", "checksum_declared", "checksum", "status", "scan_attempts", "created_at", "updated_at", "deleted_at").
		From("files").
		Where(sq.Eq{"checksum": checksum}).
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.NotEq{"status": entity.FileStatusCreated}).
		OrderBy("updated_at ASC").
		Limit(1)

	if strings.TrimSpace(excludeFileID) != "" {
		query = query.Where(sq.NotEq{"id": excludeFileID})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return entity.File{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var f entity.File
	err = exec.QueryRow(ctx, sql, args...).Scan(
		&f.ID,
		&f.UserID,
		&f.LibID,
		&f.FolderID,
		&f.Name,
		&f.Size,
		&f.MimeType,
		&f.ObjectKey,
		&f.ChecksumDeclared,
		&f.Checksum,
		&f.Status,
		&f.ScanAttempts,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.DeletedAt,
	)
	if err != nil {
		return entity.File{}, r.handleErr(err)
	}
	return f, nil
}

func (r *FileRepo) CountActiveByObjectKey(ctx context.Context, objectKey, excludeFileID string) (int64, error) {
	objectKey = strings.TrimSpace(objectKey)
	if objectKey == "" {
		return 0, nil
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	query := r.Builder.
		Select("COUNT(1)").
		From("files").
		Where(sq.Eq{"object_key": objectKey}).
		Where(sq.Eq{"deleted_at": nil})

	if strings.TrimSpace(excludeFileID) != "" {
		query = query.Where(sq.NotEq{"id": excludeFileID})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var count int64
	if err := exec.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, r.handleErr(err)
	}

	return count, nil
}

// Store - insert file record
func (r *FileRepo) Store(ctx context.Context, f *entity.File) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	var folderID interface{}
	if f.FolderID == "" {
		folderID = nil
	} else {
		folderID = f.FolderID
	}

	sql, args, err := r.Builder.
		Insert("files").
		Columns("id", "user_id", "library_id", "folder_id", "name", "size", "mime_type", "object_key", "checksum_declared", "checksum", "status", "scan_attempts", "created_at", "updated_at").
		Values(f.ID, f.UserID, f.LibID, folderID, f.Name, f.Size, f.MimeType, f.ObjectKey, f.ChecksumDeclared, f.Checksum, f.Status, f.ScanAttempts, f.CreatedAt, f.UpdatedAt).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	_, err = exec.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	return nil
}

// Update - update file record
func (r *FileRepo) Update(ctx context.Context, f *entity.File) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Update("files").
		Set("folder_id", func() interface{} {
			if strings.TrimSpace(f.FolderID) == "" {
				return nil
			}
			return f.FolderID
		}()).
		Set("name", f.Name).
		Set("object_key", f.ObjectKey).
		Set("checksum", f.Checksum).
		Set("status", f.Status).
		Set("scan_attempts", f.ScanAttempts).
		Set("deleted_at", f.DeletedAt).
		Set("updated_at", f.UpdatedAt).
		Where(sq.Eq{"id": f.ID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()

	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewAppError(entity.CodeFileNotFound, "File not found", d.KindNotFound, nil)
	}
	return nil
}

// UpdateMeta - update file name and folder
func (r *FileRepo) UpdateMeta(ctx context.Context, f *entity.File) error {
	return r.Update(ctx, f)
}

// Delete - mark file as deleted and return file
func (r *FileRepo) Delete(ctx context.Context, userID, fileID string) (entity.File, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	// get file
	f, err := r.GetByID(ctx, userID, fileID)
	if err != nil {
		return entity.File{}, err
	}

	now := time.Now().UTC()
	f.Status = entity.FileStatusDeleted
	f.DeletedAt = &now
	f.UpdatedAt = now

	sql, args, err := r.Builder.
		Update("files").
		Set("status", f.Status).
		Set("deleted_at", f.DeletedAt).
		Set("updated_at", f.UpdatedAt).
		Where(sq.Eq{"id": f.ID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return entity.File{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return entity.File{}, r.handleErr(err)
	}
	if res.RowsAffected() == 0 {
		return entity.File{}, d.NewAppError(entity.CodeFileNotFound, "File not found", d.KindNotFound, nil)
	}
	return f, nil
}

// ListByFolder - list files for user
func (r *FileRepo) ListByFolder(ctx context.Context, userID string) ([]entity.File, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Select("id", "user_id", "library_id", "COALESCE(folder_id::text, '')", "name", "size", "mime_type", "COALESCE(object_key, '')", "checksum_declared", "checksum", "status", "scan_attempts", "created_at", "updated_at", "deleted_at").
		From("files").
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, r.handleErr(err)
	}
	defer rows.Close()

	var res []entity.File
	for rows.Next() {
		var f entity.File
		if err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.LibID,
			&f.FolderID,
			&f.Name,
			&f.Size,
			&f.MimeType,
			&f.ObjectKey,
			&f.ChecksumDeclared,
			&f.Checksum,
			&f.Status,
			&f.ScanAttempts,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.DeletedAt,
		); err != nil {
			return nil, r.handleErr(err)
		}
		res = append(res, f)
	}
	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}
	return res, nil
}

func (r *FileRepo) ListByLibraryAndFolder(ctx context.Context, _ string, libID, folderID string) ([]entity.File, error) {
	libID = strings.TrimSpace(libID)
	folderID = strings.TrimSpace(folderID)
	if libID == "" {
		return nil, d.NewAppError(entity.CodeLibNotFound, "Library not found", d.KindNotFound, nil)
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	query := r.Builder.
		Select("id", "user_id", "library_id", "COALESCE(folder_id::text, '')", "name", "size", "mime_type", "COALESCE(object_key, '')", "checksum_declared", "checksum", "status", "scan_attempts", "created_at", "updated_at", "deleted_at").
		From("files").
		Where(sq.Eq{"library_id": libID}).
		Where(sq.Eq{"deleted_at": nil})

	if folderID == "" {
		query = query.Where(sq.Eq{"folder_id": nil})
	} else {
		query = query.Where(sq.Eq{"folder_id": folderID})
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

	res := make([]entity.File, 0)
	for rows.Next() {
		var f entity.File
		if err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.LibID,
			&f.FolderID,
			&f.Name,
			&f.Size,
			&f.MimeType,
			&f.ObjectKey,
			&f.ChecksumDeclared,
			&f.Checksum,
			&f.Status,
			&f.ScanAttempts,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.DeletedAt,
		); err != nil {
			return nil, r.handleErr(err)
		}
		res = append(res, f)
	}
	if err := rows.Err(); err != nil {
		return nil, r.handleErr(err)
	}

	return res, nil
}
