package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client wraps a NATS connection with JetStream support.
type Client struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// NewClient connects to NATS and initializes JetStream.
func NewClient(url string) (*Client, error) {
	nc, err := nats.Connect(url,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("creating JetStream context: %w", err)
	}

	return &Client{conn: nc, js: js}, nil
}

// EnsureStream creates the stream if it doesn't exist.
func (c *Client) EnsureStream(ctx context.Context, name string, subjects []string) error {
	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Retention: jetstream.WorkQueuePolicy,
		MaxAge:    24 * time.Hour,
	})
	return err
}

// Publish publishes a message to a NATS subject.
func (c *Client) Publish(ctx context.Context, subject string, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}

	_, err = c.js.Publish(ctx, subject, body)
	return err
}

// Subscribe creates a durable consumer and processes messages.
func (c *Client) Subscribe(ctx context.Context, stream, consumer, filterSubject string, handler func(msg jetstream.Msg)) error {
	cons, err := c.js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Durable:       consumer,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    5,
		AckWait:       30 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("creating consumer: %w", err)
	}

	_, err = cons.Consume(handler)
	return err
}

// Close closes the NATS connection.
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
