package infrastructure

import "context"

// Event represents a domain event that can be published
type Event struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	Time    int64                  `json:"time"` // Unix timestamp in millisecond
	ID      string                 `json:"id"`   // Unique Event ID

}

// EventPublisher defines the port for publishing events
type EventPublisher interface {
	// Publish publishes an event to a specified topic
	Publish(ctx context.Context, topic string, event Event) error

	// PublishAsync publishes an event asynchrounously and returns immediately
	PublishAsync(ctx context.Context, topic string, event Event) error

	// PublishBatch publishes multiple events to the same topic
	PublishBatch(ctx context.Context, topic string, events []Event)

	// Flush waits for all async events to be published
	Flush(ctx context.Context) error

	// Close closes the event publisher
	Close() error
}
