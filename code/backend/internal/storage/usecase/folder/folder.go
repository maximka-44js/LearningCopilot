package folder

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/internal/storage/repo"
	usecase "github.com/maximka-44js/lc/internal/storage/usecase"
)

type UseCase struct {
	folderRepo repo.FolderRepo
	fileRepo   repo.FileRepo
	objStorage repo.ObjectStorage
	libRepo    repo.LibraryRepo
	jobs       FolderJobPublisher
	tx         TxManager
}

type FolderJobPublisher interface {
	EnqueueFolderDeleted(ctx context.Context, folderID, userID string) error
	EnqueueFolderCompress(ctx context.Context, folderID, userID string) error
	EnqueueFolderExtract(ctx context.Context, folderID, userID string) error
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

func NewFolderUseCase(r repo.FolderRepo, f repo.FileRepo, os repo.ObjectStorage, l repo.LibraryRepo, jobs FolderJobPublisher, tx TxManager) *UseCase {
	return &UseCase{folderRepo: r, fileRepo: f, objStorage: os, libRepo: l, jobs: jobs, tx: tx}
}

func (uc *UseCase) GetByID(ctx context.Context, userID, folderID string) (entity.Folder, error) {
	folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)

	if err != nil {
		return entity.Folder{}, err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, folder.LibID)
	if err != nil {
		return entity.Folder{}, err
	}

	if !lib.CanRead(userID) {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Library not found")
	}

	return folder, nil
}

func (uc *UseCase) Create(ctx context.Context, userID, libID, name, parentID string) (entity.Folder, error) {
	lib, err := uc.libRepo.GetByID(ctx, userID, libID)

	if err != nil {
		return entity.Folder{}, err
	}

	if !lib.CanRead(userID) {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Library not found")
	} else if !lib.CanWrite(userID) {
		return entity.Folder{}, d.NewForbidden(entity.CodeFolderAccessInvalid, "You don't have rights to modify this folder")
	}

	folder, err := entity.NewFolder(
		uuid.NewString(),
		userID,
		libID,
		name,
		parentID,
	)
	if err != nil {
		return entity.Folder{}, err
	}
	err = uc.folderRepo.Store(ctx, &folder)
	if err != nil {
		return entity.Folder{}, err
	}

	return folder, nil
}

// Update(ctx context.Context, userID, folderID, name, parentID string) (entity.Folder, error)
func (uc *UseCase) Update(ctx context.Context, userID, folderID, name, parentID string) (entity.Folder, error) {
	folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
	if err != nil {
		return entity.Folder{}, err
	}

	if folder.DeletedAt != nil {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, folder.LibID)
	if err != nil {
		return entity.Folder{}, err
	}

	if !lib.CanRead(userID) {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Library not found")
	} else if !lib.CanWrite(userID) {
		return entity.Folder{}, d.NewForbidden(entity.CodeFolderAccessInvalid, "You don't have rights to modify this folder")
	}

	folder.Name = name
	folder.ParentID = parentID
	folder.UpdatedAt = time.Now().UTC()

	if err := folder.Validate(); err != nil {
		return entity.Folder{}, err
	}

	if err := uc.folderRepo.Update(ctx, &folder); err != nil {
		return entity.Folder{}, err
	}

	return folder, nil
}

func (uc *UseCase) Copy(ctx context.Context, userID, folderID, targetParentID, name string) (entity.Folder, error) {
	srcFolder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
	if err != nil {
		return entity.Folder{}, err
	}
	if srcFolder.DeletedAt != nil {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, srcFolder.LibID)
	if err != nil {
		return entity.Folder{}, err
	}
	if !lib.CanRead(userID) {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Library not found")
	}
	if !lib.CanWrite(userID) {
		return entity.Folder{}, d.NewForbidden(entity.CodeFolderAccessInvalid, "You don't have rights to copy folders in this library")
	}

	targetParentID = strings.TrimSpace(targetParentID)
	if targetParentID != "" {
		targetParent, err := uc.folderRepo.GetByID(ctx, userID, targetParentID)
		if err != nil {
			return entity.Folder{}, err
		}
		if targetParent.DeletedAt != nil || targetParent.LibID != srcFolder.LibID {
			return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
		}
		if err := uc.validateCopyTargetParent(ctx, userID, srcFolder.ID, targetParent.ID); err != nil {
			return entity.Folder{}, err
		}
	}

	copyName := strings.TrimSpace(name)
	if copyName == "" {
		copyName = srcFolder.Name
	}

	var copiedRoot entity.Folder

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		root, _, err := uc.copyFolderTree(ctx, userID, srcFolder, targetParentID, copyName)
		if err != nil {
			return err
		}
		copiedRoot = root
		return nil
	})
	if err != nil {
		return entity.Folder{}, err
	}

	return copiedRoot, nil
}

