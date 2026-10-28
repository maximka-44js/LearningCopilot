package library

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/maximka-44js/lc/internal/storage/repo"
	"github.com/maximka-44js/lc/pkg/postgres"
	"github.com/riverqueue/river"
)

const deletedJobKind = "library.deleted"

// DeletedArgs describes a library-deleted job payload.
type DeletedArgs struct {
	LibraryID string `json:"library_id"`
	UserID    string `json:"user_id"`
}

// Kind implements river.JobArgs.
func (DeletedArgs) Kind() string { return deletedJobKind }

// DeletedWorker handles library deletion jobs.
type DeletedWorker struct {
	river.WorkerDefaults[DeletedArgs]
	pg *postgres.Postgres
	s3 repo.ObjectStorage
}

// NewDeletedWorker creates a worker with the required dependencies.
func NewDeletedWorker(pg *postgres.Postgres, storage repo.ObjectStorage) *DeletedWorker {
	return &DeletedWorker{pg: pg, s3: storage}
}

// Work executes the library deletion job.
func (w *DeletedWorker) Work(ctx context.Context, job *river.Job[DeletedArgs]) error {
	if w.pg == nil {
		return fmt.Errorf("library deleted worker: postgres is nil")
	}
	if w.s3 == nil {
		return fmt.Errorf("library deleted worker: object storage is nil")
	}
	libraryID := strings.TrimSpace(job.Args.LibraryID)
	userID := strings.TrimSpace(job.Args.UserID)
	if libraryID == "" || userID == "" {
		return fmt.Errorf("library deleted worker: missing library or user id")
	}
	if _, err := uuid.Parse(libraryID); err != nil {
		return fmt.Errorf("library deleted worker: invalid library id: %w", err)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("library deleted worker: invalid user id: %w", err)
	}

	objects, err := w.loadLibraryObjects(ctx, libraryID, userID)
	if err != nil {
		return err
	}

	fileIDs, err := w.loadLibraryDeletedFileIDs(ctx, libraryID, userID)
	if err != nil {
		return err
	}

	if err := w.purgeLibrary(ctx, libraryID, userID); err != nil {
		return err
	}

	for _, fileID := range fileIDs {
		tmpPrefix := buildTmpPrefix(userID, libraryID, fileID)
		if err := w.abortMultipartUploadsByPrefix(ctx, tmpPrefix); err != nil {
			return err
		}
		if err := w.s3.DeleteByPrefix(ctx, tmpPrefix); err != nil && !errors.Is(err, repo.ErrStorageNotFound) {
			return fmt.Errorf("library deleted worker: cleanup tmp prefix: %w", err)
		}
	}

	for _, objectKey := range objects {
		if strings.TrimSpace(objectKey) == "" {
			continue
		}
		if err := w.abortMultipartUploadsByPrefix(ctx, objectKey); err != nil {
			return err
		}
		refs, err := w.countActiveRefs(ctx, objectKey)
		if err != nil {
			return err
		}
		if refs == 0 {
			if err := w.s3.Delete(ctx, objectKey); err != nil {
				if errors.Is(err, repo.ErrStorageNotFound) {
					continue
				}
				return fmt.Errorf("library deleted worker: delete object: %w", err)
			}
		}
	}

	filesPrefix := buildLibraryFilesPrefix(userID, libraryID)
	if err := w.abortMultipartUploadsByPrefix(ctx, filesPrefix); err != nil {
		return err
	}
	if err := w.s3.DeleteByPrefix(ctx, filesPrefix); err != nil && !errors.Is(err, repo.ErrStorageNotFound) {
		return fmt.Errorf("library deleted worker: cleanup files prefix: %w", err)
	}

	return nil
}

func buildLibraryFilesPrefix(userID, libID string) string {
	return "files/" + userID + "/" + libID
}

