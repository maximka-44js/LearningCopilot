package repo

import (
	"context"
	"time"

	"github.com/maximka-44js/lc/internal/access/entity"
)

type APIKeyRepo interface {
	CountByUser(ctx context.Context, userID string) (int, error)
	Store(ctx context.Context, key *entity.APIKey) error
	ListByUser(ctx context.Context, userID string) ([]entity.APIKey, error)
	GetByIDForUser(ctx context.Context, userID, apiKeyID string) (entity.APIKey, error)
	GetByPrefix(ctx context.Context, keyPrefix string) (entity.APIKey, error)
	Update(ctx context.Context, key *entity.APIKey) error
	TouchLastUsed(ctx context.Context, apiKeyID string, usedAt time.Time) error
	StoreAudit(ctx context.Context, event APIKeyAuditEvent) error
}

type APIKeyAuditEvent struct {
	ID          string
	APIKeyID    *string
	ActorUserID *string
	EventType   string
	Meta        map[string]any
	CreatedAt   time.Time
}
