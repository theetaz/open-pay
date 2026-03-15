package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidNotification = errors.New("invalid notification")
	ErrNotificationNotFound = errors.New("notification not found")
)

// Channel types.
type Channel = string

const (
	ChannelEmail Channel = "EMAIL"
	ChannelSMS   Channel = "SMS"
	ChannelPush  Channel = "PUSH"
)

var validChannels = map[string]bool{
	ChannelEmail: true, ChannelSMS: true, ChannelPush: true,
}

// Notification status values.
type Status = string

const (
	StatusPending Status = "PENDING"
	StatusSent    Status = "SENT"
	StatusFailed  Status = "FAILED"
)

// NotificationInput holds data to create a notification.
type NotificationInput struct {
	MerchantID  uuid.UUID
	Channel     Channel
	Recipient   string
	Subject     string
	Body        string
	EventType   string
	ReferenceID *uuid.UUID
}

// Notification represents a notification to be sent.
type Notification struct {
	ID            uuid.UUID
	MerchantID    uuid.UUID
	Channel       Channel
	Recipient     string
	Subject       string
	Body          string
	EventType     string
	ReferenceID   *uuid.UUID
	Status        Status
	FailureReason string
	SentAt        *time.Time
	CreatedAt     time.Time
}

// NewNotification creates a validated notification.
func NewNotification(input NotificationInput) (*Notification, error) {
	if input.Recipient == "" {
		return nil, fmt.Errorf("%w: recipient is required", ErrInvalidNotification)
	}
	if !validChannels[input.Channel] {
		return nil, fmt.Errorf("%w: unsupported channel %s", ErrInvalidNotification, input.Channel)
	}

	return &Notification{
		ID:          uuid.New(),
		MerchantID:  input.MerchantID,
		Channel:     input.Channel,
		Recipient:   input.Recipient,
		Subject:     input.Subject,
		Body:        input.Body,
		EventType:   input.EventType,
		ReferenceID: input.ReferenceID,
		Status:      StatusPending,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// MarkSent transitions the notification to SENT.
func (n *Notification) MarkSent() {
	now := time.Now().UTC()
	n.Status = StatusSent
	n.SentAt = &now
}

// MarkFailed transitions the notification to FAILED with a reason.
func (n *Notification) MarkFailed(reason string) {
	n.Status = StatusFailed
	n.FailureReason = reason
}

// RenderTemplate generates notification body from a template and variables.
func RenderTemplate(eventType string, vars map[string]string) string {
	templates := map[string]string{
		"payment.paid": "Payment confirmed! Amount: {{amount}} {{currency}}. Payment No: {{paymentNo}}.",
		"payment.expired": "Payment {{paymentNo}} has expired.",
		"payment.failed": "Payment {{paymentNo}} has failed.",
		"merchant.approved": "Your merchant account has been approved. You can now accept payments.",
		"withdrawal.completed": "Your withdrawal of {{amount}} {{currency}} has been processed. Bank reference: {{bankRef}}.",
		"subscription.payment.due": "Your subscription payment of {{amount}} {{currency}} is due.",
	}

	tmpl, ok := templates[eventType]
	if !ok {
		return "You have a new notification from Open Pay."
	}

	result := tmpl
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}
