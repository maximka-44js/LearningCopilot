package river

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/maximka-44js/lc/pkg/postgres"
	"github.com/maximka-44js/lc/pkg/queue"
	"github.com/riverqueue/river"
)

// Publisher is a River-backed implementation of queue.Publisher.
type Publisher struct {
	client *river.Client[pgx.Tx]
}

var _ queue.Publisher = (*Publisher)(nil)

func NewPublisher(client *river.Client[pgx.Tx]) *Publisher {
	return &Publisher{client: client}
}

func (p *Publisher) Publish(ctx context.Context, msg queue.Message) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("queue river publisher: client is nil")
	}
	if msg.Kind == "" {
		return fmt.Errorf("queue river publisher: kind is required")
	}

	args := rawArgs{kind: msg.Kind, payload: msg.Payload}
	opts := toInsertOpts(msg.Options)

	if tx := postgres.TxFromContext(ctx); tx != nil {
		if _, err := p.client.InsertTx(ctx, tx, args, opts); err != nil {
			return fmt.Errorf("queue river publisher: insert tx: %w", err)
		}
		return nil
	}

	if _, err := p.client.Insert(ctx, args, opts); err != nil {
		return fmt.Errorf("queue river publisher: insert: %w", err)
	}

	return nil
}

type rawArgs struct {
	kind    string
	payload any
}

func (a rawArgs) Kind() string { return a.kind }

func (a rawArgs) MarshalJSON() ([]byte, error) {
	if a.payload == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(a.payload)
}

func toInsertOpts(opts queue.Options) *river.InsertOpts {
	insertOpts := &river.InsertOpts{}

	if opts.ScheduledAt != nil {
		insertOpts.ScheduledAt = opts.ScheduledAt.UTC()
	}
	if opts.MaxAttempts > 0 {
		insertOpts.MaxAttempts = opts.MaxAttempts
	}

	if insertOpts.ScheduledAt.IsZero() && insertOpts.MaxAttempts == 0 {
		return nil
	}

	return insertOpts
}
