package domain

type Kind string

const (
	KindValidation   Kind = "validation"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindInternal     Kind = "internal"
	KindTimeOut      Kind = "time_out"
	KindLimit        Kind = "limit_exceeded"
)

type ErrorCode string

type AppError struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Kind    Kind           `json:"kind"`
	Details map[string]any `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

func NewAppError(code ErrorCode, message string, kind Kind, details map[string]any) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Kind:    kind,
		Details: details,
	}
}

func NewNotFound(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Kind:    KindNotFound,
	}
}

func NewConflict(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Kind:    KindConflict,
	}
}

func NewValidationError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Kind:    KindValidation,
	}
}

func NewForbidden(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Kind:    KindForbidden,
	}
}

func NewUnauthorized(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Kind:    KindUnauthorized,
	}
}

func NewLimitExceeded(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Kind:    KindLimit,
	}
}

func (k Kind) HTTPCode() int {
	switch k {
	case KindValidation:
		return 400
	case KindNotFound:
		return 404
	case KindConflict:
		return 409
	case KindUnauthorized:
		return 401
	case KindForbidden:
		return 403
	case KindInternal:
		return 500
	case KindTimeOut:
		return 504
	case KindLimit:
		return 402
	default:
		return 500
	}
}
