package file

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/internal/storage/repo"
	"github.com/maximka-44js/lc/pkg/postgres"
	"github.com/riverqueue/river"
)

const deletedJobKind = "file.deleted"

type DeletedArgs struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
}

func (DeletedArgs) Kind() string { return deletedJobKind }

type DeletedWorker struct {
	river.WorkerDefaults[DeletedArgs]
	pg *postgres.Postgres
	s3 repo.ObjectStorage
}

type deletedFile struct {
	UserID    string
	LibID     string
	FolderID  string
	Name      string
	ObjectKey string
	Status    entity.FileStatus
	DeletedAt *time.Time
}

func NewDeletedWorker(pg *postgres.Postgres, storage repo.ObjectStorage) *DeletedWorker {
	return &DeletedWorker{pg: pg, s3: storage}
}

// Work executes the file deletion job.
func (w *DeletedWorker) Work(ctx context.Context, job *river.Job[DeletedArgs]) error {
	if w.pg == nil {
		return fmt.Errorf("file deleted worker: postgres is nil")
	}
	if w.s3 == nil {
		return fmt.Errorf("file deleted worker: object storage is nil")
	}
	fileID := strings.TrimSpace(job.Args.FileID)
	userID := strings.TrimSpace(job.Args.UserID)
	if fileID == "" || userID == "" {
		return fmt.Errorf("file deleted worker: missing file or user id")
	}
	if _, err := uuid.Parse(fileID); err != nil {
		return fmt.Errorf("file deleted worker: invalid file id: %w", err)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("file deleted worker: invalid user id: %w", err)
	}

	file, err := w.loadFile(ctx, fileID, userID)
	if err != nil {
		return err
	}
	if file == nil {
		return nil
	}
	if file.Status != entity.FileStatusDeleted || file.DeletedAt == nil {
		return fmt.Errorf("file deleted worker: file is not logically deleted")
	}

	objectKey := strings.TrimSpace(file.ObjectKey)
	tmpPrefix := buildTmpPrefix(file.UserID, file.LibID, fileID)
	if err := w.abortMultipartUploadsByPrefix(ctx, tmpPrefix); err != nil {
		return err
	}
	if objectKey != "" {
		if err := w.abortMultipartUploadsByPrefix(ctx, objectKey); err != nil {
			return err
		}
	}

	if err := w.s3.DeleteByPrefix(ctx, tmpPrefix); err != nil && !errors.Is(err, repo.ErrStorageNotFound) {
		return fmt.Errorf("file deleted worker: cleanup tmp prefix: %w", err)
	}

	if objectKey != "" {
		refs, err := w.countActiveRefs(ctx, objectKey, fileID)
		if err != nil {
			return err
		}
		if refs == 0 {
			if err := w.deleteObjectIdempotent(ctx, objectKey); err != nil {
				return err
			}
		}
	}

	if err := w.purgeFile(ctx, fileID, userID); err != nil {
		return err
	}

	return nil
}

func (w *DeletedWorker) abortMultipartUploadsByPrefix(ctx context.Context, prefix string) error {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return nil
	}

	uploads, err := w.s3.ListMultipartUploads(ctx, repo.ListMultipartUploadsRequest{
		Prefix:    prefix,
		OlderThan: time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		if errors.Is(err, repo.ErrStorageNotSupported) {
			return nil
		}
		return fmt.Errorf("file deleted worker: list multipart uploads: %w", err)
	}

	for _, up := range uploads {
		if strings.TrimSpace(up.UploadID) == "" {
			continue
		}
		if err := w.s3.AbortMultipart(ctx, up.Key, up.UploadID); err != nil {
			if errors.Is(err, repo.ErrStorageNotFound) {
				continue
			}
			return fmt.Errorf("file deleted worker: abort multipart: %w", err)
		}
	}

	return nil
}

func (w *DeletedWorker) deleteObjectIdempotent(ctx context.Context, key string) error {
	if err := w.s3.Delete(ctx, key); err != nil {
		if !errors.Is(err, repo.ErrStorageNotFound) {
			return fmt.Errorf("file deleted worker: delete object: %w", err)
		}
	}
	if err := w.s3.Delete(ctx, key+".tmp"); err != nil {
		if !errors.Is(err, repo.ErrStorageNotFound) {
			return fmt.Errorf("file deleted worker: delete temp object: %w", err)
		}
	}
	return nil
}

func buildTmpPrefix(userID, libID, fileID string) string {
	return "files/" + userID + "/" + libID + "/.tmp/" + fileID
}

func (w *DeletedWorker) loadFile(ctx context.Context, fileID, userID string) (*deletedFile, error) {
	const loadFileSQL = `
SELECT user_id, library_id, COALESCE(folder_id::text, ''), name, COALESCE(object_key, ''), status, deleted_at
FROM files
WHERE id = $1 AND user_id = $2
`
	row := w.pg.Pool.QueryRow(ctx, loadFileSQL, fileID, userID)

	var file deletedFile
	if err := row.Scan(&file.UserID, &file.LibID, &file.FolderID, &file.Name, &file.ObjectKey, &file.Status, &file.DeletedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("file deleted worker: load file: %w", err)
	}

	return &file, nil
}

func (w *DeletedWorker) countActiveRefs(ctx context.Context, objectKey, excludeFileID string) (int64, error) {
	const countRefsSQL = `
SELECT COUNT(1)
FROM files
WHERE COALESCE(object_key, '') = $1
  AND id <> $2
	AND status <> 'deleted'
  AND deleted_at IS NULL
`

	var count int64
	if err := w.pg.Pool.QueryRow(ctx, countRefsSQL, objectKey, excludeFileID).Scan(&count); err != nil {
		return 0, fmt.Errorf("file deleted worker: count refs: %w", err)
	}

	return count, nil
}

func (w *DeletedWorker) purgeFile(ctx context.Context, fileID, userID string) error {
	return w.pg.WithinTx(ctx, func(ctx context.Context) error {
		exec := postgres.ExecutorFromContext(ctx, w.pg.Pool)

		const deleteFileSQL = `
DELETE FROM files
WHERE id = $1
	AND user_id = $2
	AND (status = 'deleted' OR deleted_at IS NOT NULL)
`
		if _, err := exec.Exec(ctx, deleteFileSQL, fileID, userID); err != nil {
			return fmt.Errorf("file deleted worker: delete file: %w", err)
		}

		return nil
	})
}
