package publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/maximka-44js/lc/pkg/queue"
)

const (
	libraryDeletedJobKind  = "library.deleted"
	libraryCompressJobKind = "library.compress"
)

// LibraryJobs publishes library-related background jobs.
type LibraryJobs struct {
	queue queue.Publisher
}

// NewLibraryJobs creates a LibraryJobs publisher that enqueues River jobs.
func NewLibraryJobs(q queue.Publisher) *LibraryJobs {
	return &LibraryJobs{queue: q}
}

type LibraryDeletedArgs struct {
	LibraryID string `json:"library_id"`
	UserID    string `json:"user_id"`
}

func (LibraryDeletedArgs) Kind() string { return libraryDeletedJobKind }

func (p *LibraryJobs) EnqueueLibraryDeleted(ctx context.Context, libraryID, userID string) error {

	args := LibraryDeletedArgs{
		LibraryID: libraryID,
		UserID:    userID,
	}

	scheduledAt := time.Now().UTC().Add(2 * time.Second)
	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    libraryDeletedJobKind,
		Payload: args,
		Options: queue.Options{ScheduledAt: &scheduledAt, MaxAttempts: 3},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueLibraryDeleted - Publish: %w", err)
	}

	return nil
}

type LibraryCompressArgs struct {
	LibraryID string `json:"library_id"`
	UserID    string `json:"user_id"`
}

func (LibraryCompressArgs) Kind() string { return libraryCompressJobKind }

func (p *LibraryJobs) EnqueueLibraryCompress(ctx context.Context, libraryID, userID string) error {
	args := LibraryCompressArgs{
		LibraryID: libraryID,
		UserID:    userID,
	}

	if err := p.queue.Publish(ctx, queue.Message{
		Kind:    libraryCompressJobKind,
		Payload: args,
		Options: queue.Options{MaxAttempts: 1},
	}); err != nil {
		return fmt.Errorf("publisher - EnqueueLibraryCompress - Publish: %w", err)
	}

	return nil
}
