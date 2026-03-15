package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/openlankapay/openlankapay/services/notification/internal/domain"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	query := `INSERT INTO notifications (id, merchant_id, channel, recipient, subject, body,
		event_type, reference_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		n.ID, n.MerchantID, string(n.Channel), n.Recipient, n.Subject, n.Body,
		n.EventType, n.ReferenceID, string(n.Status), n.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `SELECT id, merchant_id, channel, recipient, subject, body,
		event_type, reference_id, status, failure_reason, sent_at, created_at
		FROM notifications WHERE id = $1`

	var n domain.Notification
	var channel, status string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.MerchantID, &channel, &n.Recipient, &n.Subject, &n.Body,
		&n.EventType, &n.ReferenceID, &status, &n.FailureReason, &n.SentAt, &n.CreatedAt,
	)
	if err != nil {
		return nil, domain.ErrNotificationNotFound
	}

	n.Channel = domain.Channel(channel)
	n.Status = domain.Status(status)
	return &n, nil
}

func (r *NotificationRepository) Update(ctx context.Context, n *domain.Notification) error {
	query := `UPDATE notifications SET status = $2, failure_reason = $3, sent_at = $4 WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, n.ID, string(n.Status), n.FailureReason, n.SentAt)
	if err != nil {
		return fmt.Errorf("updating notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Notification, error) {
	query := `SELECT id, merchant_id, channel, recipient, subject, body,
		event_type, reference_id, status, failure_reason, sent_at, created_at
		FROM notifications WHERE merchant_id = $1 ORDER BY created_at DESC LIMIT 50`

	rows, err := r.pool.Query(ctx, query, merchantID)
	if err != nil {
		return nil, fmt.Errorf("listing notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		var channel, status string
		if err := rows.Scan(
			&n.ID, &n.MerchantID, &channel, &n.Recipient, &n.Subject, &n.Body,
			&n.EventType, &n.ReferenceID, &status, &n.FailureReason, &n.SentAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning notification: %w", err)
		}
		n.Channel = domain.Channel(channel)
		n.Status = domain.Status(status)
		notifications = append(notifications, &n)
	}
	return notifications, nil
}
