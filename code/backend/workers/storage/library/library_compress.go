package library

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

const compressJobKind = "library.compress"

type CompressArgs struct {
	LibraryID string `json:"library_id"`
	UserID    string `json:"user_id"`
}

func (CompressArgs) Kind() string { return compressJobKind }

type CompressWorker struct {
	river.WorkerDefaults[CompressArgs]
	uc CompressUseCase
}

func NewCompressWorker(uc CompressUseCase) *CompressWorker {
	return &CompressWorker{uc: uc}
}

func (w *CompressWorker) Work(ctx context.Context, job *river.Job[CompressArgs]) error {
	if w == nil || w.uc == nil {
		return fmt.Errorf("library compress worker: usecase is nil")
	}

	libraryID := strings.TrimSpace(job.Args.LibraryID)
	userID := strings.TrimSpace(job.Args.UserID)
	if libraryID == "" || userID == "" {
		return fmt.Errorf("library compress worker: missing library or user id")
	}
	if _, err := uuid.Parse(libraryID); err != nil {
		return fmt.Errorf("library compress worker: invalid library id: %w", err)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("library compress worker: invalid user id: %w", err)
	}

	if err := w.uc.CompressLibrary(ctx, userID, libraryID); err != nil {
		return fmt.Errorf("library compress worker: %w", err)
	}

	return nil
}
