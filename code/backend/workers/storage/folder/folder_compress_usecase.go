package folder

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
	storageentity "github.com/maximka-44js/lc/internal/storage/entity"
	storagerepo "github.com/maximka-44js/lc/internal/storage/repo"
)

const folderZipMimeType = "application/zip"

type CompressFolderReader interface {
	GetByID(ctx context.Context, userID, folderID string) (storageentity.Folder, error)
	ListByLibraryAndParent(ctx context.Context, userID, libID, parentID string) ([]storageentity.Folder, error)
}

type CompressLibraryReader interface {
	GetByID(ctx context.Context, userID, libID string) (storageentity.Library, error)
}

type CompressFileReader interface {
	GetByID(ctx context.Context, userID, fileID string) (storageentity.File, error)
	ListByLibraryAndFolder(ctx context.Context, userID, libID, folderID string) ([]storageentity.File, error)
	Store(ctx context.Context, file *storageentity.File) error
}

type CompressTxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type CompressUseCase interface {
	CompressFolder(ctx context.Context, userID, folderID string) error
}

type compressUseCase struct {
	libs    CompressLibraryReader
	folders CompressFolderReader
	files   CompressFileReader
	storage storagerepo.ObjectStorage
	reader  ObjectReader
	tx      CompressTxManager
}

func NewCompressUseCase(
	libs CompressLibraryReader,
	folders CompressFolderReader,
	files CompressFileReader,
	storage storagerepo.ObjectStorage,
	reader ObjectReader,
	tx CompressTxManager,
) CompressUseCase {
	return &compressUseCase{
		libs:    libs,
		folders: folders,
		files:   files,
		storage: storage,
		reader:  reader,
		tx:      tx,
	}
}

func (uc *compressUseCase) CompressFolder(ctx context.Context, userID, folderID string) error {
	if uc == nil {
		return fmt.Errorf("folder compress usecase: usecase is nil")
	}
	if uc.libs == nil || uc.folders == nil || uc.files == nil || uc.tx == nil {
		return fmt.Errorf("folder compress usecase: repositories are not configured")
	}
	if uc.storage == nil || uc.reader == nil {
		return fmt.Errorf("folder compress usecase: storage is not configured")
	}

	userID = strings.TrimSpace(userID)
	folderID = strings.TrimSpace(folderID)
	if userID == "" || folderID == "" {
		return fmt.Errorf("folder compress usecase: user_id and folder_id are required")
	}

	folder, err := uc.folders.GetByID(ctx, userID, folderID)
	if err != nil {
		return err
	}
	lib, err := uc.libs.GetByID(ctx, userID, folder.LibID)
	if err != nil {
		return err
	}
	if !lib.CanRead(userID) || !lib.CanWrite(userID) {
		return fmt.Errorf("folder compress usecase: access denied")
	}

	archiveName, err := uc.nextArchiveName(ctx, userID, folder)
	if err != nil {
		return err
	}

	archiveFileID := uuid.NewString()
	archiveObjectKey := buildFolderArchiveObjectKey(userID, folder.LibID, folder.ID, archiveFileID, archiveName)

	pr, pw := io.Pipe()
	zipErrCh := make(chan error, 1)

	go func() {
		zw := zip.NewWriter(pw)
		err := uc.writeFolderTreeToZip(ctx, zw, userID, folder.LibID, folder.ID, "")
		if err == nil {
			err = uc.writeSourceFileAtZipRootIfNeeded(ctx, zw, userID, folder)
		}
		if closeErr := zw.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		_ = pw.CloseWithError(err)
		zipErrCh <- err
	}()

	putErr := uc.storage.PutObject(ctx, storagerepo.PutObjectRequest{
		Key:      archiveObjectKey,
		MimeType: folderZipMimeType,
		Body:     pr,
	})
	zipErr := <-zipErrCh
	if putErr != nil {
		return fmt.Errorf("folder compress usecase: put zip object: %w", putErr)
	}
	if zipErr != nil {
		_ = uc.storage.Delete(context.Background(), archiveObjectKey)
		return fmt.Errorf("folder compress usecase: write zip stream: %w", zipErr)
	}

	head, err := uc.storage.Head(ctx, archiveObjectKey)
	if err != nil {
		_ = uc.storage.Delete(context.Background(), archiveObjectKey)
		return fmt.Errorf("folder compress usecase: head zip object: %w", err)
	}

	archiveFile, err := storageentity.NewFile(
		archiveFileID,
		userID,
		folder.LibID,
		folder.ParentID,
		archiveName,
		folderZipMimeType,
		head.Size,
		"",
	)
	if err != nil {
		_ = uc.storage.Delete(context.Background(), archiveObjectKey)
		return err
	}

	if err := archiveFile.ValidateStatusTransition(storageentity.FileStatusUploaded); err != nil {
		_ = uc.storage.Delete(context.Background(), archiveObjectKey)
		return err
	}

	archiveFile.Status = storageentity.FileStatusUploaded
	archiveFile.ObjectKey = archiveObjectKey
	archiveFile.Checksum = strings.TrimSpace(head.Checksum)
	archiveFile.UpdatedAt = time.Now().UTC()

	err = uc.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := uc.files.Store(txCtx, &archiveFile); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = uc.storage.Delete(context.Background(), archiveObjectKey)
		return err
	}

	return nil
}

