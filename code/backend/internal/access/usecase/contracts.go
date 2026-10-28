package usecase

import (
	"context"
	"time"

	"github.com/maximka-44js/lc/internal/access/entity"
)

type APIKeys interface {
	Create(ctx context.Context, in CreateAPIKeyInput) (CreateAPIKeyOutput, error)
	List(ctx context.Context, userID string) ([]entity.APIKey, error)
	Revoke(ctx context.Context, in RevokeAPIKeyInput) (entity.APIKey, error)
	Rotate(ctx context.Context, in RotateAPIKeyInput) (CreateAPIKeyOutput, error)
	Authenticate(ctx context.Context, in AuthenticateAPIKeyInput) (AuthenticateAPIKeyOutput, error)
}

type CreateAPIKeyInput struct {
	UserID    string
	Name      string
	Scopes    []string
	ExpiresAt *time.Time
}

type RevokeAPIKeyInput struct {
	UserID   string
	APIKeyID string
}

type RotateAPIKeyInput struct {
	UserID   string
	APIKeyID string
	Name     *string
}

type AuthenticateAPIKeyInput struct {
	PresentedKey string
}

type CreateAPIKeyOutput struct {
	APIKey       entity.APIKey
	PlaintextKey string
}

type AuthenticateAPIKeyOutput struct {
	APIKey entity.APIKey
}
