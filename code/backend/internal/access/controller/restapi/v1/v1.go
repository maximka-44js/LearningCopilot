package v1

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/maximka-44js/lc/internal/access/usecase"
	"github.com/maximka-44js/lc/pkg/logger"
)

type V1 struct {
	apiKeys usecase.APIKeys
	l       logger.Interface
	v       *validator.Validate
}

type createAPIKeyRequest struct {
	Name      string     `json:"name" validate:"required,max=120"`
	Scopes    []string   `json:"scopes" validate:"required,min=1,dive,required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type rotateAPIKeyRequest struct {
	Name *string `json:"name"`
}

type apiKeyResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"created_at"`
}

type createAPIKeyResponse struct {
	APIKey apiKeyResponse `json:"api_key"`
	Secret string         `json:"secret"`
}
