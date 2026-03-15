package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/openlankapay/openlankapay/services/notification/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	Update(ctx context.Context, n *domain.Notification) error
	ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Notification, error)
}

type NotificationService struct {
	repo NotificationRepository
}

func NewNotificationService(repo NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// Send creates and "sends" a notification (logs it for now, real SMTP/SMS in production).
func (s *NotificationService) Send(ctx context.Context, input domain.NotificationInput) (*domain.Notification, error) {
	// Render template if body is empty
	if input.Body == "" {
		input.Body = domain.RenderTemplate(input.EventType, nil)
	}

	n, err := domain.NewNotification(input)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("storing notification: %w", err)
	}

	// Mark as sent (in production, this would go through SMTP/SMS gateway)
	n.MarkSent()
	if err := s.repo.Update(ctx, n); err != nil {
		return nil, fmt.Errorf("updating notification status: %w", err)
	}

	return n, nil
}

// GetByID returns a notification by ID.
func (s *NotificationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByMerchant returns recent notifications for a merchant.
func (s *NotificationService) ListByMerchant(ctx context.Context, merchantID uuid.UUID) ([]*domain.Notification, error) {
	return s.repo.ListByMerchant(ctx, merchantID)
}
