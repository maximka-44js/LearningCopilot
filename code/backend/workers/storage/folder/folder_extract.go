package folder

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

const extractJobKind = "folder.extract"

type ExtractArgs struct {
	FolderID string `json:"folder_id"`
	UserID   string `json:"user_id"`
}

func (ExtractArgs) Kind() string { return extractJobKind }

type ExtractWorker struct {
	river.WorkerDefaults[ExtractArgs]
	uc ExtractUseCase
}

func NewExtractWorker(uc ExtractUseCase) *ExtractWorker {
	return &ExtractWorker{uc: uc}
}

func (w *ExtractWorker) Work(ctx context.Context, job *river.Job[ExtractArgs]) error {
	if w == nil || w.uc == nil {
		return fmt.Errorf("folder extract worker: usecase is nil")
	}

	folderID := strings.TrimSpace(job.Args.FolderID)
	userID := strings.TrimSpace(job.Args.UserID)
	if folderID == "" || userID == "" {
		return fmt.Errorf("folder extract worker: missing folder or user id")
	}

	if _, err := uuid.Parse(folderID); err != nil {
		return fmt.Errorf("folder extract worker: invalid folder id: %w", err)
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("folder extract worker: invalid user id: %w", err)
	}

	if err := w.uc.ExtractFolder(ctx, userID, folderID); err != nil {
		return fmt.Errorf("folder extract worker: %w", err)
	}

	return nil
}