func (uc *UseCase) Delete(ctx context.Context, userID, folderID string) (entity.Folder, error) {
	folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
	if err != nil {
		return entity.Folder{}, err
	}

	if folder.DeletedAt != nil {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, folder.LibID)
	if err != nil {
		return entity.Folder{}, err
	}

	if !lib.CanRead(userID) {
		return entity.Folder{}, d.NewNotFound(entity.CodeFolderNotFound, "Library not found")
	} else if !lib.CanWrite(userID) {
		return entity.Folder{}, d.NewForbidden(entity.CodeFolderAccessInvalid, "You don't have rights to modify this folder")
	}

	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if _, err := uc.folderRepo.Delete(ctx, folderID); err != nil {
			return err
		}
		if err := uc.jobs.EnqueueFolderDeleted(ctx, folderID, userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return entity.Folder{}, err
	}
	return folder, nil
}

func (uc *UseCase) RequestCompress(ctx context.Context, userID, folderID string) error {
	folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
	if err != nil {
		return err
	}

	if folder.DeletedAt != nil {
		return d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, folder.LibID)
	if err != nil {
		return err
	}

	if !lib.CanRead(userID) {
		return d.NewNotFound(entity.CodeFolderNotFound, "Library not found")
	} else if !lib.CanWrite(userID) {
		return d.NewForbidden(entity.CodeFolderAccessInvalid, "You don't have rights to modify this folder")
	}

	if err := uc.jobs.EnqueueFolderCompress(ctx, folder.ID, userID); err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) RequestExtract(ctx context.Context, userID, folderID string) error {
	folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
	if err != nil {
		return err
	}

	if folder.DeletedAt != nil {
		return d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, folder.LibID)
	if err != nil {
		return err
	}

	if !lib.CanRead(userID) {
		return d.NewNotFound(entity.CodeFolderNotFound, "Library not found")
	} else if !lib.CanWrite(userID) {
		return d.NewForbidden(entity.CodeFolderAccessInvalid, "You don't have rights to modify this folder")
	}

	if err := uc.jobs.EnqueueFolderExtract(ctx, folder.ID, userID); err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) GetLibraryContents(ctx context.Context, userID, libID string) (usecase.TreeContents, error) {
	lib, err := uc.libRepo.GetByID(ctx, userID, libID)
	if err != nil {
		return usecase.TreeContents{}, err
	}

	if !lib.CanRead(userID) {
		return usecase.TreeContents{}, d.NewNotFound(entity.CodeLibNotFound, "Library not found")
	}

	folders, err := uc.folderRepo.ListByLibraryAndParent(ctx, userID, libID, "")
	if err != nil {
		return usecase.TreeContents{}, err
	}

	files, err := uc.fileRepo.ListByLibraryAndFolder(ctx, userID, libID, "")
	if err != nil {
		return usecase.TreeContents{}, err
	}

	return usecase.TreeContents{
		LibraryID: libID,
		Folders:   folders,
		Files:     files,
	}, nil
}

func (uc *UseCase) GetFolderContents(ctx context.Context, userID, folderID string) (usecase.TreeContents, error) {
	folder, err := uc.folderRepo.GetByID(ctx, userID, folderID)
	if err != nil {
		return usecase.TreeContents{}, err
	}

	lib, err := uc.libRepo.GetByID(ctx, userID, folder.LibID)
	if err != nil {
		return usecase.TreeContents{}, err
	}

	if !lib.CanRead(userID) {
		return usecase.TreeContents{}, d.NewNotFound(entity.CodeFolderNotFound, "Folder not found")
	}

	folders, err := uc.folderRepo.ListByLibraryAndParent(ctx, userID, folder.LibID, folder.ID)
	if err != nil {
		return usecase.TreeContents{}, err
	}

	files, err := uc.fileRepo.ListByLibraryAndFolder(ctx, userID, folder.LibID, folder.ID)
	if err != nil {
		return usecase.TreeContents{}, err
	}

	return usecase.TreeContents{
		LibraryID: folder.LibID,
		ParentID:  folder.ID,
		Folders:   folders,
		Files:     files,
	}, nil
}

