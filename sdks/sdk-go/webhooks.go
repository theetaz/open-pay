package openpay

import "context"

// WebhooksService handles webhook operations.
type WebhooksService struct {
	client *Client
}

// WebhookConfig holds webhook configuration.
type WebhookConfig struct {
	URL    string   `json:"url"`
	Events []string `json:"events,omitempty"`
}

// Configure sets the webhook endpoint for the merchant account.
func (s *WebhooksService) Configure(ctx context.Context, config WebhookConfig) error {
	return s.client.doRequest(ctx, "POST", "/v1/sdk/webhooks/configure", config, nil)
}

// GetPublicKey retrieves the ED25519 public key for verifying webhook signatures.
func (s *WebhooksService) GetPublicKey(ctx context.Context) (string, error) {
	var resp struct {
		Data struct {
			PublicKey string `json:"publicKey"`
		} `json:"data"`
	}
	if err := s.client.doRequest(ctx, "GET", "/v1/sdk/webhooks/public-key", nil, &resp); err != nil {
		return "", err
	}
	return resp.Data.PublicKey, nil
}

// Test sends a test webhook to the configured endpoint.
func (s *WebhooksService) Test(ctx context.Context) error {
	return s.client.doRequest(ctx, "POST", "/v1/sdk/webhooks/test", nil, nil)
}
