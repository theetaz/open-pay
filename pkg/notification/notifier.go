package notification

import "context"

// Notifier is the interface for sending notifications.
// Both the HTTP Client and NATSClient implement this.
type Notifier interface {
	SendEmail(ctx context.Context, input SendEmailInput)
}
