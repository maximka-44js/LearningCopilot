package library

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

const zipMimeType = "application/zip"

type LibraryReader interface {
	GetByID(ctx context.Context, userID, libID string) (storageentity.Library, error)
}

type FolderTreeReader interface {
	ListByLibraryAndParent(ctx context.Context, userID, libID, parentID string) ([]storageentity.Folder, error)
}

type FileTreeReader interface {
	ListByLibraryAndFolder(ctx context.Context, userID, libID, folderID string) ([]storageentity.File, error)
}

type FileRegistrar interface {
	Store(ctx context.Context, file *storageentity.File) error
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type CompressUseCase interface {
	CompressLibrary(ctx context.Context, userID, libraryID string) error
}

type compressUseCase struct {
	libraries LibraryReader
	folders   FolderTreeReader
	files     FileTreeReader
	register  FileRegistrar
	storage   storagerepo.ObjectStorage
	reader    ObjectReader
	tx        TxManager
}

func NewCompressUseCase(
	libraries LibraryReader,
	folders FolderTreeReader,
	files FileTreeReader,
	register FileRegistrar,
	storage storagerepo.ObjectStorage,
	reader ObjectReader,
	tx TxManager,
) CompressUseCase {
	return &compressUseCase{
		libraries: libraries,
		folders:   folders,
		files:     files,
		register:  register,
		storage:   storage,
		reader:    reader,
		tx:        tx,
	}
}

func (uc *compressUseCase) CompressLibrary(ctx context.Context, userID, libraryID string) error {
	if uc == nil {
		return fmt.Errorf("library compress usecase: usecase is nil")
	}
	if uc.libraries == nil || uc.folders == nil || uc.files == nil || uc.register == nil {
		return fmt.Errorf("library compress usecase: repositories are not configured")
	}
	if uc.storage == nil || uc.reader == nil {
		return fmt.Errorf("library compress usecase: object storage is not configured")
	}
	if uc.tx == nil {
		return fmt.Errorf("library compress usecase: tx manager is not configured")
	}

	userID = strings.TrimSpace(userID)
	libraryID = strings.TrimSpace(libraryID)
	if userID == "" || libraryID == "" {
		return fmt.Errorf("library compress usecase: user_id and library_id are required")
	}

	lib, err := uc.libraries.GetByID(ctx, userID, libraryID)
	if err != nil {
		return err
	}
	if !lib.CanRead(userID) || !lib.CanWrite(userID) {
		return fmt.Errorf("library compress usecase: access denied")
	}

	archiveFileID := uuid.NewString()
	archiveName := buildArchiveFileName(lib.Title)
	objectKey := buildArchiveObjectKey(userID, libraryID, archiveFileID, archiveName)

	pr, pw := io.Pipe()
	zipErrCh := make(chan error, 1)

	go func() {
		zw := zip.NewWriter(pw)
		err := uc.writeLibraryTreeToZip(ctx, zw, userID, libraryID, "", "")
		if closeErr := zw.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
		_ = pw.CloseWithError(err)
		zipErrCh <- err
	}()

	putErr := uc.storage.PutObject(ctx, storagerepo.PutObjectRequest{
		Key:      objectKey,
		MimeType: zipMimeType,
		Body:     pr,
	})
	zipErr := <-zipErrCh
	if putErr != nil {
		return fmt.Errorf("library compress usecase: put zip object: %w", putErr)
	}
	if zipErr != nil {
		_ = uc.storage.Delete(context.Background(), objectKey)
		return fmt.Errorf("library compress usecase: write zip stream: %w", zipErr)
	}

	head, err := uc.storage.Head(ctx, objectKey)
	if err != nil {
		_ = uc.storage.Delete(context.Background(), objectKey)
		return fmt.Errorf("library compress usecase: head zip object: %w", err)
	}

	archiveFile, err := storageentity.NewFile(
		archiveFileID,
		userID,
		libraryID,
		"",
		archiveName,
		zipMimeType,
		head.Size,
		"",
	)
	if err != nil {
		_ = uc.storage.Delete(context.Background(), objectKey)
		return err
	}

	if err := archiveFile.ValidateStatusTransition(storageentity.FileStatusUploaded); err != nil {
		_ = uc.storage.Delete(context.Background(), objectKey)
		return err
	}

	archiveFile.Status = storageentity.FileStatusUploaded
	archiveFile.ObjectKey = objectKey
	archiveFile.Checksum = strings.TrimSpace(head.Checksum)
	archiveFile.UpdatedAt = time.Now().UTC()

	err = uc.tx.WithinTx(ctx, func(txCtx context.Context) error {
		if err := uc.register.Store(txCtx, &archiveFile); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = uc.storage.Delete(context.Background(), objectKey)
		return err
	}

	return nil
}

func (uc *compressUseCase) writeLibraryTreeToZip(ctx context.Context, zw *zip.Writer, userID, libraryID, parentFolderID, basePath string) error {
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
				return fmt.Errorf("library compress usecase: create folder zip entry: %w", err)
			}
		}

		if err := uc.writeLibraryTreeToZip(ctx, zw, userID, libraryID, folder.ID, folderPath); err != nil {
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
		if entryName == "." || entryName == "" {
			continue
		}

		reader, err := uc.reader.Open(ctx, objectKey)
		if err != nil {
			return fmt.Errorf("library compress usecase: open object %s: %w", objectKey, err)
		}

		head := &zip.FileHeader{
			Name:     entryName,
			Method:   zip.Deflate,
			Modified: file.UpdatedAt,
		}
		writer, err := zw.CreateHeader(head)
		if err != nil {
			_ = reader.Close()
			return fmt.Errorf("library compress usecase: create file zip entry: %w", err)
		}

		if _, err := io.Copy(writer, reader); err != nil {
			_ = reader.Close()
			return fmt.Errorf("library compress usecase: write file zip entry: %w", err)
		}
		if err := reader.Close(); err != nil {
			return fmt.Errorf("library compress usecase: close source object: %w", err)
		}
	}

	return nil
}

func buildArchiveObjectKey(userID, libraryID, fileID, archiveName string) string {
	return "files/" + userID + "/" + libraryID + "/archives/" + fileID + "/" + archiveName
}

func buildArchiveFileName(title string) string {
	base := strings.TrimSpace(title)
	if base == "" {
		base = "library"
	}

	timestamp := time.Now().UTC().Format("20060102-150405")
	name := base + "-archive-" + timestamp + ".zip"

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
