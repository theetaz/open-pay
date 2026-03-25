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

// EmailSender sends emails via SMTP.
type EmailSender interface {
	SendEmail(to, subject, htmlBody string) error
}

type NotificationService struct {
	repo   NotificationRepository
	sender EmailSender
}

func NewNotificationService(repo NotificationRepository, sender EmailSender) *NotificationService {
	return &NotificationService{repo: repo, sender: sender}
}

// Send creates and sends a notification.
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

	// Send via appropriate channel
	if n.Channel == domain.ChannelEmail && s.sender != nil {
		htmlBody := wrapHTML(n.Subject, n.Body)
		if err := s.sender.SendEmail(n.Recipient, n.Subject, htmlBody); err != nil {
			n.MarkFailed(err.Error())
			_ = s.repo.Update(ctx, n)
			return n, fmt.Errorf("sending email: %w", err)
		}
	}

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

// wrapHTML wraps plain text notification body in a basic HTML email template.
func wrapHTML(subject, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px; color: #1a1a1a;">
  <div style="border-bottom: 3px solid #2563eb; padding-bottom: 16px; margin-bottom: 24px;">
    <h2 style="margin: 0; color: #2563eb;">Open Pay</h2>
  </div>
  <h3 style="margin: 0 0 16px 0;">%s</h3>
  <div style="line-height: 1.6; color: #374151;">%s</div>
  <div style="margin-top: 32px; padding-top: 16px; border-top: 1px solid #e5e7eb; font-size: 12px; color: #9ca3af;">
    <p>This is an automated message from Open Pay. Please do not reply to this email.</p>
  </div>
</body>
</html>`, subject, body)
}
