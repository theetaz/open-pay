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

// ListRetryable returns deliveries that are due for retry (status=PENDING, next_attempt_at <= now).
func (r *DeliveryRepository) ListRetryable(ctx context.Context, limit int) ([]*domain.Delivery, error) {
	query := `SELECT id, webhook_config_id, merchant_id, event_type, payload,
		attempt_count, max_attempts, status, last_response_code,
		last_response_body, last_error, next_attempt_at, delivered_at,
		created_at, updated_at
		FROM webhook_deliveries
		WHERE status = 'PENDING' AND next_attempt_at IS NOT NULL AND next_attempt_at <= NOW()
		ORDER BY next_attempt_at ASC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("listing retryable deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*domain.Delivery
	for rows.Next() {
		var d domain.Delivery
		var status string
		if err := rows.Scan(
			&d.ID, &d.WebhookConfigID, &d.MerchantID, &d.EventType, &d.Payload,
			&d.AttemptCount, &d.MaxAttempts, &status, &d.LastResponseCode,
			&d.LastResponseBody, &d.LastError, &d.NextAttemptAt, &d.DeliveredAt,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning delivery: %w", err)
		}
		d.Status = domain.DeliveryStatus(status)
		deliveries = append(deliveries, &d)
	}
	return deliveries, nil
}

// ListByMerchant returns deliveries for a merchant with pagination.
func (r *DeliveryRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID, page, perPage int) ([]*domain.Delivery, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM webhook_deliveries WHERE merchant_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, merchantID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting deliveries: %w", err)
	}

	offset := (page - 1) * perPage
	query := `SELECT id, webhook_config_id, merchant_id, event_type, payload,
		attempt_count, max_attempts, status, last_response_code,
		last_response_body, last_error, next_attempt_at, delivered_at,
		created_at, updated_at
		FROM webhook_deliveries WHERE merchant_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, merchantID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing deliveries: %w", err)
	}
	defer rows.Close()

	var deliveries []*domain.Delivery
	for rows.Next() {
		var d domain.Delivery
		var status string
		if err := rows.Scan(
			&d.ID, &d.WebhookConfigID, &d.MerchantID, &d.EventType, &d.Payload,
			&d.AttemptCount, &d.MaxAttempts, &status, &d.LastResponseCode,
			&d.LastResponseBody, &d.LastError, &d.NextAttemptAt, &d.DeliveredAt,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning delivery: %w", err)
		}
		d.Status = domain.DeliveryStatus(status)
		deliveries = append(deliveries, &d)
	}
	return deliveries, total, nil
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
