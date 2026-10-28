package persistent

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maximka-44js/lc/internal/access/entity"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

func handleRepoErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return d.NewAppError(entity.CodeUndefined, "Request timed out", d.KindTimeOut, nil)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.NewNotFound(entity.CodeAPIKeyNotFound, "API key not found")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return d.NewConflict(entity.CodeAPIKeyInvalid, "API key already exists")
		case "23514", "23502", "22P02":
			return d.NewValidationError(entity.CodeAPIKeyInvalid, "API key data is invalid")
		}
	}

	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}
