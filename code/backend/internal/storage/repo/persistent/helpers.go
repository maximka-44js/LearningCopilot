package persistent

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/internal/storage/entity"
)

type RepoErrMapping struct {
	DuplicateCode   d.ErrorCode
	DuplicateMsg    string
	ReferenceCode   d.ErrorCode
	ReferenceMsg    string
	InvalidDataCode d.ErrorCode
	InvalidDataMsg  string
	NotFoundCode    d.ErrorCode
	NotFoundMsg     string
}

func isNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func isTimeoutOrCanceled(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

func handleRepoErr(err error, mapping RepoErrMapping) *d.AppError {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return d.NewConflict(
				fallbackCode(mapping.DuplicateCode, entity.CodeUndefined),
				fallbackMsg(mapping.DuplicateMsg, "Resource already exists"))
		}
		if pgErr.Code == "23503" {
			return d.NewValidationError(
				fallbackCode(mapping.ReferenceCode, entity.CodeUndefined),
				fallbackMsg(mapping.ReferenceMsg, "Referenced resource not found"))
		}
		if pgErr.Code == "23514" || pgErr.Code == "23502" || pgErr.Code == "22P02" {
			return d.NewValidationError(
				fallbackCode(mapping.InvalidDataCode, entity.CodeUndefined),
				fallbackMsg(mapping.InvalidDataMsg, "Invalid data"))
		}
	}

	if isTimeoutOrCanceled(err) {
		return d.NewAppError(entity.CodeUndefined, "Request timed out", d.KindTimeOut, nil)
	}

	if isNoRows(err) {
		return d.NewNotFound(
			fallbackCode(mapping.NotFoundCode, entity.CodeUndefined),
			fallbackMsg(mapping.NotFoundMsg, "Not found"))
	}
	return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
}

func fallbackCode(code d.ErrorCode, fallback d.ErrorCode) d.ErrorCode {
	if code == "" {
		return fallback
	}
	return code
}

func fallbackMsg(msg string, fallback string) string {
	if msg == "" {
		return fallback
	}
	return msg
}
