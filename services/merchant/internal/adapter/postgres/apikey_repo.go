package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/merchant/internal/domain"
)

// APIKeyRepository is the PostgreSQL implementation for API key persistence.
type APIKeyRepository struct {
	pool *pgxpool.Pool
}

// NewAPIKeyRepository creates a new PostgreSQL-backed APIKeyRepository.
func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

const apiKeySelectCols = `id, merchant_id, key_id, secret_hash, COALESCE(secret_hmac_key, ''), name, environment,
	is_active, revoked_at, revoked_reason, last_used_at, created_at`

func scanAPIKey(scan func(dest ...any) error) (*domain.APIKey, error) {
	var key domain.APIKey
	err := scan(
		&key.ID, &key.MerchantID, &key.KeyID, &key.SecretHash, &key.SecretHMACKey,
		&key.Name, &key.Environment, &key.IsActive,
		&key.RevokedAt, &key.RevokedReason, &key.LastUsedAt, &key.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (id, merchant_id, key_id, secret_hash, secret_hmac_key, name, environment, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		key.ID, key.MerchantID, key.KeyID, key.SecretHash, key.SecretHMACKey,
		key.Name, key.Environment, key.IsActive, key.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting API key: %w", err)
	}
	return nil
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	query := `SELECT ` + apiKeySelectCols + ` FROM api_keys WHERE id = $1`
	key, err := scanAPIKey(r.pool.QueryRow(ctx, query, id).Scan)
	if err != nil {
		return nil, domain.ErrAPIKeyNotFound
	}
	return key, nil
}

func (r *APIKeyRepository) GetByKeyID(ctx context.Context, keyID string) (*domain.APIKey, error) {
	query := `SELECT ` + apiKeySelectCols + ` FROM api_keys WHERE key_id = $1`
	key, err := scanAPIKey(r.pool.QueryRow(ctx, query, keyID).Scan)
	if err != nil {
		return nil, domain.ErrAPIKeyNotFound
	}
	return key, nil
}

// GetHMACKeyByKeyID returns the HMAC signing key and merchant ID for a given key ID.
// Used by the gateway for HMAC signature verification.
func (r *APIKeyRepository) GetHMACKeyByKeyID(ctx context.Context, keyID string) (hmacKey string, merchantID uuid.UUID, err error) {
	query := `SELECT COALESCE(secret_hmac_key, ''), merchant_id FROM api_keys WHERE key_id = $1 AND is_active = TRUE`
	err = r.pool.QueryRow(ctx, query, keyID).Scan(&hmacKey, &merchantID)
	if err != nil {
		return "", uuid.Nil, domain.ErrAPIKeyNotFound
	}
	if hmacKey == "" {
		return "", uuid.Nil, fmt.Errorf("api key has no HMAC key (created before HMAC support)")
	}
	return hmacKey, merchantID, nil
}

func (r *APIKeyRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.APIKey, error) {
	query := `SELECT ` + apiKeySelectCols + ` FROM api_keys WHERE merchant_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing API keys: %w", err)
	}
	defer rows.Close()

	var keys []*domain.APIKey
	for rows.Next() {
		key, err := scanAPIKey(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("scanning API key: %w", err)
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (r *APIKeyRepository) Update(ctx context.Context, key *domain.APIKey) error {
	query := `UPDATE api_keys SET is_active = $2, revoked_at = $3, revoked_reason = $4, last_used_at = $5 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, key.ID, key.IsActive, key.RevokedAt, key.RevokedReason, key.LastUsedAt)
	if err != nil {
		return fmt.Errorf("updating API key: %w", err)
	}
	return nil
}

func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting API key: %w", err)
	}
	return nil
}
