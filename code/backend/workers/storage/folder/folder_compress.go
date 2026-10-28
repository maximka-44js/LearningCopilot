package folder

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

const compressJobKind = "folder.compress"

type CompressArgs struct {
	FolderID string `json:"folder_id"`
	UserID   string `json:"user_id"`
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
		return fmt.Errorf("folder compress worker: usecase is nil")
	}

	folderID := strings.TrimSpace(job.Args.FolderID)
	userID := strings.TrimSpace(job.Args.UserID)
	if folderID == "" || userID == "" {
		return fmt.Errorf("folder compress worker: missing folder or user id")
	}
	if _, err := uuid.Parse(folderID); err != nil {
		return fmt.Errorf("folder compress worker: invalid folder id: %w", err)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("folder compress worker: invalid user id: %w", err)
	}

	if err := w.uc.CompressFolder(ctx, userID, folderID); err != nil {
		return fmt.Errorf("folder compress worker: %w", err)
	}

	return nil
}
