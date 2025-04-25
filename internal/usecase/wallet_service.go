package usecase

import (
	"context"
	"errors"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/ports"
	"ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
	"ports-and-adapters-architecture/internal/ports/secondary/persistence"
	"time"
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

// NewWalletService creates a new wallet service
func NewWalletService(
	walletRepo persistence.WalletRepository,
	userRepo persistence.UserRepository,
	transactionRepo persistence.TransactionRepository,
	eventPublisher infrastructure.EventPublisher,
	cache infrastructure.Cache,
) *WalletService {
	return &WalletService{
		walletRepo:      walletRepo,
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		eventPublisher:  eventPublisher,
		cache:           cache,
	}
}

// CreateWallet creates a new wallet for a user
func (s *WalletService) CreateWallet(ctx context.Context, userID int, currencyCode, description string) (*domain.Wallet, error) {
	// Verify the user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if user already has a wallet with this currency
	existingWallets, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing wallets: %w", err)
	}

	for _, wallet := range existingWallets {
		if wallet.currencyCode == currencyCode {
			return nil, ErrWalletAlreadyExists
		}
	}

	// Create and save new wallet
	wallet := domain.NewWallet(userID, currencyCode, description)
	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to save wallet: %w", err)
	}

	// Publish wallet created event
	event := infrastructure.Event{
		Type: "wallet.created",
		Payload: map[string]interface{}{
			"wallet_id":       wallet.ID,
			"user_id":         wallet.UserID,
			"currency_code":   wallet.CurrencyCode,
			"initial_balance": wallet.Balance,
		},
	}

	// Non-blocking event publishing
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if s.eventPublisher != nil {
			_ = s.eventPublisher.Publish(ctx, "wallets", event)
		}
	}()

	return wallet, nil
}

// GetWallet retrieves a wallet by ID
func (s *WalletService) GetWallet(ctx context.Context, walletID int) (*domain.Wallet, error) {
	// Try to get from cache first
	var wallet *domain.Wallet

	if s.cache != nil {
		cacheKey := fmt.Sprintf("wallet:%d", walletID)
		err := s.cache.GetObject(ctx, cacheKey, &wallet)
		if err == nil && wallet != nil {
			return wallet, nil
		}
	}

	// Fetch from database
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	if wallet == nil {
		return nil, ErrWalletNotFound
	}

	// Cache the wallet for future requests
	if s.cache != nil {
		cacheKey := fmt.Sprintf("wallet:%d", walletID)
		_ = s.cache.SetObject(ctx, cacheKey, wallet, 5*time.Minute)
	}

	return wallet, nil
}
