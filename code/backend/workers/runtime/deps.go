package runtime

import (
	"context"

	"github.com/maximka-44js/lc/internal/storage/repo"
	"github.com/maximka-44js/lc/pkg/logger"
	"github.com/maximka-44js/lc/pkg/postgres"
)

// Deps groups shared worker dependencies.
type Deps struct {
	PG              *postgres.Postgres
	Log             logger.Interface
	ObjectStorage   repo.ObjectStorage
	LibraryCompress LibraryCompressUseCase
	FolderCompress  FolderCompressUseCase
	FolderExtract   FolderExtractUseCase
}

type LibraryCompressUseCase interface {
	CompressLibrary(ctx context.Context, userID, libraryID string) error
}

type FolderCompressUseCase interface {
	CompressFolder(ctx context.Context, userID, folderID string) error
}

type FolderExtractUseCase interface {
	ExtractFolder(ctx context.Context, userID, folderID string) error
}
