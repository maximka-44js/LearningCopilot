package queue

import (
	"context"
	"time"
)

// Message carries a transport-agnostic async job payload.
type Message struct {
	Kind    string
	Payload any
	Options Options
}

// Options controls delivery behavior for queue backends.
type Options struct {
	ScheduledAt *time.Time
	MaxAttempts int
}

// Publisher abstracts job publishing away from specific brokers.
type Publisher interface {
	Publish(ctx context.Context, msg Message) error
}
