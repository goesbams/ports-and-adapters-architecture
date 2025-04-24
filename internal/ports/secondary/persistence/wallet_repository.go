package persistence

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// WalletRepository defines the port for wallet data operations
type WalletRepository interface {
	// FindByID retrieves a wallet by its ID
	FindByID(ctx context.Context, id int) (*domain.Wallet, error)

	// FindByUserID retrieves all wallet for a user
	FindByUserID(ctx context.Context, UserID int) ([]*domain.Wallet, error)

	// Save creates or updates a wallet
	Save(ctx context.Context, wallet *domain.Wallet) error

	// UpdateBalance updates only the wallet balance
	UpdateBalance(ctx context.Context, walletID int, newBalance int) error

	// Delete removes a wallet
	Delete(ctx context.Context, id int) error
}
