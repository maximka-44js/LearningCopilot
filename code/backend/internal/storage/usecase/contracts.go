package usecase

import (
	"context"
	"io"

	"github.com/maximka-44js/lc/internal/storage/entity"
	"github.com/maximka-44js/lc/internal/storage/repo"
)

type (
	Library interface {
		Create(ctx context.Context, userID, title, description string, public bool, accesses []entity.LibraryAccess) (entity.Library, error)
		GetByID(ctx context.Context, userID, libID string) (entity.Library, error)
		List(ctx context.Context, userID string) ([]entity.Library, error)
		Update(ctx context.Context, userID, libID, title, description string, public bool, accesses []entity.LibraryAccess) (entity.Library, error)
		RequestCompress(ctx context.Context, userID, libID string) error
		Delete(ctx context.Context, userID, libID string) (entity.Library, error)
	}
	Folder interface {
		Create(ctx context.Context, userID, libID, parenID, name string) (entity.Folder, error)
		GetByID(ctx context.Context, userID, folderID string) (entity.Folder, error)
		Update(ctx context.Context, userID, folderID, name, parentID string) (entity.Folder, error)
		Copy(ctx context.Context, userID, folderID, targetParentID, name string) (entity.Folder, error)
		RequestCompress(ctx context.Context, userID, folderID string) error
		RequestExtract(ctx context.Context, userID, folderID string) error
		Delete(ctx context.Context, userID, folderID string) (entity.Folder, error)
		GetLibraryContents(ctx context.Context, userID, libID string) (TreeContents, error)
		GetFolderContents(ctx context.Context, userID, folderID string) (TreeContents, error)
		// List(ctx context.Context, userID, libID string) ([]entity.Folder, error)
		// Update(ctx context.Context, userID, folderID, name string) (entity.Folder, error)
		// Delete(ctx context.Context, userID, folderID string) error
	}
	// InitUploadResult - result of initializing an upload.

	File interface {
		Upload(ctx context.Context, userID string, input UploadFileInput) (entity.File, error)
		Get(ctx context.Context, userID, fileID string) (entity.File, error)
		GetDownload(ctx context.Context, userID, fileID string) (DownloadFileResult, error)
		Update(ctx context.Context, userID, fileID, name, folderID string) (entity.File, error)
		Copy(ctx context.Context, userID, fileID, targetFolderID, name string) (entity.File, error)
		InitUpload(ctx context.Context, userID, libID, folderID, name, mimeType string, size int64, checksumDeclared string) (InitUploadResult, error)
		PresignUploadPart(ctx context.Context, userID, fileID, uploadID, checksumSHA256 string, partNumber int, size int64) (string, error)
		ConfirmUpload(ctx context.Context, userID, fileID, uploadID string, completedParts []repo.CompletedPart) (entity.File, error)
		AbortUpload(ctx context.Context, userID, fileID, uploadID string) error
		Delete(ctx context.Context, userID, fileID string) (entity.File, error)
	}

	InitUploadResult struct {
		File       entity.File
		PresignURL string
		UploadID   string
	}

	UploadFileInput struct {
		LibraryID        string
		FolderID         string
		Name             string
		MimeType         string
		Size             int64
		ChecksumDeclared string
		Body             io.Reader
	}

	DownloadFileResult struct {
		URL       string
		Name      string
		MimeType  string
		ObjectKey string
	}

	TreeContents struct {
		LibraryID string          `json:"library_id"`
		ParentID  string          `json:"parent_id,omitempty"`
		Folders   []entity.Folder `json:"folders"`
		Files     []entity.File   `json:"files"`
	}
)
