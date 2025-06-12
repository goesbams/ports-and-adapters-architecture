package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
	"sync"

	"github.com/segmentio/kafka-go"
)

// KafkaEventConsumer implements the EventConsumer interface using Kafka
type KafkaEventConsumer struct {
	readers   map[string]*kafka.Reader
	handlers  map[string]infrastructure.EventHandler
	mu        sync.RWMutex
	brokers   []string
	groupID   string
	isRunning bool
	cancel    context.CancelFunc
}

// NewKafkaEventConsumer creates a new Kafka event consumer
func NewKafkaEventConsumer(brokers []string, groupID string) *KafkaEventConsumer {
	return &KafkaEventConsumer{
		readers:  make(map[string]*kafka.Reader),
		handlers: make(map[string]infrastructure.EventHandler),
		brokers:  brokers,
		groupID:  groupID,
	}
}

// Subscribe registers a handler for a specific topic
func (c *KafkaEventConsumer) Subscribe(topic string, handler infrastructure.EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return fmt.Errorf("cannot subscribe while consumer is running")
	}

	c.handlers[topic] = handler
	return nil
}

// SubscribeWithGroup registers a handler for a specific topic with a consumer group
func (c *KafkaEventConsumer) SubscribeWithGroup(topic string, groupID string, handler infrastructure.EventHandler) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return fmt.Errorf("cannot subscribe while consumer is running")
	}

	// Create a reader with specific group ID
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	c.readers[topic] = reader
	c.handlers[topic] = handler
	return nil
}

// Unsubscribe removes a handler for a specific topic
func (c *KafkaEventConsumer) Unsubscribe(topic string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return fmt.Errorf("cannot unsubscribe while consumer is running")
	}

	delete(c.handlers, topic)

	if reader, exists := c.readers[topic]; exists {
		reader.Close()
		delete(c.readers, topic)
	}

	return nil
}

// Start begins consuming events
func (c *KafkaEventConsumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.isRunning {
		c.mu.Unlock()
		return fmt.Errorf("consumer is already running")
	}

	// Create context with cancel
	consumerCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.isRunning = true

	// Create readers for topics without custom readers
	for topic := range c.handlers {
		if _, exists := c.readers[topic]; !exists {
			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:  c.brokers,
				Topic:    topic,
				GroupID:  c.groupID,
				MinBytes: 10e3, // 10KB
				MaxBytes: 10e6, // 10MB
			})
			c.readers[topic] = reader
		}
	}
	c.mu.Unlock()

	// Start a goroutine for each topic
	var wg sync.WaitGroup
	for topic, reader := range c.readers {
		wg.Add(1)
		go func(topic string, reader *kafka.Reader) {
			defer wg.Done()
			c.consumeTopic(consumerCtx, topic, reader)
		}(topic, reader)
	}

	// Wait for context cancellation
	<-consumerCtx.Done()

	// Wait for all goroutines to finish
	wg.Wait()

	return nil
}

// consumeTopic consumes events from a specific topic
func (c *KafkaEventConsumer) consumeTopic(ctx context.Context, topic string, reader *kafka.Reader) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read message
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // Context cancelled
				}
				// Log error and continue
				fmt.Printf("failed to read message from topic %s: %v\n", topic, err)
				continue
			}

			// Unmarshal event
			var event infrastructure.Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				fmt.Printf("failed to unmarshal event from topic %s: %v\n", topic, err)
				continue
			}

			// Get handler
			c.mu.RLock()
			handler, exists := c.handlers[topic]
			c.mu.RUnlock()

			if !exists {
				fmt.Printf("no handler found for topic %s\n", topic)
				continue
			}

			// Handle event
			if err := handler(ctx, event); err != nil {
				fmt.Printf("failed to handle event from topic %s: %v\n", topic, err)
				// In production, you might want to implement retry logic or dead letter queue
			}
		}
	}
}

// Stop stops consuming events
func (c *KafkaEventConsumer) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isRunning {
		return fmt.Errorf("consumer is not running")
	}

	// Cancel context
	if c.cancel != nil {
		c.cancel()
	}

	// Close all readers
	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			return fmt.Errorf("failed to close reader: %w", err)
		}
	}

	c.readers = make(map[string]*kafka.Reader)
	c.isRunning = false

	return nil
}

// GetStats returns statistics about the consumer
func (c *KafkaEventConsumer) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["is_running"] = c.isRunning
	stats["subscribed_topics"] = len(c.handlers)

	topics := make([]string, 0, len(c.handlers))
	for topic := range c.handlers {
		topics = append(topics, topic)
	}
	stats["topics"] = topics

	return stats
}

// IsRunning checks if the consumer is currently running
func (c *KafkaEventConsumer) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isRunning
}
