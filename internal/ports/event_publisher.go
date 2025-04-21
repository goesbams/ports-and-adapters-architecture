package ports

import "context"

// Event represents a domain event that can be published
type Event struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// EventPublisher defines the port for publishing events
type EventPublisher interface {
	// Publish publishes an event to a specified topic
	Publish(ctx context.Context, topic string, event Event) error

	// Close closes the event publisher
	Close() error
}
