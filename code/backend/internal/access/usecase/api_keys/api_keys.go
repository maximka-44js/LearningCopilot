package api_keys

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maximka-44js/lc/internal/access/entity"
	"github.com/maximka-44js/lc/internal/access/repo"
	"github.com/maximka-44js/lc/internal/access/usecase"
	d "github.com/maximka-44js/lc/internal/app/domain"
)

const (
	apiKeyPrefix             = "flk_"
	apiKeyIDBytes            = 8
	apiKeySecretBytes        = 32
	defaultMaxKeysPerUser    = 20
	defaultRotateGraceWindow = 24 * time.Hour
)

type Config struct {
	Enabled           bool
	HashPepper        string
	MaxKeysPerUser    int
	RotateGraceWindow time.Duration
}

type UseCase struct {
	repo repo.APIKeyRepo
	cfg  Config
}

var _ usecase.APIKeys = (*UseCase)(nil)

func NewUseCase(r repo.APIKeyRepo, cfg Config) *UseCase {
	if cfg.MaxKeysPerUser <= 0 {
		cfg.MaxKeysPerUser = defaultMaxKeysPerUser
	}
	if cfg.RotateGraceWindow <= 0 {
		cfg.RotateGraceWindow = defaultRotateGraceWindow
	}

	return &UseCase{repo: r, cfg: cfg}
}

func IsAPIKeyToken(token string) bool {
	prefix, secret, ok := parsePresentedKey(token)
	return ok && prefix != "" && secret != ""
}

func (uc *UseCase) Create(ctx context.Context, in usecase.CreateAPIKeyInput) (usecase.CreateAPIKeyOutput, error) {
	if err := ensureUUID(in.UserID); err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}
	if !uc.cfg.Enabled {
		return usecase.CreateAPIKeyOutput{}, d.NewForbidden(entity.CodeAPIKeyForbidden, "API keys are disabled")
	}
	if strings.TrimSpace(uc.cfg.HashPepper) == "" {
		return usecase.CreateAPIKeyOutput{}, d.NewAppError(entity.CodeUndefined, "API key configuration error", d.KindInternal, nil)
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return usecase.CreateAPIKeyOutput{}, d.NewValidationError(entity.CodeAPIKeyInvalid, "API key name is required")
	}

	normalizedScopes, err := entity.NormalizeScopes(in.Scopes)
	if err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}

	now := time.Now().UTC()
	if in.ExpiresAt != nil && !in.ExpiresAt.After(now) {
		return usecase.CreateAPIKeyOutput{}, d.NewValidationError(entity.CodeAPIKeyInvalid, "API key expiration must be in the future")
	}

	count, err := uc.repo.CountByUser(ctx, strings.TrimSpace(in.UserID))
	if err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}
	if count >= uc.cfg.MaxKeysPerUser {
		return usecase.CreateAPIKeyOutput{}, d.NewLimitExceeded(entity.CodeAPIKeyLimit, "API key limit exceeded")
	}

	plaintext, prefix, err := generateAPIKey()
	if err != nil {
		return usecase.CreateAPIKeyOutput{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	apiKey := entity.APIKey{
		ID:         uuid.NewString(),
		UserID:     strings.TrimSpace(in.UserID),
		Name:       name,
		KeyPrefix:  prefix,
		SecretHash: uc.hashKey(plaintext),
		Scopes:     normalizedScopes,
		ExpiresAt:  in.ExpiresAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := apiKey.Validate(); err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}

	if err := uc.repo.Store(ctx, &apiKey); err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}

	uc.audit(ctx, repo.APIKeyAuditEvent{
		ID:        uuid.NewString(),
		APIKeyID:  &apiKey.ID,
		EventType: "created",
		Meta: map[string]any{
			"prefix": apiKey.KeyPrefix,
		},
		CreatedAt: now,
	})

	return usecase.CreateAPIKeyOutput{APIKey: apiKey, PlaintextKey: plaintext}, nil
}

func (uc *UseCase) List(ctx context.Context, userID string) ([]entity.APIKey, error) {
	if err := ensureUUID(userID); err != nil {
		return nil, err
	}

	return uc.repo.ListByUser(ctx, strings.TrimSpace(userID))
}

