package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// Executor provides the shared query/exec surface for pool and tx.
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func contextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromContext returns a transaction if one is attached to the context.
func TxFromContext(ctx context.Context) pgx.Tx {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return nil
	}
	return tx
}

// ExecutorFromContext returns tx if present; otherwise pool.
func ExecutorFromContext(ctx context.Context, pool *pgxpool.Pool) Executor {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return pool
}

// WithinTx runs fn inside a transaction. Nested calls reuse the existing tx.
func (p *Postgres) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if TxFromContext(ctx) != nil {
		return fn(ctx)
	}

	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres - WithinTx - Begin: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctx = contextWithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres - WithinTx - Commit: %w", err)
	}

	return nil
}
