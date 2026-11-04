package repo

import (
	"context"
	"time"
)

// EventPublisher provides a minimal contract for publishing async events.
type EventPublisher interface {
	Store(ctx context.Context, msg *Event) error
}

// Event represents a domain event prepared for async processing.
type Event struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       []byte
	OccurredAt    time.Time
}
