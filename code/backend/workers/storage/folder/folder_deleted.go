package folder

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

const deletedJobKind = "folder.deleted"

type DeletedArgs struct {
	FolderID string `json:"folder_id"`
	UserID   string `json:"user_id"`
}

func (DeletedArgs) Kind() string { return deletedJobKind }

type DeletedWorker struct {
	river.WorkerDefaults[DeletedArgs]
	pg *postgres.Postgres
	s3 repo.ObjectStorage
}

func NewDeletedWorker(pg *postgres.Postgres, storage repo.ObjectStorage) *DeletedWorker {
	return &DeletedWorker{pg: pg, s3: storage}
}

// Work executes the folder deletion job.
func (w *DeletedWorker) Work(ctx context.Context, job *river.Job[DeletedArgs]) error {
	if w.pg == nil {
		return fmt.Errorf("folder deleted worker: postgres is nil")
	}
	if w.s3 == nil {
		return fmt.Errorf("folder deleted worker: object storage is nil")
	}
	folderID := strings.TrimSpace(job.Args.FolderID)
	userID := strings.TrimSpace(job.Args.UserID)
	if folderID == "" || userID == "" {
		return fmt.Errorf("folder deleted worker: missing folder or user id")
	}
	if _, err := uuid.Parse(folderID); err != nil {
		return fmt.Errorf("folder deleted worker: invalid folder id: %w", err)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("folder deleted worker: invalid user id: %w", err)
	}

	prefixes, err := w.loadFolderStoragePrefixes(ctx, folderID, userID)
	if err != nil {
		return err
	}

	objects, err := w.loadFolderObjects(ctx, folderID, userID)
	if err != nil {
		return err
	}

	fileRefs, err := w.loadFolderDeletedFileRefs(ctx, folderID, userID)
	if err != nil {
		return err
	}

	if err := w.purgeFolder(ctx, folderID, userID); err != nil {
		return err
	}

	for _, ref := range fileRefs {
		tmpPrefix := buildTmpPrefix(userID, ref.libID, ref.fileID)
		if err := w.abortMultipartUploadsByPrefix(ctx, tmpPrefix); err != nil {
			return err
		}
		if err := w.s3.DeleteByPrefix(ctx, tmpPrefix); err != nil && !errors.Is(err, repo.ErrStorageNotFound) {
			return fmt.Errorf("folder deleted worker: cleanup tmp prefix: %w", err)
		}
	}

	for _, object := range objects {
		if strings.TrimSpace(object.objectKey) == "" {
			continue
		}
		if err := w.abortMultipartUploadsByPrefix(ctx, object.objectKey); err != nil {
			return err
		}
		refs, err := w.countActiveRefs(ctx, object.objectKey)
		if err != nil {
			return err
		}
		if refs == 0 {
			if err := w.s3.Delete(ctx, object.objectKey); err != nil {
				if errors.Is(err, repo.ErrStorageNotFound) {
					continue
				}
				return fmt.Errorf("folder deleted worker: delete object: %w", err)
			}
		}
	}

	for _, prefix := range prefixes {
		if err := w.abortMultipartUploadsByPrefix(ctx, prefix); err != nil {
			return err
		}
		if err := w.s3.DeleteByPrefix(ctx, prefix); err != nil && !errors.Is(err, repo.ErrStorageNotFound) {
			return fmt.Errorf("folder deleted worker: cleanup folder prefix: %w", err)
		}
	}

	return nil
}

func (w *DeletedWorker) loadFolderStoragePrefixes(ctx context.Context, folderID, userID string) ([]string, error) {
	const loadFolderPrefixesSQL = `
WITH RECURSIVE folder_tree AS (
	SELECT id, library_id
	FROM folders
	WHERE id = $1 AND user_id = $2
	UNION ALL
	SELECT f.id, f.library_id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.user_id = $2
)
SELECT id::text, library_id::text
FROM folder_tree
`

	rows, err := w.pg.Pool.Query(ctx, loadFolderPrefixesSQL, folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("folder deleted worker: load storage prefixes: %w", err)
	}
	defer rows.Close()

	prefixes := make([]string, 0)
	for rows.Next() {
		var fdID string
		var libID string
		if err := rows.Scan(&fdID, &libID); err != nil {
			return nil, fmt.Errorf("folder deleted worker: scan storage prefixes: %w", err)
		}
		if strings.TrimSpace(fdID) == "" || strings.TrimSpace(libID) == "" {
			continue
		}
		prefixes = append(prefixes, buildFolderPrefix(userID, libID, fdID))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("folder deleted worker: iterate storage prefixes: %w", err)
	}

	return prefixes, nil
}

