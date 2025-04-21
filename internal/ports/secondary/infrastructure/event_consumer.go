package ports

import "context"

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, event Event) error

// EventConsumer defines the port for consuming events
type EventConsumer interface {
	// Subsribe registers a handler for a specific topic
	Subscribe(topic string, handler EventHandler) error

	// Start begins consuming events, typically in a separate goroutine
	Start(ctx context.Context) error

	// Stop stops consuming events
	Stop() error
}
