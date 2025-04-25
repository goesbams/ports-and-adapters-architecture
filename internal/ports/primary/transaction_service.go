package primary

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// TransactionService defines contract for transaction application service
type TransactionService interface {

	// GetTransaction retrieves a transaction by ID
	GetTransaction(ctx context.Context, transactionID int) (*domain.Transaction, error)

	// GetTransactionsByWalletID
	GetTransactionByWalletID(ctx context.Context, walletID int, limit, offset int) ([]*domain.Transaction, int, error)

	// CreateTransaction creates a new transaction
	CreateTransaction(ctx context.Context, transaction *domain.Transaction) error

	// UpdateTransactionStatus updates the status of a transaction
	UpdateTransactionStatus(ctx context.Context, transactionID int, status domain.TransactionStatus) error

	// ReconcileFailedTransactions attempts to fix transactions
	ReconcileFailedTransactions(ctx context.Context) error
}