func buildFolderPrefix(userID, libID, folderID string) string {
	return "files/" + userID + "/" + libID + "/" + folderID
}

type deletedFileRef struct {
	libID  string
	fileID string
}

func (w *DeletedWorker) loadFolderDeletedFileRefs(ctx context.Context, folderID, userID string) ([]deletedFileRef, error) {
	const loadFolderFileIDsSQL = `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE id = $1 AND user_id = $2
	UNION ALL
	SELECT f.id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.user_id = $2
)
	SELECT fi.library_id::text, fi.id::text
FROM files fi
WHERE fi.folder_id IN (SELECT id FROM folder_tree)
  AND fi.user_id = $2
  AND fi.deleted_at IS NOT NULL
`

	rows, err := w.pg.Pool.Query(ctx, loadFolderFileIDsSQL, folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("folder deleted worker: load file ids: %w", err)
	}
	defer rows.Close()

	refs := make([]deletedFileRef, 0)
	for rows.Next() {
		var ref deletedFileRef
		if err := rows.Scan(&ref.libID, &ref.fileID); err != nil {
			return nil, fmt.Errorf("folder deleted worker: scan file ids: %w", err)
		}
		refs = append(refs, ref)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("folder deleted worker: iterate file ids: %w", err)
	}

	return refs, nil
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
		return fmt.Errorf("folder deleted worker: list multipart uploads: %w", err)
	}

	for _, up := range uploads {
		if strings.TrimSpace(up.UploadID) == "" {
			continue
		}
		if err := w.s3.AbortMultipart(ctx, up.Key, up.UploadID); err != nil {
			if errors.Is(err, repo.ErrStorageNotFound) {
				continue
			}
			return fmt.Errorf("folder deleted worker: abort multipart: %w", err)
		}
	}

	return nil
}

func buildTmpPrefix(userID, libID, fileID string) string {
	return "files/" + userID + "/" + libID + "/.tmp/" + fileID
}

type folderObject struct {
	objectKey string
}

func (w *DeletedWorker) loadFolderObjects(ctx context.Context, folderID, userID string) ([]folderObject, error) {
	const loadFolderObjectsSQL = `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE id = $1 AND user_id = $2
	UNION ALL
	SELECT f.id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.user_id = $2
)
SELECT DISTINCT fi.object_key
FROM files fi
WHERE fi.folder_id IN (SELECT id FROM folder_tree)
  AND fi.user_id = $2
  AND fi.deleted_at IS NOT NULL
	AND fi.object_key IS NOT NULL
	AND fi.object_key <> ''
`

	rows, err := w.pg.Pool.Query(ctx, loadFolderObjectsSQL, folderID, userID)
	if err != nil {
		return nil, fmt.Errorf("folder deleted worker: load objects: %w", err)
	}
	defer rows.Close()

	objects := make([]folderObject, 0)
	for rows.Next() {
		var entry folderObject
		if err := rows.Scan(&entry.objectKey); err != nil {
			return nil, fmt.Errorf("folder deleted worker: scan objects: %w", err)
		}
		objects = append(objects, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("folder deleted worker: iterate objects: %w", err)
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
		return 0, fmt.Errorf("folder deleted worker: count refs: %w", err)
	}

	return count, nil
}

func (w *DeletedWorker) purgeFolder(ctx context.Context, folderID, userID string) error {
	return w.pg.WithinTx(ctx, func(ctx context.Context) error {
		exec := postgres.ExecutorFromContext(ctx, w.pg.Pool)

		const deleteFilesSQL = `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE id = $1 AND user_id = $2
	UNION ALL
	SELECT f.id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.user_id = $2
)
DELETE FROM files
WHERE folder_id IN (SELECT id FROM folder_tree)
	AND user_id = $2
`
		if _, err := exec.Exec(ctx, deleteFilesSQL, folderID, userID); err != nil {
			return fmt.Errorf("folder deleted worker: delete files: %w", err)
		}

		const deleteFoldersSQL = `
WITH RECURSIVE folder_tree AS (
	SELECT id
	FROM folders
	WHERE id = $1 AND user_id = $2
	UNION ALL
	SELECT f.id
	FROM folders f
	JOIN folder_tree ft ON f.parent_id = ft.id
	WHERE f.user_id = $2
)
DELETE FROM folders
WHERE id IN (SELECT id FROM folder_tree)
	AND user_id = $2
`
		if _, err := exec.Exec(ctx, deleteFoldersSQL, folderID, userID); err != nil {
			return fmt.Errorf("folder deleted worker: delete folders: %w", err)
		}

		return nil
	})
}