func (uc *UseCase) Revoke(ctx context.Context, in usecase.RevokeAPIKeyInput) (entity.APIKey, error) {
	if err := ensureUUID(in.UserID, in.APIKeyID); err != nil {
		return entity.APIKey{}, err
	}

	apiKey, err := uc.repo.GetByIDForUser(ctx, strings.TrimSpace(in.UserID), strings.TrimSpace(in.APIKeyID))
	if err != nil {
		return entity.APIKey{}, err
	}

	now := time.Now().UTC()
	apiKey.RevokedAt = &now
	apiKey.UpdatedAt = now

	if err := uc.repo.Update(ctx, &apiKey); err != nil {
		return entity.APIKey{}, err
	}

	uc.audit(ctx, repo.APIKeyAuditEvent{
		ID:          uuid.NewString(),
		APIKeyID:    &apiKey.ID,
		ActorUserID: &apiKey.UserID,
		EventType:   "revoked",
		CreatedAt:   now,
	})

	return apiKey, nil
}

func (uc *UseCase) Rotate(ctx context.Context, in usecase.RotateAPIKeyInput) (usecase.CreateAPIKeyOutput, error) {
	if err := ensureUUID(in.UserID, in.APIKeyID); err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}

	oldKey, err := uc.repo.GetByIDForUser(ctx, strings.TrimSpace(in.UserID), strings.TrimSpace(in.APIKeyID))
	if err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}

	now := time.Now().UTC()
	if oldKey.IsRevokedAt(now) {
		return usecase.CreateAPIKeyOutput{}, d.NewForbidden(entity.CodeAPIKeyRevoked, "API key is revoked")
	}
	if oldKey.IsExpiredAt(now) {
		return usecase.CreateAPIKeyOutput{}, d.NewForbidden(entity.CodeAPIKeyExpired, "API key is expired")
	}

	newName := oldKey.Name
	if in.Name != nil && strings.TrimSpace(*in.Name) != "" {
		newName = strings.TrimSpace(*in.Name)
	}

	created, err := uc.Create(ctx, usecase.CreateAPIKeyInput{
		UserID:    oldKey.UserID,
		Name:      newName,
		Scopes:    oldKey.Scopes,
		ExpiresAt: oldKey.ExpiresAt,
	})
	if err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}

	revokedAt := now.Add(uc.cfg.RotateGraceWindow)
	oldKey.RevokedAt = &revokedAt
	oldKey.UpdatedAt = now
	if err := uc.repo.Update(ctx, &oldKey); err != nil {
		return usecase.CreateAPIKeyOutput{}, err
	}

	uc.audit(ctx, repo.APIKeyAuditEvent{
		ID:          uuid.NewString(),
		APIKeyID:    &oldKey.ID,
		ActorUserID: &oldKey.UserID,
		EventType:   "rotated",
		Meta: map[string]any{
			"new_api_key_id": created.APIKey.ID,
			"grace_until":    revokedAt,
		},
		CreatedAt: now,
	})

	return created, nil
}