func (uc *compressUseCase) writeFolderTreeToZip(ctx context.Context, zw *zip.Writer, userID, libraryID, parentFolderID, basePath string) error {
	folders, err := uc.folders.ListByLibraryAndParent(ctx, userID, libraryID, parentFolderID)
	if err != nil {
		return err
	}

	files, err := uc.files.ListByLibraryAndFolder(ctx, userID, libraryID, parentFolderID)
	if err != nil {
		return err
	}

	for _, folder := range folders {
		if err := ctx.Err(); err != nil {
			return err
		}

		folderPath := path.Join(basePath, strings.TrimSpace(folder.Name))
		if folderPath == "." {
			folderPath = ""
		}
		if folderPath != "" {
			head := &zip.FileHeader{
				Name:     folderPath + "/",
				Method:   zip.Store,
				Modified: folder.UpdatedAt,
			}
			if _, err := zw.CreateHeader(head); err != nil {
				return fmt.Errorf("folder compress usecase: create folder zip entry: %w", err)
			}
		}

		if err := uc.writeFolderTreeToZip(ctx, zw, userID, libraryID, folder.ID, folderPath); err != nil {
			return err
		}
	}

	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}

		objectKey := strings.TrimSpace(file.ObjectKey)
		if objectKey == "" {
			continue
		}

		entryName := path.Join(basePath, strings.TrimSpace(file.Name))
		if entryName == "" || entryName == "." {
			continue
		}

		if err := uc.writeObjectToZip(ctx, zw, objectKey, entryName, file.UpdatedAt); err != nil {
			return err
		}
	}

	return nil
}

func (uc *compressUseCase) writeSourceFileAtZipRootIfNeeded(ctx context.Context, zw *zip.Writer, userID string, folder storageentity.Folder) error {
	_ = ctx
	_ = zw
	_ = userID
	_ = folder
	return nil
}

func (uc *compressUseCase) writeObjectToZip(ctx context.Context, zw *zip.Writer, objectKey, entryName string, modified time.Time) error {
	r, err := uc.reader.Open(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("folder compress usecase: open object %s: %w", objectKey, err)
	}

	head := &zip.FileHeader{
		Name:     entryName,
		Method:   zip.Deflate,
		Modified: modified,
	}
	w, err := zw.CreateHeader(head)
	if err != nil {
		_ = r.Close()
		return fmt.Errorf("folder compress usecase: create file zip entry: %w", err)
	}

	if _, err := io.Copy(w, r); err != nil {
		_ = r.Close()
		return fmt.Errorf("folder compress usecase: write file zip entry: %w", err)
	}
	if err := r.Close(); err != nil {
		return fmt.Errorf("folder compress usecase: close source object: %w", err)
	}

	return nil
}

func (uc *compressUseCase) nextArchiveName(ctx context.Context, userID string, folder storageentity.Folder) (string, error) {
	baseName := strings.TrimSpace(folder.Name)
	if baseName == "" {
		baseName = "folder"
	}
	timestamp := time.Now().UTC().Format("20060102-150405")
	seed := baseName + "-archive-" + timestamp

	existing, err := uc.files.ListByLibraryAndFolder(ctx, userID, folder.LibID, folder.ParentID)
	if err != nil {
		return "", err
	}

	used := make(map[string]struct{}, len(existing))
	for _, f := range existing {
		used[f.Name] = struct{}{}
	}

	candidate := seed + ".zip"
	if _, ok := used[candidate]; !ok {
		return fitZipName(candidate), nil
	}

	idx := 2
	for {
		candidate = fmt.Sprintf("%s (%d).zip", seed, idx)
		if _, ok := used[candidate]; !ok {
			return fitZipName(candidate), nil
		}
		idx++
	}
}

func fitZipName(name string) string {
	const maxNameLength = 100
	r := []rune(name)
	if len(r) <= maxNameLength {
		return name
	}

	suffix := []rune(".zip")
	if len(suffix) >= maxNameLength {
		return string(suffix)
	}

	prefixLen := maxNameLength - len(suffix)
	return string(r[:prefixLen]) + string(suffix)
}

func buildFolderArchiveObjectKey(userID, libraryID, folderID, fileID, archiveName string) string {
	return "files/" + userID + "/" + libraryID + "/" + folderID + "/archives/" + fileID + "/" + archiveName
}
