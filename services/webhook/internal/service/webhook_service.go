package service

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/openlankapay/openlankapay/services/webhook/internal/domain"
)

// ConfigRepository defines data access for webhook configs.
type ConfigRepository interface {
	Create(ctx context.Context, cfg *domain.WebhookConfig) error
	GetByMerchantID(ctx context.Context, merchantID uuid.UUID) (*domain.WebhookConfig, error)
	Update(ctx context.Context, cfg *domain.WebhookConfig) error
}

// DeliveryRepository defines data access for webhook deliveries.
type DeliveryRepository interface {
	Create(ctx context.Context, d *domain.Delivery) error
	Update(ctx context.Context, d *domain.Delivery) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error)
	ListRetryable(ctx context.Context, limit int) ([]*domain.Delivery, error)
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, page, perPage int) ([]*domain.Delivery, int, error)
}

// WebhookService orchestrates webhook operations.
type WebhookService struct {
	configs    ConfigRepository
	deliveries DeliveryRepository
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(configs ConfigRepository, deliveries DeliveryRepository) *WebhookService {
	return &WebhookService{configs: configs, deliveries: deliveries}
}

// Configure creates or updates a webhook configuration for a merchant.
func (s *WebhookService) Configure(ctx context.Context, merchantID uuid.UUID, url string) (*domain.WebhookConfig, error) {
	cfg, err := domain.NewWebhookConfig(merchantID, url)
	if err != nil {
		return nil, err
	}

	if err := s.configs.Create(ctx, cfg); err != nil {
		return nil, fmt.Errorf("storing webhook config: %w", err)
	}

	return cfg, nil
}

// GetPublicKey returns the ED25519 public key for a merchant's webhook.
func (s *WebhookService) GetPublicKey(ctx context.Context, merchantID uuid.UUID) (string, error) {
	cfg, err := s.configs.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return "", err
	}
	return cfg.SigningPublicKey, nil
}

// Deliver sends a webhook to the merchant's endpoint with ED25519 signature.
// httpClient can be nil (uses http.DefaultClient) or a custom client for testing.
func (s *WebhookService) Deliver(ctx context.Context, merchantID uuid.UUID, eventType string, payload []byte, httpClient *http.Client) (*domain.Delivery, error) {
	cfg, err := s.configs.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("getting webhook config: %w", err)
	}

	if !cfg.IsActive {
		return nil, fmt.Errorf("webhook config is inactive")
	}

	delivery := domain.NewDelivery(cfg.ID, merchantID, eventType, payload)
	if err := s.deliveries.Create(ctx, delivery); err != nil {
		return nil, fmt.Errorf("creating delivery record: %w", err)
	}

	// Sign the payload
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature, err := signPayload(cfg.SigningPrivateKey, timestamp, payload)
	if err != nil {
		delivery.RecordFailure(0, "", fmt.Sprintf("signing error: %v", err))
		_ = s.deliveries.Update(ctx, delivery)
		return delivery, err
	}

	// Send the request
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.URL, bytes.NewReader(payload))
	if err != nil {
		delivery.RecordFailure(0, "", fmt.Sprintf("request creation error: %v", err))
		_ = s.deliveries.Update(ctx, delivery)
		return delivery, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Timestamp", timestamp)
	req.Header.Set("X-Webhook-Attempt", strconv.Itoa(delivery.AttemptCount+1))
	req.Header.Set("X-Webhook-Event", eventType)
	req.Header.Set("X-Webhook-ID", delivery.ID.String())
	req.Header.Set("User-Agent", "OpenLankaPayment-Webhook/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		delivery.RecordFailure(0, "", err.Error())
		_ = s.deliveries.Update(ctx, delivery)
		return delivery, nil // Return delivery without error (retry scheduled)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.RecordSuccess(resp.StatusCode)
	} else {
		delivery.RecordFailure(resp.StatusCode, string(body), "")
	}

	_ = s.deliveries.Update(ctx, delivery)
	return delivery, nil
}

// RetryPending picks up deliveries due for retry and re-attempts them.
// Returns the number of deliveries processed.
func (s *WebhookService) RetryPending(ctx context.Context) (int, error) {
	deliveries, err := s.deliveries.ListRetryable(ctx, 50)
	if err != nil {
		return 0, fmt.Errorf("listing retryable: %w", err)
	}

	processed := 0
	httpClient := &http.Client{Timeout: 10 * time.Second}

	for _, d := range deliveries {
		cfg, err := s.configs.GetByMerchantID(ctx, d.MerchantID)
		if err != nil || !cfg.IsActive {
			d.RecordFailure(0, "", "webhook config not found or inactive")
			_ = s.deliveries.Update(ctx, d)
			continue
		}

		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		signature, err := signPayload(cfg.SigningPrivateKey, timestamp, d.Payload)
		if err != nil {
			d.RecordFailure(0, "", fmt.Sprintf("signing error: %v", err))
			_ = s.deliveries.Update(ctx, d)
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", cfg.URL, bytes.NewReader(d.Payload))
		if err != nil {
			d.RecordFailure(0, "", fmt.Sprintf("request error: %v", err))
			_ = s.deliveries.Update(ctx, d)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("X-Webhook-Timestamp", timestamp)
		req.Header.Set("X-Webhook-Attempt", strconv.Itoa(d.AttemptCount+1))
		req.Header.Set("X-Webhook-Event", d.EventType)
		req.Header.Set("X-Webhook-ID", d.ID.String())
		req.Header.Set("User-Agent", "OpenLankaPayment-Webhook/1.0")

		resp, err := httpClient.Do(req)
		if err != nil {
			d.RecordFailure(0, "", err.Error())
			_ = s.deliveries.Update(ctx, d)
			processed++
			continue
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			d.RecordSuccess(resp.StatusCode)
		} else {
			d.RecordFailure(resp.StatusCode, string(body), "")
		}

		_ = s.deliveries.Update(ctx, d)
		processed++
	}

	return processed, nil
}

// ListDeliveries returns webhook deliveries for a merchant.
func (s *WebhookService) ListDeliveries(ctx context.Context, merchantID uuid.UUID, page, perPage int) ([]*domain.Delivery, int, error) {
	return s.deliveries.ListByMerchant(ctx, merchantID, page, perPage)
}

func signPayload(privateKeyBase64, timestamp string, payload []byte) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", fmt.Errorf("decoding private key: %w", err)
	}

	privKey := ed25519.PrivateKey(privBytes)
	message := append([]byte(timestamp), payload...)
	signature := ed25519.Sign(privKey, message)
	return base64.StdEncoding.EncodeToString(signature), nil
}
