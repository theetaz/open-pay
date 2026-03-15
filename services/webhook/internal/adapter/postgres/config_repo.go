package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/webhook/internal/domain"
)

// ConfigRepository is the PostgreSQL implementation for webhook config persistence.
type ConfigRepository struct {
	pool *pgxpool.Pool
}

// NewConfigRepository creates a new PostgreSQL-backed ConfigRepository.
func NewConfigRepository(pool *pgxpool.Pool) *ConfigRepository {
	return &ConfigRepository{pool: pool}
}

func (r *ConfigRepository) Create(ctx context.Context, cfg *domain.WebhookConfig) error {
	eventsJSON, _ := json.Marshal(cfg.Events)

	query := `INSERT INTO webhook_configs (id, merchant_id, url, signing_public_key, signing_private_key, events, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (merchant_id) DO UPDATE SET
			url = $3, signing_public_key = $4, signing_private_key = $5,
			events = $6, is_active = $7, updated_at = $9`

	_, err := r.pool.Exec(ctx, query,
		cfg.ID, cfg.MerchantID, cfg.URL,
		cfg.SigningPublicKey, cfg.SigningPrivateKey,
		eventsJSON, cfg.IsActive, cfg.CreatedAt, cfg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting webhook config: %w", err)
	}
	return nil
}

func (r *ConfigRepository) GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*domain.WebhookConfig, error) {
	query := `SELECT id, merchant_id, url, signing_public_key, signing_private_key, events, is_active, created_at, updated_at
		FROM webhook_configs WHERE merchant_id = $1`

	var cfg domain.WebhookConfig
	var eventsJSON []byte

	err := r.pool.QueryRow(ctx, query, merchantID).Scan(
		&cfg.ID, &cfg.MerchantID, &cfg.URL,
		&cfg.SigningPublicKey, &cfg.SigningPrivateKey,
		&eventsJSON, &cfg.IsActive, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrWebhookConfigNotFound
	}

	_ = json.Unmarshal(eventsJSON, &cfg.Events)
	return &cfg, nil
}

func (r *ConfigRepository) Update(ctx context.Context, cfg *domain.WebhookConfig) error {
	eventsJSON, _ := json.Marshal(cfg.Events)

	query := `UPDATE webhook_configs SET url = $2, events = $3, is_active = $4, updated_at = $5
		WHERE id = $1`

	result, err := r.pool.Exec(ctx, query,
		cfg.ID, cfg.URL, eventsJSON, cfg.IsActive, cfg.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating webhook config: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrWebhookConfigNotFound
	}
	return nil
}
