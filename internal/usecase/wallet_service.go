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
	ErrTransactionNotFound = errors.New("transaction not found")
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

// GetWalletsByUserId retrieves all wallets for a user
func (s *WalletService) GetWalletsByUserID(ctx context.Context, userID int) ([]*domain.Wallet, error) {
	// Verify the user exists
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	wallets, err := s.walletRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find wallets: %w", err)
	}

	return wallets, nil
}

// UpdateWalletStatus updates the status of a wallet
func (s *WalletService) UpdateWalletStatus(ctx context.Context, walletID int, status domain.WalletStatus) error {
	// Get the wallet first to verify it exists
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to find wallet: %w", err)
	}

	if wallet == nil {
		return ErrWalletNotFound
	}

	// Update the status
	err = s.walletRepo.UpdateStatus(ctx, walletID, status)
	if err != nil {
		return fmt.Errorf("failed to update wallet status: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("wallet:%d", walletID)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish wallet status updated event
	event := infrastructure.Event{
		Type: "wallet.status_updated",
		Payload: map[string]interface{}{
			"wallet_id": walletID,
			"status":    string(status),
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

	return nil
}

// Deposit adds funds to a wallet
func (s *WalletService) Deposit(ctx context.Context, walletID int, amount int, description string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get wallet
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	if wallet == nil {
		return nil, ErrWalletNotFound
	}

	// Ensure wallet is active
	if !wallet.IsActive() {
		return nil, domain.ErrWalletNotActive
	}

	// Create pending transaction
	transaction, err := domain.NewTransaction(walletID, domain.TransactionTypeDeposit, amount, description)
	if err != nil {
		return nil, err
	}

	transaction.Status = domain.TransactionStatusPending

	// Save the transaction
	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %2", err)
	}

	// Credit the wallet
	err = wallet.Credit(amount)
	if err != nil {
		// Mark transaction as failed
		transaction.Fail()
		_ = s.transactionRepo.Update(ctx, transaction)
	}

	// Update wallet in database
	err = s.walletRepo.Save(ctx, wallet)
	if err != nil {
		// Mark transaction as failed if wallet update fails
		transaction.Fail()
		_ = s.transactionRepo.Update(ctx, transaction)
		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Mark transaction as completed
	transaction.Complete()
	err = s.transactionRepo.Update(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("wallet:%d", walletID)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish deposit event
	event := infrastructure.Event{
		Type: "wallet.deposit",
		Payload: map[string]interface{}{
			"wallet_id":      wallet.ID,
			"transaction_id": transaction.ID,
			"amount":         amount,
			"new_balance":    wallet.Balance,
		},
	}

	// Non-blocking event publishing
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if s.eventPublisher != nil {
			_ = s.eventPublisher.Publish(ctx, "transactions", event)
		}
	}()

	return transaction, nil
}

// Withdraw removes funds from a wallet
func (s *WalletService) Withdraw(ctx context.Context, walletID int, amount int, description string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get wallet
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	if wallet == nil {
		return nil, ErrWalletNotFound
	}

	// Check if wallet has sufficient balance
	if wallet.Balance < amount {
		return nil, ErrInsufficientBalance
	}

	// Create pending transaction
	transaction, err := domain.NewTransaction(walletID, domain.TransactionTypeWithdrawal, amount, description)
	if err != nil {
		return nil, err
	}

	transaction.Status = domain.TransactionStatusPending

	// Save the transaction
	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Debit the wallet (will perform additional validation)
	err = wallet.Debit(amount)
	if err != nil {
		// Mark transaction as failed if debit fails
		transaction.Fail()
		_ = s.transactionRepo.Update(ctx, transaction)

		return nil, err
	}

	// Update wallet in database
	err = s.walletRepo.Save(ctx, wallet)
	if err != nil {
		// Mark transaction as failed if wallet update fails
		transaction.Fail()
		_ = s.transactionRepo.Update(ctx, transaction)

		return nil, fmt.Errorf("failed to update wallet balance: %w", err)
	}

	// Mark transaction as completed
	transaction.Complete()
	err = s.transactionRepo.Update(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("wallet:%d", walletID)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish withdrawal event
	event := infrastructure.Event{
		Type: "wallet.withdrawal",
		Payload: map[string]interface{}{
			"wallet_id":      wallet.ID,
			"transaction_id": transaction.ID,
			"amount":         amount,
			"new_balance":    wallet.Balance,
		},
	}

	// Non-blocking event publishing
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if s.eventPublisher != nil {
			_ = s.eventPublisher.Publish(ctx, "transactions", event)
		}
	}()

	return transaction, nil
}

// Transfer transfers funds from one wallet to another
func (s *WalletService) Transfer(ctx context.Context, fromWalletID int, toWalletID int, amount int, description string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get source wallet
	fromWallet, err := s.walletRepo.FindByID(ctx, fromWalletID)
	if err != nil {
		return nil, fmt.Errorf("failed to find source wallet: %w", err)
	}

	if fromWallet == nil {
		return nil, ErrWalletNotFound
	}

	// Get destination wallet
	toWallet, err := s.walletRepo.FindByID(ctx, toWalletID)
	if err != nil {
		return nil, fmt.Errorf("failed to find destination wallet: %w", err)
	}

	if toWallet == nil {
		return nil, ErrWalletNotFound
	}

	// Check if wallets have the same currency
	if fromWallet.CurrencyCode != toWallet.CurrencyCode {
		return nil, errors.New("cannot transfer between wallets with different currencies")
	}

	// Create transfer transaction
	transaction, err := domain.NewTransferTransaction(fromWalletID, toWalletID, amount, description)
	if err != nil {
		return nil, err
	}

	// Save the transaction
	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Debit from source wallet
	err = fromWallet.Debit(amount)
	if err != nil {
		// Mark transaction as failed if debit fails
		transaction.Fail()
		_ = s.transactionRepo.Update(ctx, transaction)

		return nil, err
	}

	// Credit desitination wallet
	err = toWallet.Credit(amount)
	if err != nil {
		// Mark transaction as failed if credit fails
		transaction.Fail()
		_ = s.transactionRepo.Updated(ctx, transaction)
		return nil, err
	}

	// Update wallets in database (ideally in a transaction)
	err = s.walletRepo.Save(ctx, fromWallet)
	if err != nil {
		// Mark transaction as failed if source wallet update fails
		transaction.Fail()
		_ = s.transactionRepo.Update(ctx, transaction)
		return nil, fmt.Errorf("failed to update source wallet: %w", err)
	}

	err = s.walletRepo.Save(ctx, toWallet)
	if err != nil {
		// This is a critical error - money has been deducted but not credited
		// In a real system, this would require more sophisticated recovery
		transaction.Fail()
		_ = s.transactionRepo.Update(ctx, transaction)

		// Try to refund the source wallet
		fromWallet.Credit(amount)
		_ = s.walletRepo.Save(ctx, fromWallet)

		return nil, fmt.Errorf("failed to update destination wallet: %w", err)
	}

	// Mark transaction as completed
	transaction.Complete()
	err = s.transactionRepo.Update(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Invalidate cache for both wallets
	if s.cache != nil {
		_ = s.cache.Delete(ctx, fmt.Sprintf("wallet:%d", fromWalletID))
		_ = s.cache.Delete(ctx, fmt.Sprintf("wallet:%d", toWalletID))
	}

	// Publish transfer event
	event := infrastructure.Event{
		Type: "wallet.transfer",
		Payload: map[string]interface{}{
			"transaction_id": transaction.ID,
			"from_wallet_id": fromWalletID,
			"to_wallet_id":   toWalletID,
			"amount":         amount,
			"from_balance":   fromWallet.Balance,
			"to_balance":     toWallet.Balance,
		},
	}

	// Non-blocking event publishing
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if s.eventPublisher != nil {
			_ = s.eventPublisher.Publish(ctx, "transactions", event)
		}
	}()

	return transaction, nil
}

// GetTransactionHistory retrieves transaction history for a wallet
func (s *WalletService) GetTransactionHistory(
	ctx context.Context,
	walletID int,
	limit, offset int,
) ([]*domain.Transaction, int, error) {
	// Verify wallet exists
	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find wallet: %w", err)
	}

	if wallet == nil {
		return nil, 0, ErrWalletNotFound
	}

	// Get transactions
	transactions, err := s.transactionRepo.FindByWalletID(ctx, walletID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find transactions: %w", err)
	}

	// Get total count
	totalCount, err := s.transactionRepo.CountByWalletID(ctx, walletID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return transactions, totalCount, nil
}

// GetBalance gets the current balance of a wallet
func (s *WalletService) GetBalance(ctx context.Context, walletID int) (int, string, error) {
	wallet, err := s.GetWallet(ctx, walletID)
	if err != nil {
		return 0, "", err
	}

	return wallet.Balance, wallet.CurrencyCode, nil
}
