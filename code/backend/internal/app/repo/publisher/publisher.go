package publisher

import (
	"context"
	"fmt"

	"github.com/maximka-44js/lc/internal/app/repo"
	"github.com/maximka-44js/lc/pkg/queue"
)

const appEventJobKind = "app.event"

// Publisher adapts domain events into River jobs.
type Publisher struct {
	queue queue.Publisher
}

var _ repo.EventPublisher = (*Publisher)(nil)

// New creates a Publisher that enqueues async event jobs.
func New(q queue.Publisher) *Publisher {
	return &Publisher{queue: q}
}

// EventJobArgs describes a domain event payload for async processing.
type EventJobArgs struct {
	ID            string `json:"id"`
	AggregateType string `json:"aggregate_type"`
	AggregateID   string `json:"aggregate_id"`
	EventType     string `json:"event_type"`
	Payload       []byte `json:"payload"`
	OccurredAt    string `json:"occurred_at"`
}

// Kind implements river.JobArgs.
// Store publishes the event using the configured queue transport.
func (p *Publisher) Store(ctx context.Context, msg *repo.Event) error {
	if p == nil || p.queue == nil {
		return fmt.Errorf("publisher - Store: queue publisher is nil")
	}
	if msg == nil {
		return fmt.Errorf("publisher - Store: message is nil")
	}

	args := EventJobArgs{
		ID:            msg.ID,
		AggregateType: msg.AggregateType,
		AggregateID:   msg.AggregateID,
		EventType:     msg.EventType,
		Payload:       msg.Payload,
		OccurredAt:    msg.OccurredAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
	}

	if err := p.queue.Publish(ctx, queue.Message{Kind: appEventJobKind, Payload: args}); err != nil {
		return fmt.Errorf("publisher - Store - Publish: %w", err)
	}

	return nil
}
