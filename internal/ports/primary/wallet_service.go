package primary

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// WalletService defines the contract for wallet application service
type WalletService interface {
	// CreateWallet creates a new wallet for a user
	CreateWallet(ctx context.Context, userID int, currencyCode, description string) (*domain.Wallet, error)

	// GetWallet retrieves a wallet by ID
	GetWallet(ctx context.Context, walletID int) (*domain.Wallet, error)

	// GetWalletByUserID retrieves all wallets for a user
	GetWalletsByUserID(ctx context.Context, userID int) ([]*domain.Wallet, error)

	// Deposit add funds to a wallet
	Deposit(ctx context.Context, walletID int, amount int, description string) (*domain.Transaction, error)

	// Withdraw removes funds from a wallet
	Withdraw(ctx context.Context, walletID int, amount int, description string) (*domain.Transaction, error)

	// Transfer transfers fund from one wallet to another
	Transfer(
		ctx context.Context,
		fromWalletID int,
		toWalletID int,
		amount int,
		description string,
	) (*domain.Transaction, error)

	// GetTransactionHistory retrieves transaction history for a wallet
	GetTransactionHistory(
		ctx context.Context,
		walletID int,
		limit, offset int,
	) ([]*domain.Transaction, int, error)

	// GetBalance gets the current balance of a wallet
	GetBalance(ctx context.Context, walletID int) (int, string, error)
}
