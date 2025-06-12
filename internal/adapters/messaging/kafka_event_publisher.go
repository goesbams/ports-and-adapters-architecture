package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaEventPublisher implements the EventPublisher interface using Kafka
type KafkaEventPublisher struct {
	writers map[string]*kafka.Writer
	mu      sync.RWMutex
	brokers []string
}

// NewKafkaEventPublisher creates a new Kafka event publisher
func NewKafkaEventPublisher(brokers []string) *KafkaEventPublisher {
	return &KafkaEventPublisher{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
	}
}

// getWriter gets or creates a writer for a topic
func (p *KafkaEventPublisher) getWriter(topic string) *kafka.Writer {
	p.mu.RLock()
	writer, exists := p.writers[topic]
	p.mu.RUnlock()

	if exists {
		return writer
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	writer, exists = p.writers[topic]
	if exists {
		return writer
	}

	// Create new writer
	writer = &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}

	p.writers[topic] = writer
	return writer
}

// Publish publishes an event to a specified topic
func (p *KafkaEventPublisher) Publish(ctx context.Context, topic string, event infrastructure.Event) error {
	// Set event metadata if not already set
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Time == 0 {
		event.Time = time.Now().UnixMilli()
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Get writer for topic
	writer := p.getWriter(topic)

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(event.ID),
		Value: data,
	}

	// Write message
	err = writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to publish event to topic %s: %w", topic, err)
	}

	return nil
}

// PublishAsync publishes an event asynchronously and returns immediately
func (p *KafkaEventPublisher) PublishAsync(ctx context.Context, topic string, event infrastructure.Event) error {
	// Set event metadata if not already set
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Time == 0 {
		event.Time = time.Now().UnixMilli()
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Get writer for topic
	writer := p.getWriter(topic)

	// Create Kafka message
	msg := kafka.Message{
		Key:   []byte(event.ID),
		Value: data,
	}

	// Write message asynchronously
	go func() {
		// Create a new context with timeout for async operation
		asyncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := writer.WriteMessages(asyncCtx, msg); err != nil {
			// In production, this should be logged properly
			fmt.Printf("failed to publish async event to topic %s: %v\n", topic, err)
		}
	}()

	return nil
}

// PublishBatch publishes multiple events to the same topic
func (p *KafkaEventPublisher) PublishBatch(ctx context.Context, topic string, events []infrastructure.Event) error {
	if len(events) == 0 {
		return nil
	}

	// Get writer for topic
	writer := p.getWriter(topic)

	// Prepare messages
	messages := make([]kafka.Message, 0, len(events))
	for _, event := range events {
		// Set event metadata if not already set
		if event.ID == "" {
			event.ID = generateEventID()
		}
		if event.Time == 0 {
			event.Time = time.Now().UnixMilli()
		}

		// Marshal event to JSON
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		messages = append(messages, kafka.Message{
			Key:   []byte(event.ID),
			Value: data,
		})
	}

	// Write messages
	err := writer.WriteMessages(ctx, messages...)
	if err != nil {
		return fmt.Errorf("failed to publish batch events to topic %s: %w", topic, err)
	}

	return nil
}

// Flush waits for all async events to be published
func (p *KafkaEventPublisher) Flush(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			return fmt.Errorf("failed to close writer for topic %s: %w", topic, err)
		}
	}

	return nil
}

// Close closes the event publisher
func (p *KafkaEventPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			return fmt.Errorf("failed to close writer for topic %s: %w", topic, err)
		}
	}

	p.writers = make(map[string]*kafka.Writer)
	return nil
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return fmt.Sprintf("evt_%d_%s", time.Now().UnixNano(), generateRandomString(8))
}

// generateRandomString generates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