func (uc *UseCase) copyFolderTree(ctx context.Context, userID string, srcFolder entity.Folder, targetParentID, copyName string) (entity.Folder, int64, error) {
	copiedFolder, err := entity.NewFolder(uuid.NewString(), userID, srcFolder.LibID, copyName, targetParentID)
	if err != nil {
		return entity.Folder{}, 0, err
	}
	if err := uc.folderRepo.Store(ctx, &copiedFolder); err != nil {
		return entity.Folder{}, 0, err
	}

	totalBytes := int64(0)

	files, err := uc.fileRepo.ListByLibraryAndFolder(ctx, userID, srcFolder.LibID, srcFolder.ID)
	if err != nil {
		return entity.Folder{}, 0, err
	}
	for _, srcFile := range files {
		copiedBytes, err := uc.copyFileIntoFolder(ctx, userID, srcFile, copiedFolder.ID)
		if err != nil {
			return entity.Folder{}, 0, err
		}
		totalBytes += copiedBytes
	}

	children, err := uc.folderRepo.ListByLibraryAndParent(ctx, userID, srcFolder.LibID, srcFolder.ID)
	if err != nil {
		return entity.Folder{}, 0, err
	}
	for _, child := range children {
		_, copiedBytes, err := uc.copyFolderTree(ctx, userID, child, copiedFolder.ID, child.Name)
		if err != nil {
			return entity.Folder{}, 0, err
		}
		totalBytes += copiedBytes
	}

	return copiedFolder, totalBytes, nil
}

func (uc *UseCase) copyFileIntoFolder(ctx context.Context, userID string, srcFile entity.File, targetFolderID string) (int64, error) {
	copiedFile, err := entity.NewFile(
		uuid.NewString(),
		userID,
		srcFile.LibID,
		targetFolderID,
		srcFile.Name,
		srcFile.MimeType,
		srcFile.Size,
		srcFile.ChecksumDeclared,
	)
	if err != nil {
		return 0, err
	}
	if err := copiedFile.ValidateStatusTransition(entity.FileStatusUploaded); err != nil {
		return 0, err
	}
	copiedFile.Status = entity.FileStatusUploaded
	copiedFile.ScanAttempts = 0
	copiedFile.Checksum = srcFile.Checksum
	objectKey := strings.TrimSpace(srcFile.ObjectKey)
	if objectKey == "" {
		return 0, d.NewValidationError(entity.CodeFileStatusInvalid, "source file object key is missing")
	}
	copiedFile.ObjectKey = objectKey
	copiedFile.UpdatedAt = time.Now().UTC()

	if err := uc.fileRepo.Store(ctx, &copiedFile); err != nil {
		return 0, err
	}

	return copiedFile.Size, nil
}

func (uc *UseCase) measureFolderSize(ctx context.Context, userID, libID, folderID string) (int64, error) { //nolint:unused // Reserved for folder size calculation.
	files, err := uc.fileRepo.ListByLibraryAndFolder(ctx, userID, libID, folderID)
	if err != nil {
		return 0, err
	}

	total := int64(0)
	for _, f := range files {
		total += f.Size
	}

	children, err := uc.folderRepo.ListByLibraryAndParent(ctx, userID, libID, folderID)
	if err != nil {
		return 0, err
	}
	for _, child := range children {
		size, err := uc.measureFolderSize(ctx, userID, libID, child.ID)
		if err != nil {
			return 0, err
		}
		total += size
	}

	return total, nil
}

func (uc *UseCase) validateCopyTargetParent(ctx context.Context, userID, sourceFolderID, targetParentID string) error {
	if sourceFolderID == targetParentID {
		return d.NewValidationError(entity.CodeFolderAccessInvalid, "Target folder cannot be the same as source folder")
	}

	currentID := targetParentID
	for currentID != "" {
		folder, err := uc.folderRepo.GetByID(ctx, userID, currentID)
		if err != nil {
			return err
		}
		if folder.ID == sourceFolderID {
			return d.NewValidationError(entity.CodeFolderAccessInvalid, "Cannot copy folder into its own subtree")
		}
		currentID = strings.TrimSpace(folder.ParentID)
	}

	return nil
}
