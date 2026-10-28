package file

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/internal/storage/repo"
	usecase "github.com/maximka-44js/lc/internal/storage/usecase"
)

type UseCase struct {
	fileRepo   repo.FileRepo
	objStorage repo.ObjectStorage
	folderRepo repo.FolderRepo
	libRepo    repo.LibraryRepo
	jobs       FileJobPublisher
	tx         TxManager
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type FileJobPublisher interface {
	EnqueueFileDeleted(ctx context.Context, fileID, userID string) error
}

func NewFileUseCase(fr repo.FileRepo, os repo.ObjectStorage, f repo.FolderRepo, l repo.LibraryRepo, jobs FileJobPublisher, tx TxManager) *UseCase {
	return &UseCase{fileRepo: fr, objStorage: os, folderRepo: f, libRepo: l, jobs: jobs, tx: tx}
}

func (uc *UseCase) Get(ctx context.Context, userID, fileID string) (entity.File, error) {
	f, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return entity.File{}, err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, f.LibID)
	if err != nil {
		return entity.File{}, err
	}
	if !lib.CanRead(userID) {
		return entity.File{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}

	return f, nil
}

func (uc *UseCase) GetDownload(ctx context.Context, userID, fileID string) (usecase.DownloadFileResult, error) {
	if uc.objStorage == nil {
		return usecase.DownloadFileResult{}, d.NewValidationError(entity.CodeFileTypeInvalid, "object storage is not configured")
	}

	f, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return usecase.DownloadFileResult{}, err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, f.LibID)
	if err != nil {
		return usecase.DownloadFileResult{}, err
	}
	if !lib.CanRead(userID) {
		return usecase.DownloadFileResult{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}

	if f.Status == entity.FileStatusCreated || f.Status == entity.FileStatusDeleted {
		return usecase.DownloadFileResult{}, d.NewValidationError(entity.CodeFileStatusInvalid, "file is not available for download")
	}

	objectKey, err := requireObjectKey(f)
	if err != nil {
		return usecase.DownloadFileResult{}, err
	}

	if _, err := uc.objStorage.Head(ctx, objectKey); err != nil {
		if errors.Is(err, repo.ErrStorageNotFound) {
			return usecase.DownloadFileResult{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
		}
		return usecase.DownloadFileResult{}, err
	}

	url, err := uc.objStorage.PresignGet(ctx, objectKey)
	if err != nil {
		if errors.Is(err, repo.ErrStorageNotFound) {
			return usecase.DownloadFileResult{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
		}
		return usecase.DownloadFileResult{}, err
	}

	return usecase.DownloadFileResult{
		URL:       url,
		Name:      f.Name,
		MimeType:  f.MimeType,
		ObjectKey: objectKey,
	}, nil
}

func (uc *UseCase) Upload(ctx context.Context, userID string, input usecase.UploadFileInput) (entity.File, error) {
	if uc.objStorage == nil {
		return entity.File{}, d.NewValidationError(entity.CodeFileTypeInvalid, "object storage is not configured")
	}
	if input.Body == nil {
		return entity.File{}, d.NewValidationError(entity.CodeFileTypeInvalid, "file body is required")
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, input.LibraryID)
	if err != nil {
		return entity.File{}, err
	}
	if !lib.CanRead(userID) {
		return entity.File{}, d.NewNotFound(entity.CodeLibNotFound, "Library not found")
	} else if !lib.CanWrite(userID) {
		return entity.File{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "You don't have rights to upload files to this library")
	}

	if input.FolderID != "" {
		folder, err := uc.folderRepo.GetByID(ctx, userID, input.FolderID)
		if err != nil {
			return entity.File{}, err
		}
		if folder.DeletedAt != nil || folder.LibID != input.LibraryID {
			return entity.File{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
		}
	}

	id := uuid.NewString()
	f, err := entity.NewFile(id, userID, input.LibraryID, input.FolderID, input.Name, input.MimeType, input.Size, input.ChecksumDeclared)
	if err != nil {
		return entity.File{}, err
	}

	tempKey := buildTemporaryObjectKey(f.UserID, lib.ID, f.ID, f.Name)
	f.ObjectKey = tempKey

	if err := uc.objStorage.PutObject(ctx, repo.PutObjectRequest{
		Key:      tempKey,
		MimeType: input.MimeType,
		Size:     input.Size,
		Body:     input.Body,
	}); err != nil {
		return entity.File{}, err
	}

	head, err := uc.objStorage.Head(ctx, tempKey)
	if err != nil {
		_ = uc.objStorage.Delete(context.Background(), tempKey)
		return entity.File{}, err
	}

	if head.Size != f.Size {
		_ = uc.objStorage.Delete(context.Background(), tempKey)
		return entity.File{}, d.NewValidationError(entity.CodeFileSizeInvalid, "file size mismatch")
	}

	if strings.TrimSpace(f.ChecksumDeclared) != "" {
		if strings.TrimSpace(head.Checksum) == "" {
			_ = uc.objStorage.Delete(context.Background(), tempKey)
			return entity.File{}, d.NewValidationError(entity.CodeFileChecksumInvalid, "file checksum is missing")
		}
		if !strings.EqualFold(strings.TrimSpace(head.Checksum), strings.TrimSpace(f.ChecksumDeclared)) {
			_ = uc.objStorage.Delete(context.Background(), tempKey)
			return entity.File{}, d.NewValidationError(entity.CodeFileChecksumInvalid, "file checksum mismatch")
		}
	}

	f.Checksum = strings.TrimSpace(head.Checksum)
	if err := f.ValidateStatusTransition(entity.FileStatusUploaded); err != nil {
		_ = uc.objStorage.Delete(context.Background(), tempKey)
		return entity.File{}, err
	}
	f.Status = entity.FileStatusUploaded
	f.UpdatedAt = time.Now().UTC()

	if err := uc.attachOrPromoteObjectKey(ctx, &f, tempKey); err != nil {
		_ = uc.objStorage.Delete(context.Background(), tempKey)
		return entity.File{}, err
	}

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.fileRepo.Store(ctx, &f); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if f.ObjectKey == tempKey {
			_ = uc.objStorage.Delete(context.Background(), tempKey)
		}
		return entity.File{}, err
	}

	return f, nil
}

func (uc *UseCase) Update(ctx context.Context, userID, fileID, name, folderID string) (entity.File, error) {
	f, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return entity.File{}, err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, f.LibID)
	if err != nil {
		return entity.File{}, err
	}
	if !lib.CanWrite(userID) {
		return entity.File{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "You don't have rights to modify this file")
	} else if !lib.CanRead(userID) {
		return entity.File{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}

	f.Name = name
	f.FolderID = folderID
	f.UpdatedAt = time.Now().UTC()
	if err := f.Validate(); err != nil {
		return entity.File{}, err
	}

	if folderID != "" {
		folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
		if err != nil {
			return entity.File{}, err
		}
		if folder.DeletedAt != nil || folder.LibID != f.LibID {
			return entity.File{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
		}
	}

	if err := uc.fileRepo.UpdateMeta(ctx, &f); err != nil {
		return entity.File{}, err
	}

	return f, nil
}

func (uc *UseCase) Copy(ctx context.Context, userID, fileID, targetFolderID, name string) (entity.File, error) {
	srcFile, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return entity.File{}, err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, srcFile.LibID)
	if err != nil {
		return entity.File{}, err
	}
	if !lib.CanRead(userID) {
		return entity.File{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}
	if !lib.CanWrite(userID) {
		return entity.File{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "You don't have rights to copy files in this library")
	}

	targetFolderID = strings.TrimSpace(targetFolderID)
	if targetFolderID != "" {
		targetFolder, err := uc.folderRepo.GetByID(ctx, userID, targetFolderID)
		if err != nil {
			return entity.File{}, err
		}
		if targetFolder.DeletedAt != nil || targetFolder.LibID != srcFile.LibID {
			return entity.File{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
		}
	}

	copyName := strings.TrimSpace(name)
	if copyName == "" {
		copyName = srcFile.Name
	}

	copiedFile, err := entity.NewFile(
		uuid.NewString(),
		srcFile.UserID,
		srcFile.LibID,
		targetFolderID,
		copyName,
		srcFile.MimeType,
		srcFile.Size,
		srcFile.ChecksumDeclared,
	)
	if err != nil {
		return entity.File{}, err
	}

	if err := copiedFile.ValidateStatusTransition(entity.FileStatusUploaded); err != nil {
		return entity.File{}, err
	}
	copiedFile.Status = entity.FileStatusUploaded
	copiedFile.ScanAttempts = 0
	copiedFile.Checksum = srcFile.Checksum
	objectKey, err := requireObjectKey(srcFile)
	if err != nil {
		return entity.File{}, err
	}
	copiedFile.ObjectKey = objectKey
	copiedFile.UpdatedAt = time.Now().UTC()

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.fileRepo.Store(ctx, &copiedFile); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return entity.File{}, err
	}

	return copiedFile, nil
}

func (uc *UseCase) InitUpload(ctx context.Context, userID, libID, folderID, name, mimeType string, size int64, checksumDeclared string) (usecase.InitUploadResult, error) {
	// validate library
	lib, err := uc.libRepo.GetByID(ctx, userID, libID)
	if err != nil {
		return usecase.InitUploadResult{}, err
	}
	if !lib.CanRead(userID) {
		return usecase.InitUploadResult{}, d.NewNotFound(entity.CodeLibNotFound, "Library not found")
	} else if !lib.CanWrite(userID) {
		return usecase.InitUploadResult{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "You don't have rights to upload files to this library")
	}

	// validate folder if provided
	if folderID != "" {
		folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
		if err != nil {
			return usecase.InitUploadResult{}, err
		}
		if folder.DeletedAt != nil || folder.LibID != libID {
			return usecase.InitUploadResult{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
		}
	}

	// create file entity
	id := uuid.NewString()
	f, err := entity.NewFile(id, userID, libID, folderID, name, mimeType, size, checksumDeclared)
	if err != nil {
		return usecase.InitUploadResult{}, err
	}

	key := buildTemporaryObjectKey(f.UserID, lib.ID, f.ID, f.Name)
	f.ObjectKey = key

	var presignURL string
	var uploadID string

	// choose multipart for large files
	multipartThreshold := int64(15 * 1024 * 1024) // 15MB

	if uc.objStorage == nil {
		return usecase.InitUploadResult{}, d.NewValidationError(entity.CodeFileTypeInvalid, "object storage is not configured")
	}

	if f.Size > multipartThreshold {
		// create multipart upload first so client can start uploading parts
		res, err := uc.objStorage.CreateMultipart(ctx, repo.CreateMultipartRequest{Key: key, MimeType: f.MimeType, Size: f.Size})
		if err != nil {
			return usecase.InitUploadResult{}, err
		}
		uploadID = res.UploadID
	} else {
		pres, err := uc.objStorage.PresignPut(ctx, repo.PresignPutRequest{Key: key, MimeType: f.MimeType, Size: f.Size})
		if err != nil {
			return usecase.InitUploadResult{}, err
		}
		presignURL = pres.URL
	}

	// store file record inside tx
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.fileRepo.Store(ctx, &f); err != nil {
			return err
		}

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// attempt to abort multipart if it was created
		if uploadID != "" {
			_ = uc.objStorage.AbortMultipart(context.Background(), key, uploadID)
		}
		return usecase.InitUploadResult{}, err
	}

	// return result
	return usecase.InitUploadResult{File: f, PresignURL: presignURL, UploadID: uploadID}, nil
}

func (uc *UseCase) PresignUploadPart(ctx context.Context, userID, fileID, uploadID, checksumSHA256 string, partNumber int, size int64) (string, error) {
	f, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return "", err
	}

	// permission check via library
	lib, err := uc.libRepo.GetByID(ctx, userID, f.LibID)
	if err != nil {
		return "", err
	}
	if !lib.CanWrite(userID) {
		return "", d.NewForbidden(entity.CodeLibraryAccessInvalid, "access denied")
	} else if !lib.CanRead(userID) {
		return "", d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}

	key, err := requireObjectKey(f)
	if err != nil {
		return "", err
	}

	res, err := uc.objStorage.PresignUploadPart(ctx, repo.PresignUploadPartRequest{
		Key:            key,
		UploadID:       uploadID,
		PartNumber:     partNumber,
		ChecksumSHA256: checksumSHA256,
		Size:           size,
	})
	if err != nil {
		return "", err
	}
	return res.URL, nil
}

func (uc *UseCase) ConfirmUpload(ctx context.Context, userID, fileID, uploadID string, completedParts []repo.CompletedPart) (entity.File, error) {
	f, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return entity.File{}, err
	}
	// permission check
	lib, err := uc.libRepo.GetByID(ctx, userID, f.LibID)
	if err != nil {
		return entity.File{}, err
	}
	if !lib.CanWrite(userID) {
		return entity.File{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "access denied")
	} else if !lib.CanRead(userID) {
		return entity.File{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}

	if uc.objStorage == nil {
		return entity.File{}, d.NewValidationError(entity.CodeFileTypeInvalid, "object storage is not configured")
	}

	tempKey, err := requireObjectKey(f)
	if err != nil {
		return entity.File{}, err
	}

	if uploadID != "" {
		if len(completedParts) == 0 {
			return entity.File{}, d.NewValidationError(entity.CodeFileChecksumInvalid, "completed parts are required for multipart upload")
		}
		if err := uc.objStorage.CompleteMultipart(ctx, repo.CompleteMultipartRequest{
			Key:            tempKey,
			UploadID:       uploadID,
			CompletedParts: completedParts,
		}); err != nil {
			return entity.File{}, err
		}
	}

	// head object to verify presence and metadata
	head, err := uc.objStorage.Head(ctx, tempKey)
	if err != nil {
		return entity.File{}, err
	}

	//validate size
	if head.Size != f.Size {
		return entity.File{}, d.NewValidationError(entity.CodeFileSizeInvalid, "file size mismatch")
	}

	if strings.TrimSpace(f.ChecksumDeclared) != "" {
		if strings.TrimSpace(head.Checksum) == "" {
			return entity.File{}, d.NewValidationError(entity.CodeFileChecksumInvalid, "file checksum is missing")
		}
		if !strings.EqualFold(strings.TrimSpace(head.Checksum), strings.TrimSpace(f.ChecksumDeclared)) {
			return entity.File{}, d.NewValidationError(entity.CodeFileChecksumInvalid, "file checksum mismatch")
		}
	}

	f.Checksum = strings.TrimSpace(head.Checksum)

	// update status and persist + publish task
	if err := f.ValidateStatusTransition(entity.FileStatusUploaded); err != nil {
		return entity.File{}, err
	}
	f.Status = entity.FileStatusUploaded
	f.UpdatedAt = time.Now().UTC()

	if err := uc.attachOrPromoteObjectKey(ctx, &f, tempKey); err != nil {
		return entity.File{}, err
	}

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.fileRepo.Update(ctx, &f); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return entity.File{}, err
	}

	return f, nil
}

func (uc *UseCase) AbortUpload(ctx context.Context, userID, fileID, uploadID string) error {
	f, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, f.LibID)
	if err != nil {
		return err
	}
	if !lib.CanWrite(userID) {
		return d.NewForbidden(entity.CodeLibraryAccessInvalid, "access denied")
	} else if !lib.CanRead(userID) {
		return d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}

	if f.Status != entity.FileStatusCreated {
		return d.NewValidationError(entity.CodeFileStatusInvalid, "only unfinished uploads can be aborted")
	}

	key, err := requireObjectKey(f)
	if err != nil {
		return err
	}

	if err := uc.abortMultipartUploadsForKey(ctx, key, strings.TrimSpace(uploadID)); err != nil {
		return err
	}

	_ = uc.objStorage.Delete(context.Background(), key)

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := uc.fileRepo.Delete(ctx, userID, fileID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) abortMultipartUploadsForKey(ctx context.Context, key, uploadID string) error {
	if uploadID != "" {
		if err := uc.objStorage.AbortMultipart(ctx, key, uploadID); err != nil {
			return err
		}
	}

	uploads, err := uc.objStorage.ListMultipartUploads(ctx, repo.ListMultipartUploadsRequest{
		Prefix:    key,
		OlderThan: time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		return err
	}

	for _, up := range uploads {
		if strings.TrimSpace(up.UploadID) == "" {
			continue
		}
		if uploadID != "" && up.UploadID == uploadID {
			continue
		}
		if err := uc.objStorage.AbortMultipart(ctx, key, up.UploadID); err != nil {
			return err
		}
	}

	return nil
}

func (uc *UseCase) Delete(ctx context.Context, userID, fileID string) (entity.File, error) {
	f, err := uc.fileRepo.GetByID(ctx, userID, fileID)
	if err != nil {
		return entity.File{},
			err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, f.LibID)
	if err != nil {
		return entity.File{}, err
	}
	if !lib.CanWrite(userID) {
		return entity.File{}, d.NewForbidden(entity.CodeLibraryAccessInvalid, "access denied")
	} else if !lib.CanRead(userID) {
		return entity.File{}, d.NewNotFound(entity.CodeFileNotFound, "file not found")
	}

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		_, err := uc.fileRepo.Delete(ctx, userID, fileID)
		if err != nil {
			return err
		}
		if err := uc.jobs.EnqueueFileDeleted(ctx, f.ID, f.UserID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return entity.File{}, err
	}

	return f, nil
}

func buildTemporaryObjectKey(userID, libID, fileID, name string) string {
	return "files/" + userID + "/" + libID + "/.tmp/" + fileID + "/" + name
}

func buildChecksumObjectKey(checksum string) string {
	checksum = strings.TrimSpace(strings.ToLower(checksum))
	if checksum == "" {
		return ""
	}
	return "objects/sha256/" + checksum
}

func requireObjectKey(f entity.File) (string, error) {
	objectKey := strings.TrimSpace(f.ObjectKey)
	if objectKey == "" {
		return "", d.NewValidationError(entity.CodeFileStatusInvalid, "file object key is missing")
	}
	return objectKey, nil
}

func (uc *UseCase) attachOrPromoteObjectKey(ctx context.Context, f *entity.File, tempKey string) error {
	if strings.TrimSpace(f.Checksum) == "" {
		f.ObjectKey = tempKey
		return nil
	}

	dedupKey := buildChecksumObjectKey(f.Checksum)
	if dedupKey == "" {
		f.ObjectKey = tempKey
		return nil
	}

	dup, err := uc.fileRepo.FindActiveByChecksum(ctx, f.Checksum, f.ID)
	if err == nil {
		referenceKey, err := requireObjectKey(dup)
		if err != nil {
			return err
		}
		f.ObjectKey = referenceKey
		if referenceKey != tempKey {
			_ = uc.objStorage.Delete(context.Background(), tempKey)
		}
		return nil
	}
	if appErr, ok := err.(*d.AppError); !ok || appErr.Kind != d.KindNotFound {
		return err
	}

	if dedupKey != tempKey {
		if err := uc.objStorage.Copy(ctx, tempKey, dedupKey, f.MimeType); err != nil {
			return err
		}
		_ = uc.objStorage.Delete(context.Background(), tempKey)
	}

	f.ObjectKey = dedupKey
	return nil
}
