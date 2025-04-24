package infrastructure

import "context"

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, event Event) error

// EventConsumer defines the port for consuming events
type EventConsumer interface {
	// Subsribe registers a handler for a specific topic
	Subscribe(topic string, handler EventHandler) error

	// SubscribeWithGroup registers a handler for a specific topic with a consumer group
	SubscribeWithGroup(topic string, groupID string, handler EventHandler)

	// Unsubscribe removes a handler for a spesific topic
	Unsubscribe(topic string) error

	// Start begins consuming events, typically in a separate goroutine
	Start(ctx context.Context) error

	// Stop stops consuming events
	Stop() error

	// GetStats returns statistics about the consumer
	GetStats() map[string]interface{}

	// IsRunning checks if the consumer is currently running
	IsRunning() bool
}