func (w *DeletedWorker) loadLibraryDeletedFileIDs(ctx context.Context, libraryID, userID string) ([]string, error) {
	const loadLibraryFileIDsSQL = `
SELECT f.id::text
FROM files f
WHERE f.library_id = $1
	AND f.user_id = $2
	AND f.deleted_at IS NOT NULL
`

	rows, err := w.pg.Pool.Query(ctx, loadLibraryFileIDsSQL, libraryID, userID)
	if err != nil {
		return nil, fmt.Errorf("library deleted worker: load file ids: %w", err)
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("library deleted worker: scan file ids: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("library deleted worker: iterate file ids: %w", err)
	}

	return ids, nil
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
		return fmt.Errorf("library deleted worker: list multipart uploads: %w", err)
	}

	for _, up := range uploads {
		if strings.TrimSpace(up.UploadID) == "" {
			continue
		}
		if err := w.s3.AbortMultipart(ctx, up.Key, up.UploadID); err != nil {
			if errors.Is(err, repo.ErrStorageNotFound) {
				continue
			}
			return fmt.Errorf("library deleted worker: abort multipart: %w", err)
		}
	}

	return nil
}

func buildTmpPrefix(userID, libID, fileID string) string {
	return "files/" + userID + "/" + libID + "/.tmp/" + fileID
}

func (w *DeletedWorker) loadLibraryObjects(ctx context.Context, libraryID, userID string) ([]string, error) {
	const loadLibraryObjectsSQL = `
SELECT DISTINCT f.object_key
FROM files f
WHERE f.library_id = $1
	AND f.user_id = $2
	AND f.deleted_at IS NOT NULL
	AND f.object_key IS NOT NULL
	AND f.object_key <> ''
`

	rows, err := w.pg.Pool.Query(ctx, loadLibraryObjectsSQL, libraryID, userID)
	if err != nil {
		return nil, fmt.Errorf("library deleted worker: load objects: %w", err)
	}
	defer rows.Close()

	objects := make([]string, 0)
	for rows.Next() {
		var objectKey string
		if err := rows.Scan(&objectKey); err != nil {
			return nil, fmt.Errorf("library deleted worker: scan objects: %w", err)
		}
		objects = append(objects, objectKey)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("library deleted worker: iterate objects: %w", err)
	}

	return objects, nil
}

func (w *DeletedWorker) countActiveRefs(ctx context.Context, objectKey string) (int64, error) {
	const countRefsSQL = `
SELECT COUNT(1)
FROM files
WHERE COALESCE(object_key, '') = $1
	AND deleted_at IS NULL
`

	var count int64
	if err := w.pg.Pool.QueryRow(ctx, countRefsSQL, objectKey).Scan(&count); err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("library deleted worker: count refs: %w", err)
	}

	return count, nil
}

func (w *DeletedWorker) purgeLibrary(ctx context.Context, libraryID, userID string) error {
	return w.pg.WithinTx(ctx, func(ctx context.Context) error {
		exec := postgres.ExecutorFromContext(ctx, w.pg.Pool)

		const deleteFilesSQL = `
DELETE FROM files
WHERE library_id = $1 AND user_id = $2
`
		if _, err := exec.Exec(ctx, deleteFilesSQL, libraryID, userID); err != nil {
			return fmt.Errorf("library deleted worker: delete files: %w", err)
		}

		const deleteFoldersSQL = `
DELETE FROM folders
WHERE library_id = $1 AND user_id = $2
`
		if _, err := exec.Exec(ctx, deleteFoldersSQL, libraryID, userID); err != nil {
			return fmt.Errorf("library deleted worker: delete folders: %w", err)
		}

		const deleteAccessSQL = `
DELETE FROM library_accesses
WHERE library_id = $1
`
		if _, err := exec.Exec(ctx, deleteAccessSQL, libraryID); err != nil {
			return fmt.Errorf("library deleted worker: delete access: %w", err)
		}

		const deleteLibrarySQL = `
DELETE FROM libraries
WHERE id = $1 AND user_id = $2
`
		if _, err := exec.Exec(ctx, deleteLibrarySQL, libraryID, userID); err != nil {
			return fmt.Errorf("library deleted worker: delete library: %w", err)
		}

		return nil
	})
}
