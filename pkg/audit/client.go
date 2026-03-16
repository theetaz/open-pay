package audit

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

// Client sends audit log entries to the admin service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new audit log client.
func NewClient(adminServiceURL string) *Client {
	return &Client{
		baseURL:    adminServiceURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// LogEntry represents an audit log entry to create.
type LogEntry struct {
	ActorID      uuid.UUID  `json:"actorId"`
	ActorType    string     `json:"actorType"` // ADMIN, MERCHANT_USER, SYSTEM
	MerchantID   *uuid.UUID `json:"merchantId,omitempty"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resourceType"`
	ResourceID   *uuid.UUID `json:"resourceId,omitempty"`
	IPAddress    string     `json:"ipAddress,omitempty"`
	UserAgent    string     `json:"userAgent,omitempty"`
}

// Log sends an audit log entry to the admin service. Non-blocking — errors are logged to stderr.
// Uses a background context to avoid cancellation when the HTTP response is already sent.
func (c *Client) Log(_ context.Context, entry LogEntry) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.logSync(bgCtx, entry); err != nil {
			fmt.Fprintf(os.Stderr, "audit: failed to log %s: %v\n", entry.Action, err)
		}
	}()
}

func (c *Client) logSync(ctx context.Context, entry LogEntry) error {
	body, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling audit entry: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/audit-logs", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating audit request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending audit log: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("audit service returned %d", resp.StatusCode)
	}

	return nil
}
