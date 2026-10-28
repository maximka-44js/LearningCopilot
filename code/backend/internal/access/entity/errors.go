package entity

import (
	d "github.com/maximka-44js/lc/internal/app/domain"
)

const (
	CodeUndefined         d.ErrorCode = "undefined"
	CodeInvalidID         d.ErrorCode = "invalid_uuid"
	CodeAPIKeyInvalid     d.ErrorCode = "api_key.invalid"
	CodeAPIKeyNotFound    d.ErrorCode = "api_key.not_found"
	CodeAPIKeyForbidden   d.ErrorCode = "api_key.forbidden"
	CodeAPIKeyExpired     d.ErrorCode = "api_key.expired"
	CodeAPIKeyRevoked     d.ErrorCode = "api_key.revoked"
	CodeAPIKeyLimit       d.ErrorCode = "api_key.limit_exceeded"
	CodeAPIKeyScopeDenied d.ErrorCode = "api_key.scope_denied"
)
