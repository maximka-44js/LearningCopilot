package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/maximka-44js/lc/pkg/queue"
)

const (
	folderDeletedJobKind  = "folder.deleted"
	folderCompressJobKind = "folder.compress"
	folderExtractJobKind  = "folder.extract"
)

type FolderJobs struct {
	queue queue.Publisher
}

func NewFolderJobs(q queue.Publisher) *FolderJobs {
	return &FolderJobs{queue: q}
}

type FolderDeletedArgs struct {
	FolderID string `json:"folder_id"`
	UserID   string `json:"user_id"`
}

func (FolderDeletedArgs) Kind() string { return folderDeletedJobKind }

func (p *FolderJobs) EnqueueFolderDeleted(ctx context.Context, folderID, userID string) error {

	args := FolderDeletedArgs{
		FolderID: folderID,
		UserID:   userID,
	}

	scheduledAt := time.Now().UTC().Add(2 * time.Second)
	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    folderDeletedJobKind,
		Payload: args,
		Options: queue.Options{ScheduledAt: &scheduledAt, MaxAttempts: 3},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueFolderDeleted - Publish: %w", err)
	}

	return nil
}

type FolderCompressArgs struct {
	FolderID string `json:"folder_id"`
	UserID   string `json:"user_id"`
}

func (FolderCompressArgs) Kind() string { return folderCompressJobKind }

func (p *FolderJobs) EnqueueFolderCompress(ctx context.Context, folderID, userID string) error {
	args := FolderCompressArgs{
		FolderID: folderID,
		UserID:   userID,
	}

	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    folderCompressJobKind,
		Payload: args,
		Options: queue.Options{MaxAttempts: 1},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueFolderCompress - Publish: %w", err)
	}

	return nil
}

type FolderExtractArgs struct {
	FolderID string `json:"folder_id"`
	UserID   string `json:"user_id"`
}

func (FolderExtractArgs) Kind() string { return folderExtractJobKind }

func (p *FolderJobs) EnqueueFolderExtract(ctx context.Context, folderID, userID string) error {
	args := FolderExtractArgs{
		FolderID: folderID,
		UserID:   userID,
	}

	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    folderExtractJobKind,
		Payload: args,
		Options: queue.Options{MaxAttempts: 1},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueFolderExtract - Publish: %w", err)
	}

	return nil
}
