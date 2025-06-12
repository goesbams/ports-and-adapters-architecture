package usecase

import (
	"context"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
	"ports-and-adapters-architecture/internal/ports/secondary/persistence"
	"time"
)

// TransactionService implements the transaction application service
type TransactionService struct {
	transactionRepo persistence.TransactionRepository
	walletRepo      persistence.WalletRepository
	eventPublisher  infrastructure.EventPublisher
	cache           infrastructure.Cache
}

// NewTransactionService creates a new transaction service
func NewTransactionService(
	transactionRepo persistence.TransactionRepository,
	walletRepo persistence.WalletRepository,
	eventPublisher infrastructure.EventPublisher,
	cache infrastructure.Cache,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		eventPublisher:  eventPublisher,
		cache:           cache,
	}
}

// GetTransaction retrieves a transaction by ID
func (s *TransactionService) GetTransaction(ctx context.Context, transactionID int) (*domain.Transaction, error) {
	// Try to get from cache first
	var transaction *domain.Transaction

	if s.cache != nil {
		cacheKey := fmt.Sprintf("transaction:%d", transactionID)
		err := s.cache.GetObject(ctx, cacheKey, &transaction)
		if err == nil && transaction != nil {
			return transaction, nil
		}
	}

	// Fetch from database
	transaction, err := s.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}

	if transaction == nil {
		return nil, ErrTransactionNotFound
	}

	// Cache the transaction for future requests
	if s.cache != nil {
		cacheKey := fmt.Sprintf("transaction:%d", transactionID)
		_ = s.cache.SetObject(ctx, cacheKey, transaction, 5*time.Minute)
	}

	return transaction, nil
}

// GetTransactionsByWalletID retrieves transactions for a wallet with pagination
func (s *TransactionService) GetTransactionByWalletID(ctx context.Context, walletID int, limit, offset int) ([]*domain.Transaction, int, error) {
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

// CreateTransaction creates a new transaction
func (s *TransactionService) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	// Validate transaction
	if transaction.Amount <= 0 {
		return ErrInvalidAmount
	}

	// Verify wallet exists
	wallet, err := s.walletRepo.FindByID(ctx, transaction.WalletID)
	if err != nil {
		return fmt.Errorf("failed to find wallet: %w", err)
	}

	if wallet == nil {
		return ErrWalletNotFound
	}

	// For transfers, verify destination wallet
	if transaction.Type == domain.TransactionTypeTransfer && transaction.ToWalletID != nil {
		toWallet, err := s.walletRepo.FindByID(ctx, *transaction.ToWalletID)
		if err != nil {
			return fmt.Errorf("failed to find destination wallet: %w", err)
		}

		if toWallet == nil {
			return fmt.Errorf("destination wallet not found")
		}
	}

	// Set initial status
	transaction.Status = domain.TransactionStatusPending

	// Create transaction
	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Publish transaction created event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "transaction.created",
			Payload: map[string]interface{}{
				"transaction_id": transaction.ID,
				"wallet_id":      transaction.WalletID,
				"type":           string(transaction.Type),
				"amount":         transaction.Amount,
				"status":         string(transaction.Status),
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "transactions", event)
		}()
	}

	return nil
}

// UpdateTransactionStatus updates the status of a transaction
func (s *TransactionService) UpdateTransactionStatus(ctx context.Context, transactionID int, status domain.TransactionStatus) error {
	// Get transaction
	transaction, err := s.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("failed to find transaction: %w", err)
	}

	if transaction == nil {
		return ErrTransactionNotFound
	}

	// Update status
	err = s.transactionRepo.UpdateStatus(ctx, transactionID, status)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("transaction:%d", transactionID)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish status update event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "transaction.status_updated",
			Payload: map[string]interface{}{
				"transaction_id": transactionID,
				"old_status":     string(transaction.Status),
				"new_status":     string(status),
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "transactions", event)
		}()
	}

	return nil
}

// ReconcileFailedTransactions attempts to fix failed transactions
func (s *TransactionService) ReconcileFailedTransactions(ctx context.Context) error {
	// Find old pending transactions (older than 30 minutes)
	cutoffTime := time.Now().Add(-30 * time.Minute)

	pendingTransactions, err := s.transactionRepo.FindPendingTransactions(ctx, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to find pending transactions: %w", err)
	}

	reconciled := 0
	failed := 0

	for _, transaction := range pendingTransactions {
		// Mark old pending transactions as failed
		err := s.transactionRepo.UpdateStatus(ctx, transaction.ID, domain.TransactionStatusFailed)
		if err != nil {
			failed++
			continue
		}

		reconciled++

		// Publish reconciliation event
		if s.eventPublisher != nil {
			event := infrastructure.Event{
				Type: "transaction.reconciled",
				Payload: map[string]interface{}{
					"transaction_id": transaction.ID,
					"reason":         "timeout",
				},
			}

			go func(txID int) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = s.eventPublisher.Publish(ctx, "reconciliation", event)
			}(transaction.ID)
		}
	}

	if failed > 0 {
		return fmt.Errorf("reconciled %d transactions, but %d failed", reconciled, failed)
	}

	return nil
}
