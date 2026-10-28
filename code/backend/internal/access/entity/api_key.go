package entity

import (
	"strings"
	"time"

	d "github.com/maximka-44js/lc/internal/app/domain"
)

type APIKey struct {
	ID         string
	UserID     string
	Name       string
	KeyPrefix  string
	SecretHash string
	Scopes     []string
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (k *APIKey) Validate() error {
	if strings.TrimSpace(k.ID) == "" {
		return d.NewValidationError(CodeInvalidID, "api key id is required")
	}
	if strings.TrimSpace(k.UserID) == "" {
		return d.NewValidationError(CodeInvalidID, "user id is required")
	}
	if strings.TrimSpace(k.Name) == "" {
		return d.NewValidationError(CodeAPIKeyInvalid, "api key name is required")
	}
	if len(strings.TrimSpace(k.Name)) > 120 {
		return d.NewValidationError(CodeAPIKeyInvalid, "api key name is too long")
	}
	if strings.TrimSpace(k.KeyPrefix) == "" {
		return d.NewValidationError(CodeAPIKeyInvalid, "api key prefix is required")
	}
	if strings.TrimSpace(k.SecretHash) == "" {
		return d.NewValidationError(CodeAPIKeyInvalid, "api key hash is required")
	}
	if _, err := NormalizeScopes(k.Scopes); err != nil {
		return err
	}
	if k.ExpiresAt != nil && !k.ExpiresAt.After(k.CreatedAt.Add(-time.Second)) {
		return d.NewValidationError(CodeAPIKeyInvalid, "api key expiration must be in the future")
	}

	return nil
}

func (k APIKey) IsRevokedAt(now time.Time) bool {
	if k.RevokedAt == nil {
		return false
	}
	return !k.RevokedAt.After(now)
}

func (k APIKey) IsExpiredAt(now time.Time) bool {
	if k.ExpiresAt == nil {
		return false
	}
	return !k.ExpiresAt.After(now)
}
