package primary

import (
	"context"
	infraports "ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
)

// EventProcessor defines the contract for event processing service
type EventProcessor interface {

	// Start registers all event handlers and starts consuming events
	Start(ctx context.Context) error

	// Stop stops consuming events
	Stop() error

	// HandleEvent process a specific event
	// This is generaaly not called directly but is useful for testing
	HandleEvent(ctx context.Context, topic string, event infraports.Event) error
}
