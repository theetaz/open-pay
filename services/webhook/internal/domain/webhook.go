package domain

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidWebhookConfig = errors.New("invalid webhook configuration")
	ErrWebhookConfigNotFound = errors.New("webhook config not found")
	ErrDeliveryNotFound      = errors.New("webhook delivery not found")
)

// Delivery status values.
type DeliveryStatus string

const (
	DeliveryPending    DeliveryStatus = "PENDING"
	DeliveryDelivering DeliveryStatus = "DELIVERING"
	DeliveryDelivered  DeliveryStatus = "DELIVERED"
	DeliveryFailed     DeliveryStatus = "FAILED"
	DeliveryExhausted  DeliveryStatus = "EXHAUSTED"
)

// WebhookConfig holds a merchant's webhook configuration with ED25519 keys.
type WebhookConfig struct {
	ID                uuid.UUID
	MerchantID        uuid.UUID
	URL               string
	SigningPublicKey   string // base64-encoded ED25519 public key
	SigningPrivateKey  string // base64-encoded ED25519 private key
	Events            []string
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewWebhookConfig creates a new webhook config with generated ED25519 key pair.
func NewWebhookConfig(merchantID uuid.UUID, url string) (*WebhookConfig, error) {
	if url == "" {
		return nil, fmt.Errorf("%w: URL is required", ErrInvalidWebhookConfig)
	}
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("%w: URL must use HTTPS", ErrInvalidWebhookConfig)
	}

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("generating ED25519 key pair: %w", err)
	}

	now := time.Now().UTC()
	return &WebhookConfig{
		ID:               uuid.New(),
		MerchantID:       merchantID,
		URL:              url,
		SigningPublicKey:  base64.StdEncoding.EncodeToString(pub),
		SigningPrivateKey: base64.StdEncoding.EncodeToString(priv),
		Events:           []string{"payment.*"},
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

// Delivery represents a single webhook delivery attempt.
type Delivery struct {
	ID               uuid.UUID
	WebhookConfigID  uuid.UUID
	MerchantID       uuid.UUID
	EventType        string
	Payload          []byte
	AttemptCount     int
	MaxAttempts      int
	Status           DeliveryStatus
	LastResponseCode int
	LastResponseBody string
	LastError        string
	NextAttemptAt    *time.Time
	DeliveredAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// NewDelivery creates a new pending webhook delivery.
func NewDelivery(configID, merchantID uuid.UUID, eventType string, payload []byte) *Delivery {
	now := time.Now().UTC()
	return &Delivery{
		ID:              uuid.New(),
		WebhookConfigID: configID,
		MerchantID:      merchantID,
		EventType:       eventType,
		Payload:         payload,
		AttemptCount:    0,
		MaxAttempts:     5,
		Status:          DeliveryPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// RecordSuccess marks the delivery as successfully delivered.
func (d *Delivery) RecordSuccess(statusCode int) {
	now := time.Now().UTC()
	d.AttemptCount++
	d.Status = DeliveryDelivered
	d.LastResponseCode = statusCode
	d.DeliveredAt = &now
	d.UpdatedAt = now
}

// RecordFailure records a failed delivery attempt and schedules retry.
func (d *Delivery) RecordFailure(statusCode int, responseBody, errMsg string) {
	d.AttemptCount++
	d.LastResponseCode = statusCode
	d.LastResponseBody = responseBody
	d.LastError = errMsg
	d.UpdatedAt = time.Now().UTC()

	if d.AttemptCount >= d.MaxAttempts {
		d.Status = DeliveryExhausted
		return
	}

	// Exponential backoff: 1, 2, 4, 8 minutes
	delay := time.Duration(math.Pow(2, float64(d.AttemptCount-1))) * time.Minute
	next := time.Now().Add(delay)
	d.NextAttemptAt = &next
	d.Status = DeliveryPending
}

// CanRetry returns true if the delivery is eligible for another attempt.
func (d *Delivery) CanRetry() bool {
	return d.Status == DeliveryPending && d.AttemptCount < d.MaxAttempts
}

// ResetForRetry resets an exhausted delivery for manual retry.
func (d *Delivery) ResetForRetry() {
	d.Status = DeliveryPending
	d.AttemptCount = 0
	d.NextAttemptAt = nil
	d.UpdatedAt = time.Now().UTC()
}
