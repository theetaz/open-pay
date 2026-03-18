package notification

import (
	"context"
	"fmt"
	"os"

	"github.com/openlankapay/openlankapay/pkg/messaging"
)

const (
	StreamName       = "NOTIFICATIONS"
	EmailSubject     = "notifications.email"
)

// NATSClient publishes notification events to NATS JetStream.
type NATSClient struct {
	nats *messaging.Client
}

// NewNATSClient creates a NATS-backed notification client.
func NewNATSClient(natsClient *messaging.Client) (*NATSClient, error) {
	// Ensure the notifications stream exists
	ctx := context.Background()
	if err := natsClient.EnsureStream(ctx, StreamName, []string{"notifications.>"}); err != nil {
		return nil, fmt.Errorf("ensuring notifications stream: %w", err)
	}

	return &NATSClient{nats: natsClient}, nil
}

// SendEmail publishes an email notification to NATS. Non-blocking.
func (c *NATSClient) SendEmail(_ context.Context, input SendEmailInput) {
	go func() {
		ctx := context.Background()
		if err := c.nats.Publish(ctx, EmailSubject, input); err != nil {
			fmt.Fprintf(os.Stderr, "nats-notification: failed to publish %s to %s: %v\n", input.EventType, input.Recipient, err)
		}
	}()
}