func (uc *UseCase) Authenticate(ctx context.Context, in usecase.AuthenticateAPIKeyInput) (usecase.AuthenticateAPIKeyOutput, error) {
	if !uc.cfg.Enabled {
		return usecase.AuthenticateAPIKeyOutput{}, d.NewUnauthorized(entity.CodeAPIKeyForbidden, "API keys are disabled")
	}
	if strings.TrimSpace(uc.cfg.HashPepper) == "" {
		return usecase.AuthenticateAPIKeyOutput{}, d.NewUnauthorized(entity.CodeAPIKeyInvalid, "invalid API key")
	}

	prefix, _, ok := parsePresentedKey(in.PresentedKey)
	if !ok {
		uc.audit(ctx, repo.APIKeyAuditEvent{
			ID:        uuid.NewString(),
			EventType: "denied",
			Meta: map[string]any{
				"reason": "malformed",
			},
			CreatedAt: time.Now().UTC(),
		})
		return usecase.AuthenticateAPIKeyOutput{}, d.NewUnauthorized(entity.CodeAPIKeyInvalid, "invalid API key")
	}

	apiKey, err := uc.repo.GetByPrefix(ctx, prefix)
	if err != nil {
		uc.audit(ctx, repo.APIKeyAuditEvent{
			ID:        uuid.NewString(),
			EventType: "denied",
			Meta: map[string]any{
				"reason": "unknown_prefix",
				"prefix": prefix,
			},
			CreatedAt: time.Now().UTC(),
		})
		return usecase.AuthenticateAPIKeyOutput{}, d.NewUnauthorized(entity.CodeAPIKeyInvalid, "invalid API key")
	}

	expectedHash := uc.hashKey(in.PresentedKey)
	if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(apiKey.SecretHash)) != 1 {
		uc.audit(ctx, repo.APIKeyAuditEvent{
			ID:          uuid.NewString(),
			APIKeyID:    &apiKey.ID,
			ActorUserID: &apiKey.UserID,
			EventType:   "denied",
			Meta: map[string]any{
				"reason": "hash_mismatch",
			},
			CreatedAt: time.Now().UTC(),
		})
		return usecase.AuthenticateAPIKeyOutput{}, d.NewUnauthorized(entity.CodeAPIKeyInvalid, "invalid API key")
	}

	now := time.Now().UTC()
	if apiKey.IsRevokedAt(now) {
		uc.audit(ctx, repo.APIKeyAuditEvent{
			ID:          uuid.NewString(),
			APIKeyID:    &apiKey.ID,
			ActorUserID: &apiKey.UserID,
			EventType:   "denied",
			Meta: map[string]any{
				"reason": "revoked",
			},
			CreatedAt: now,
		})
		return usecase.AuthenticateAPIKeyOutput{}, d.NewUnauthorized(entity.CodeAPIKeyRevoked, "API key is revoked")
	}
	if apiKey.IsExpiredAt(now) {
		uc.audit(ctx, repo.APIKeyAuditEvent{
			ID:          uuid.NewString(),
			APIKeyID:    &apiKey.ID,
			ActorUserID: &apiKey.UserID,
			EventType:   "denied",
			Meta: map[string]any{
				"reason": "expired",
			},
			CreatedAt: now,
		})
		return usecase.AuthenticateAPIKeyOutput{}, d.NewUnauthorized(entity.CodeAPIKeyExpired, "API key is expired")
	}

	_ = uc.repo.TouchLastUsed(ctx, apiKey.ID, now)
	uc.audit(ctx, repo.APIKeyAuditEvent{
		ID:          uuid.NewString(),
		APIKeyID:    &apiKey.ID,
		ActorUserID: &apiKey.UserID,
		EventType:   "used",
		CreatedAt:   now,
	})

	apiKey.LastUsedAt = &now

	return usecase.AuthenticateAPIKeyOutput{APIKey: apiKey}, nil
}

func (uc *UseCase) hashKey(plaintext string) string {
	sum := sha256.Sum256([]byte(uc.cfg.HashPepper + ":" + plaintext))
	return hex.EncodeToString(sum[:])
}

func (uc *UseCase) audit(ctx context.Context, event repo.APIKeyAuditEvent) {
	_ = uc.repo.StoreAudit(ctx, event)
}

func generateAPIKey() (plaintext, keyPrefix string, err error) {
	idRaw := make([]byte, apiKeyIDBytes)
	if _, err = rand.Read(idRaw); err != nil {
		return "", "", err
	}
	secretRaw := make([]byte, apiKeySecretBytes)
	if _, err = rand.Read(secretRaw); err != nil {
		return "", "", err
	}

	keyID := hex.EncodeToString(idRaw)
	keyPrefix = apiKeyPrefix + keyID
	secret := base64.RawURLEncoding.EncodeToString(secretRaw)

	plaintext = fmt.Sprintf("%s.%s", keyPrefix, secret)
	return plaintext, keyPrefix, nil
}

func parsePresentedKey(token string) (prefix string, secret string, ok bool) {
	trimmed := strings.TrimSpace(token)
	if !strings.HasPrefix(trimmed, apiKeyPrefix) {
		return "", "", false
	}

	parts := strings.SplitN(trimmed, ".", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	if strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}

func ensureUUID(ids ...string) error {
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			return d.NewValidationError(entity.CodeInvalidID, "UUID is required")
		}
		if _, err := uuid.Parse(trimmed); err != nil {
			return d.NewValidationError(entity.CodeInvalidID, "UUID is invalid")
		}
	}
	return nil
}
