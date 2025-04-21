package ports

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// TransactionRepostory defines the port for transaction data operations
type TransactionRepository interface {
	// FindByID retrieves a transaction by its ID
	FindByID(ctx context.Context, id int) (*domain.Transaction, error)

	// FindByWalletID retrieves all transactions for a wallet
	FindByWalletID(ctx context.Context, walletID int, limit, offset int) ([]*domain.Transaction, error)

	// CountByWalletID counts all transactions for a wallet
	CountByWalletID(ctx context.Context, walletID int) (int, error)

	// Create saves a new transaction
	Create(ctx context.Context, transaction *domain.Transaction)

	// Update updates an existing transaction
	Update(ctx context.Context, transaction *domain.Transaction)
}
