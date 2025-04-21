package usecase

import (
	"errors"
	"ports-and-adapters-architecture/internal/ports"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrUserNotFound        = errors.New("user not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be greater than zeror")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrTransferFailed      = errors.New("transfer failed")
	ErrWalletAlreadyExists = errors.New("wallet already exists for this user and currency")
)

// WalletService defines the application logic for wallet operations
type WalletService struct {
	walletRepo      ports.WalletRepository
	userRepo        ports.UserRepository
	transactionRepo ports.TransactionRepository
	eventPublisher  ports.EventPublisher
	cache           ports.Cache
}
