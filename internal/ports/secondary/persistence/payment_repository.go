package persistence

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// PaymentRepository defines the port for payment data operations
type PaymentRepository interface {

	// FindByID retrieves a payment by its ID
	FindByID(ctx context.Context, id int) (*domain.Payment, error)

	// FindByTransactionID retrieves all payment for a transaction
	FindByTransactionID(ctx context.Context, transactionID int) ([]*domain.Payment, error)

	//FindByExternalID retrieves all payment for a transaction
	FindByExternalID(ctx context.Context, externalID string) (*domain.Payment, error)

	// FindPendingPayments retrieves all pending payments with optional age limit in minutes
	FindPendingPayments(ctx context.Context, olderThanMinutes int) ([]*domain.Payment, error)

	// Create saves a new payment
	Create(ctx context.Context, payment *domain.Payment) error

	// Update updates an existing payment
	Update(ctx context.Context, payment *domain.Payment) error
}
