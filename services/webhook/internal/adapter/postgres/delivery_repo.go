package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/webhook/internal/domain"
)

// DeliveryRepository is the PostgreSQL implementation for webhook delivery persistence.
type DeliveryRepository struct {
	pool *pgxpool.Pool
}

// NewDeliveryRepository creates a new PostgreSQL-backed DeliveryRepository.
func NewDeliveryRepository(pool *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{pool: pool}
}

func (r *DeliveryRepository) Create(ctx context.Context, d *domain.Delivery) error {
	query := `INSERT INTO webhook_deliveries (id, webhook_config_id, merchant_id, event_type, payload,
		attempt_count, max_attempts, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		d.ID, d.WebhookConfigID, d.MerchantID, d.EventType, d.Payload,
		d.AttemptCount, d.MaxAttempts, string(d.Status),
		d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting delivery: %w", err)
	}
	return nil
}

func (r *DeliveryRepository) Update(ctx context.Context, d *domain.Delivery) error {
	query := `UPDATE webhook_deliveries SET
		attempt_count = $2, status = $3, last_response_code = $4,
		last_response_body = $5, last_error = $6, next_attempt_at = $7,
		delivered_at = $8, updated_at = $9
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query,
		d.ID, d.AttemptCount, string(d.Status), d.LastResponseCode,
		d.LastResponseBody, d.LastError, d.NextAttemptAt,
		d.DeliveredAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating delivery: %w", err)
	}
	return nil
}

func (r *DeliveryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error) {
	query := `SELECT id, webhook_config_id, merchant_id, event_type, payload,
		attempt_count, max_attempts, status, last_response_code,
		last_response_body, last_error, next_attempt_at, delivered_at,
		created_at, updated_at
		FROM webhook_deliveries WHERE id = $1`

	var d domain.Delivery
	var status string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&d.ID, &d.WebhookConfigID, &d.MerchantID, &d.EventType, &d.Payload,
		&d.AttemptCount, &d.MaxAttempts, &status, &d.LastResponseCode,
		&d.LastResponseBody, &d.LastError, &d.NextAttemptAt, &d.DeliveredAt,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, domain.ErrDeliveryNotFound
	}

	d.Status = domain.DeliveryStatus(status)
	return &d, nil
}
