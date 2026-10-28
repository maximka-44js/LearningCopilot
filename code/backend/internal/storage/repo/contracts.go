package repo

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/maximka-44js/lc/internal/storage/entity"
)

type (
	LibraryRepo interface {
		Store(ctx context.Context, library *entity.Library) error
		GetByID(ctx context.Context, userID, libID string) (entity.Library, error)
		Update(ctx context.Context, library *entity.Library) error
		Delete(ctx context.Context, userID, libID string) (int64, error)
		List(ctx context.Context, userID string) ([]entity.Library, error)
	}
	FolderRepo interface {
		Store(ctx context.Context, folder *entity.Folder) error
		GetByID(ctx context.Context, userID, folderID string) (entity.Folder, error)
		Update(ctx context.Context, folder *entity.Folder) error
		Delete(ctx context.Context, folderID string) (int64, error)
		ListByLibraryAndParent(ctx context.Context, userID, libID, parentID string) ([]entity.Folder, error)
		// GetByID(ctx context.Context, userID, folderID string) (entity.Folder, error)
	}
	FileRepo interface {
		Store(ctx context.Context, file *entity.File) error
		GetByID(ctx context.Context, userID, fileID string) (entity.File, error)
		FindActiveByChecksum(ctx context.Context, checksum, excludeFileID string) (entity.File, error)
		CountActiveByObjectKey(ctx context.Context, objectKey, excludeFileID string) (int64, error)
		Update(ctx context.Context, file *entity.File) error
		UpdateMeta(ctx context.Context, file *entity.File) error
		Delete(ctx context.Context, userID, fileID string) (entity.File, error)
		ListByFolder(ctx context.Context, userID string) ([]entity.File, error)
		ListByLibraryAndFolder(ctx context.Context, userID, libID, folderID string) ([]entity.File, error)
	}
	ObjectStorage interface {
		PutObject(ctx context.Context, req PutObjectRequest) error
		GetObject(ctx context.Context, key string) (io.ReadCloser, error)
		Copy(ctx context.Context, sourceKey, destinationKey, mimeType string) error
		PresignPut(ctx context.Context, req PresignPutRequest) (PresignPutResult, error)
		CreateMultipart(ctx context.Context, req CreateMultipartRequest) (CreateMultipartResult, error)
		PresignUploadPart(ctx context.Context, req PresignUploadPartRequest) (PresignUploadPartResult, error)
		CompleteMultipart(ctx context.Context, req CompleteMultipartRequest) error
		AbortMultipart(ctx context.Context, key, uploadID string) error
		ListMultipartUploads(ctx context.Context, req ListMultipartUploadsRequest) ([]MultipartUpload, error)
		Head(ctx context.Context, key string) (ObjectHead, error)
		PresignGet(ctx context.Context, key string) (string, error)
		Delete(ctx context.Context, key string) error
		DeleteByPrefix(ctx context.Context, prefix string) error
	}
)

var (
	ErrStorageNotFound     = errors.New("storage: not found")
	ErrStorageTemporary    = errors.New("storage: temporary error")
	ErrStorageNotSupported = errors.New("storage: not supported")
)

type PresignPutRequest struct {
	Key            string
	MimeType       string
	Size           int64
	ChecksumSHA256 string
}

type PutObjectRequest struct {
	Key      string
	MimeType string
	Size     int64
	Body     io.Reader
}

type PresignPutResult struct {
	URL string
}

type CreateMultipartRequest struct {
	Key      string
	MimeType string
	Size     int64
}

type CreateMultipartResult struct {
	UploadID string
}

type PresignUploadPartRequest struct {
	Key            string
	UploadID       string
	PartNumber     int
	ChecksumSHA256 string
	Size           int64
}

type PresignUploadPartResult struct {
	URL string
}

type CompletedPart struct {
	ETag       string
	PartNumber int
}

type CompleteMultipartRequest struct {
	Key            string
	UploadID       string
	CompletedParts []CompletedPart
}

type ListMultipartUploadsRequest struct {
	Prefix    string
	OlderThan time.Time
}

type MultipartUpload struct {
	Key         string
	UploadID    string
	InitiatedAt time.Time
}

type ObjectHead struct {
	Size         int64
	ETag         string
	LastModified time.Time
	ContentType  string
	Checksum     string
}
