package persistent

import (
	"context"
	"encoding/json"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/maximka-44js/lc/internal/access/entity"
	"github.com/maximka-44js/lc/internal/access/repo"
	d "github.com/maximka-44js/lc/internal/app/domain"
	"github.com/maximka-44js/lc/pkg/postgres"
)

type APIKeyRepo struct {
	*postgres.Postgres
}

func NewAPIKeyRepo(pg *postgres.Postgres) *APIKeyRepo {
	return &APIKeyRepo{Postgres: pg}
}

var _ repo.APIKeyRepo = (*APIKeyRepo)(nil)

func (r *APIKeyRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Select("count(*)").
		From("api_keys").
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return 0, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	var count int
	if err := exec.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return 0, handleRepoErr(err)
	}

	return count, nil
}

func (r *APIKeyRepo) Store(ctx context.Context, key *entity.APIKey) error {
	if key == nil {
		return d.NewValidationError(entity.CodeAPIKeyInvalid, "API key payload is required")
	}

	normalizedScopes, err := entity.NormalizeScopes(key.Scopes)
	if err != nil {
		return err
	}
	scopeJSON, err := json.Marshal(normalizedScopes)
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Insert("api_keys").
		Columns(
			"id",
			"user_id",
			"name",
			"key_prefix",
			"secret_hash",
			"scopes",
			"expires_at",
			"last_used_at",
			"revoked_at",
			"created_at",
			"updated_at",
		).Values(
		key.ID,
		key.UserID,
		key.Name,
		key.KeyPrefix,
		key.SecretHash,
		scopeJSON,
		key.ExpiresAt,
		key.LastUsedAt,
		key.RevokedAt,
		key.CreatedAt,
		key.UpdatedAt,
	).ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, sql, args...); err != nil {
		return handleRepoErr(err)
	}

	return nil
}

func (r *APIKeyRepo) ListByUser(ctx context.Context, userID string) ([]entity.APIKey, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Select(
			"id",
			"user_id",
			"name",
			"key_prefix",
			"secret_hash",
			"scopes",
			"expires_at",
			"last_used_at",
			"revoked_at",
			"created_at",
			"updated_at",
		).From("api_keys").
		Where(sq.Eq{"user_id": userID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, handleRepoErr(err)
	}
	defer rows.Close()

	out := make([]entity.APIKey, 0)
	for rows.Next() {
		k, scanErr := scanAPIKey(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		return nil, handleRepoErr(err)
	}

	return out, nil
}

func (r *APIKeyRepo) GetByIDForUser(ctx context.Context, userID, apiKeyID string) (entity.APIKey, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Select(
			"id",
			"user_id",
			"name",
			"key_prefix",
			"secret_hash",
			"scopes",
			"expires_at",
			"last_used_at",
			"revoked_at",
			"created_at",
			"updated_at",
		).From("api_keys").
		Where(sq.Eq{"id": apiKeyID}).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return entity.APIKey{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	k, scanErr := scanAPIKey(exec.QueryRow(ctx, sql, args...))
	if scanErr != nil {
		return entity.APIKey{}, scanErr
	}

	return k, nil
}

func (r *APIKeyRepo) GetByPrefix(ctx context.Context, keyPrefix string) (entity.APIKey, error) {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Select(
			"id",
			"user_id",
			"name",
			"key_prefix",
			"secret_hash",
			"scopes",
			"expires_at",
			"last_used_at",
			"revoked_at",
			"created_at",
			"updated_at",
		).From("api_keys").
		Where(sq.Eq{"key_prefix": keyPrefix}).
		ToSql()
	if err != nil {
		return entity.APIKey{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	k, scanErr := scanAPIKey(exec.QueryRow(ctx, sql, args...))
	if scanErr != nil {
		return entity.APIKey{}, scanErr
	}

	return k, nil
}

func (r *APIKeyRepo) Update(ctx context.Context, key *entity.APIKey) error {
	if key == nil {
		return d.NewValidationError(entity.CodeAPIKeyInvalid, "API key payload is required")
	}
	if err := key.Validate(); err != nil {
		return err
	}

	normalizedScopes, err := entity.NormalizeScopes(key.Scopes)
	if err != nil {
		return err
	}
	scopeJSON, err := json.Marshal(normalizedScopes)
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Update("api_keys").
		Set("name", key.Name).
		Set("scopes", scopeJSON).
		Set("expires_at", key.ExpiresAt).
		Set("last_used_at", key.LastUsedAt).
		Set("revoked_at", key.RevokedAt).
		Set("updated_at", key.UpdatedAt).
		Where(sq.Eq{"id": key.ID}).
		Where(sq.Eq{"user_id": key.UserID}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	res, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		return handleRepoErr(err)
	}
	if res.RowsAffected() == 0 {
		return d.NewNotFound(entity.CodeAPIKeyNotFound, "API key not found")
	}

	return nil
}

func (r *APIKeyRepo) TouchLastUsed(ctx context.Context, apiKeyID string, usedAt time.Time) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	sql, args, err := r.Builder.
		Update("api_keys").
		Set("last_used_at", usedAt).
		Set("updated_at", usedAt).
		Where(sq.Eq{"id": apiKeyID}).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	_, err = exec.Exec(ctx, sql, args...)
	if err != nil {
		return handleRepoErr(err)
	}

	return nil
}

func (r *APIKeyRepo) StoreAudit(ctx context.Context, event repo.APIKeyAuditEvent) error {
	exec := postgres.ExecutorFromContext(ctx, r.Pool)
	meta := event.Meta
	if meta == nil {
		meta = map[string]any{}
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	sql, args, err := r.Builder.
		Insert("api_key_audit_logs").
		Columns("id", "api_key_id", "actor_user_id", "event_type", "meta", "created_at").
		Values(event.ID, event.APIKeyID, event.ActorUserID, event.EventType, metaJSON, event.CreatedAt).
		ToSql()
	if err != nil {
		return d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	if _, err := exec.Exec(ctx, sql, args...); err != nil {
		return handleRepoErr(err)
	}

	return nil
}

func scanAPIKey(row interface{ Scan(dest ...any) error }) (entity.APIKey, error) {
	var k entity.APIKey
	var scopesRaw []byte

	err := row.Scan(
		&k.ID,
		&k.UserID,
		&k.Name,
		&k.KeyPrefix,
		&k.SecretHash,
		&scopesRaw,
		&k.ExpiresAt,
		&k.LastUsedAt,
		&k.RevokedAt,
		&k.CreatedAt,
		&k.UpdatedAt,
	)
	if err != nil {
		return entity.APIKey{}, handleRepoErr(err)
	}

	if err := json.Unmarshal(scopesRaw, &k.Scopes); err != nil {
		return entity.APIKey{}, d.NewAppError(entity.CodeUndefined, "Internal error", d.KindInternal, nil)
	}

	return k, nil
}
