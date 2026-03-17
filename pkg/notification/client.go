package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// Client sends notification requests to the notification service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new notification client.
func NewClient(notificationServiceURL string) *Client {
	return &Client{
		baseURL:    notificationServiceURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendEmailInput holds data to send an email notification.
type SendEmailInput struct {
	MerchantID uuid.UUID
	Recipient  string
	Subject    string
	Body       string
	EventType  string
}

// SendEmail sends an email notification via the notification service. Non-blocking.
func (c *Client) SendEmail(_ context.Context, input SendEmailInput) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.sendSync(bgCtx, input); err != nil {
			fmt.Fprintf(os.Stderr, "notification: failed to send %s to %s: %v\n", input.EventType, input.Recipient, err)
		}
	}()
}

func (c *Client) sendSync(ctx context.Context, input SendEmailInput) error {
	payload := map[string]string{
		"merchantId": input.MerchantID.String(),
		"channel":    "EMAIL",
		"recipient":  input.Recipient,
		"subject":    input.Subject,
		"body":       input.Body,
		"eventType":  input.EventType,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/notifications/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("notification service returned %d", resp.StatusCode)
	}

	return nil
}
