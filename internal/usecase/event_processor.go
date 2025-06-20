package usecase

import (
	"context"
	"fmt"
	"log"
	"ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
)

// EventProcessor implements the event processing service
type EventProcessor struct {
	consumer       infrastructure.EventConsumer
	walletService  *WalletService
	paymentService *PaymentService
}

// NewEventProcessor creates a new event processor
func NewEventProcessor(
	consumer infrastructure.EventConsumer,
	walletService *WalletService,
	paymentService *PaymentService,
) *EventProcessor {
	return &EventProcessor{
		consumer:       consumer,
		walletService:  walletService,
		paymentService: paymentService,
	}
}

// Start registers all event handlers and starts consuming events
func (p *EventProcessor) Start(ctx context.Context) error {
	// Register handlers for different event types

	// Wallet events
	if err := p.consumer.Subscribe("wallets", p.handleWalletEvent); err != nil {
		return fmt.Errorf("failed to subscribe to wallet events: %w", err)
	}

	// Payment events
	if err := p.consumer.Subscribe("payments", p.handlePaymentEvent); err != nil {
		return fmt.Errorf("failed to subscribe to payment events: %w", err)
	}

	// Transaction events
	if err := p.consumer.Subscribe("transactions", p.handleTransactionEvent); err != nil {
		return fmt.Errorf("failed to subscribe to transaction events: %w", err)
	}

	// Start consuming
	return p.consumer.Start(ctx)
}

// Stop stops consuming events
func (p *EventProcessor) Stop() error {
	return p.consumer.Stop()
}

// HandleEvent processes a specific event
func (p *EventProcessor) HandleEvent(ctx context.Context, topic string, event infrastructure.Event) error {
	switch topic {
	case "wallets":
		return p.handleWalletEvent(ctx, event)
	case "payments":
		return p.handlePaymentEvent(ctx, event)
	case "transactions":
		return p.handleTransactionEvent(ctx, event)
	default:
		return fmt.Errorf("unknown topic: %s", topic)
	}
}

// handleWalletEvent processes wallet-related events
func (p *EventProcessor) handleWalletEvent(ctx context.Context, event infrastructure.Event) error {
	log.Printf("Processing wallet event: %s", event.Type)

	switch event.Type {
	case "wallet.created":
		// Handle wallet creation event
		// Could send welcome email, initialize features, etc.
		return nil

	case "wallet.deposit":
		// Handle deposit event
		// Could send notification, update analytics, etc.
		return nil

	case "wallet.withdrawal":
		// Handle withdrawal event
		// Could check for suspicious activity, send notification, etc.
		return nil

	case "wallet.transfer":
		// Handle transfer event
		// Could update analytics, check limits, etc.
		return nil

	default:
		log.Printf("Unknown wallet event type: %s", event.Type)
		return nil
	}
}

// handlePaymentEvent processes payment-related events
func (p *EventProcessor) handlePaymentEvent(ctx context.Context, event infrastructure.Event) error {
	log.Printf("Processing payment event: %s", event.Type)

	switch event.Type {
	case "payment.initiated":
		// Handle payment initiated event
		// Could set up monitoring, send notification, etc.
		return nil

	case "payment.status_updated":
		// Handle payment status update
		// Could trigger wallet update, send notification, etc.
		return nil

	case "payment.cancelled":
		// Handle payment cancellation
		// Could refund, notify user, etc.
		return nil

	default:
		log.Printf("Unknown payment event type: %s", event.Type)
		return nil
	}
}

// handleTransactionEvent processes transaction-related events
func (p *EventProcessor) handleTransactionEvent(ctx context.Context, event infrastructure.Event) error {
	log.Printf("Processing transaction event: %s", event.Type)

	switch event.Type {
	case "transaction.created":
		// Handle transaction creation
		// Could validate, check limits, etc.
		return nil

	case "transaction.status_updated":
		// Handle status update
		// Could update related records, notify, etc.
		return nil

	case "transaction.reconciled":
		// Handle reconciliation
		// Could update reports, notify admins, etc.
		return nil

	default:
		log.Printf("Unknown transaction event type: %s", event.Type)
		return nil
	}
}
