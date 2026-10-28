package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/maximka-44js/lc/pkg/queue"
)

const (
	fileDeletedJobKind = "file.deleted"
	fileToCompress     = "file.compress"
	fileToExtract      = "file.extract"
)

type FileJobs struct {
	queue queue.Publisher
}

func NewFileJobs(q queue.Publisher) *FileJobs {
	return &FileJobs{queue: q}
}

type FileDeletedArgs struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
}

func (FileDeletedArgs) Kind() string { return fileDeletedJobKind }

func (p *FileJobs) EnqueueFileDeleted(ctx context.Context, fileID, userID string) error {
	args := FileDeletedArgs{
		FileID: fileID,
		UserID: userID,
	}

	scheduledAt := time.Now().UTC().Add(2 * time.Second)
	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    fileDeletedJobKind,
		Payload: args,
		Options: queue.Options{ScheduledAt: &scheduledAt, MaxAttempts: 1},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueFileDeleted - Publish: %w", err)
	}

	return nil
}

type FileCompressArgs struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
}

func (FileCompressArgs) Kind() string { return fileToCompress }

func (p *FileJobs) EnqueueFileCompress(ctx context.Context, fileID, userID string) error {
	args := FileCompressArgs{
		FileID: fileID,
		UserID: userID,
	}

	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    fileToCompress,
		Payload: args,
		Options: queue.Options{MaxAttempts: 1},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueFileCompress - Publish: %w", err)
	}

	return nil
}

type FileExtractArgs struct {
	FileID string `json:"file_id"`
	UserID string `json:"user_id"`
}

func (FileExtractArgs) Kind() string { return fileToExtract }

func (p *FileJobs) EnqueueFileExtract(ctx context.Context, fileID, userID string) error {
	args := FileExtractArgs{
		FileID: fileID,
		UserID: userID,
	}

	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    fileToExtract,
		Payload: args,
		Options: queue.Options{MaxAttempts: 1},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueFileExtract - Publish: %w", err)
	}

	return nil
}
