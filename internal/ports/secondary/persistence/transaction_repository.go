package persistence

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
	"time"
)

// TransactionRepostory defines the port for transaction data operations
type TransactionRepository interface {
	// FindByID retrieves a transaction by its ID
	FindByID(ctx context.Context, id int) (*domain.Transaction, error)

	// FindByWalletID retrieves all transactions for a wallet
	FindByWalletID(ctx context.Context, walletID int, limit, offset int) ([]*domain.Transaction, error)

	// FindByStatus retrieves transactions by status with optional pagination
	FindByStatus(ctx context.Context, status domain.TransactionStatus, limit, offset int) ([]*domain.Transaction, error)

	// FindPendingTransactions retrieves pending transactions older than a specified time
	FindPendingTransactions(ctx context.Context, olderThan time.Time) ([]*domain.Transaction, error)

	// CountByWalletID counts all transactions for a wallet
	CountByWalletID(ctx context.Context, walletID int) (int, error)

	// Create saves a new transaction
	Create(ctx context.Context, transaction *domain.Transaction)

	// Update updates an existing transaction
	Update(ctx context.Context, transaction *domain.Transaction)

	// UpdateStatus updates only the status of a transaction
	UpdateStatus(ctx context.Context, id int, status domain.TransactionStatus) error
}
