package memory

import (
	"context"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"sync"
	"time"
)

// InMemoryTransactionRepository implements TransactionRepository interface for testing
type InMemoryTransactionRepository struct {
	mu           sync.RWMutex
	transactions map[int]*domain.Transaction
	nextID       int
}

// NewInMemoryTransactionRepository creates a new in-memory transaction repository
func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[int]*domain.Transaction),
		nextID:       1,
	}
}

func (r *InMemoryTransactionRepository) FindByID(ctx context.Context, id int) (*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transaction, exists := r.transactions[id]
	if !exists {
		return nil, nil
	}

	txCopy := *transaction
	return &txCopy, nil
}

func (r *InMemoryTransactionRepository) FindByWalletID(ctx context.Context, walletID int, limit, offset int) ([]*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var transactions []*domain.Transaction
	for _, tx := range r.transactions {
		if tx.WalletID == walletID || (tx.ToWalletID != nil && *tx.ToWalletID == walletID) {
			txCopy := *tx
			transactions = append(transactions, &txCopy)
		}
	}

	// Apply pagination
	start := offset
	if start > len(transactions) {
		return []*domain.Transaction{}, nil
	}

	end := start + limit
	if end > len(transactions) {
		end = len(transactions)
	}

	return transactions[start:end], nil
}

func (r *InMemoryTransactionRepository) FindByStatus(ctx context.Context, status domain.TransactionStatus, limit, offset int) ([]*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var transactions []*domain.Transaction
	for _, tx := range r.transactions {
		if tx.Status == status {
			txCopy := *tx
			transactions = append(transactions, &txCopy)
		}
	}

	// Apply pagination
	start := offset
	if start > len(transactions) {
		return []*domain.Transaction{}, nil
	}

	end := start + limit
	if end > len(transactions) {
		end = len(transactions)
	}

	return transactions[start:end], nil
}

func (r *InMemoryTransactionRepository) FindPendingTransactions(ctx context.Context, olderThan time.Time) ([]*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var transactions []*domain.Transaction
	for _, tx := range r.transactions {
		if tx.Status == domain.TransactionStatusPending && tx.CreatedAt.Before(olderThan) {
			txCopy := *tx
			transactions = append(transactions, &txCopy)
		}
	}

	return transactions, nil
}

func (r *InMemoryTransactionRepository) CountByWalletID(ctx context.Context, walletID int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, tx := range r.transactions {
		if tx.WalletID == walletID || (tx.ToWalletID != nil && *tx.ToWalletID == walletID) {
			count++
		}
	}

	return count, nil
}

func (r *InMemoryTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if transaction.ID == 0 {
		transaction.ID = r.nextID
		r.nextID++
	}

	txCopy := *transaction
	r.transactions[transaction.ID] = &txCopy

	return nil
}

func (r *InMemoryTransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transactions[transaction.ID]; !exists {
		return fmt.Errorf("transaction not found: %d", transaction.ID)
	}

	txCopy := *transaction
	r.transactions[transaction.ID] = &txCopy

	return nil
}

func (r *InMemoryTransactionRepository) UpdateStatus(ctx context.Context, id int, status domain.TransactionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	transaction, exists := r.transactions[id]
	if !exists {
		return fmt.Errorf("transaction not found: %d", id)
	}

	transaction.Status = status
	transaction.UpdatedAt = time.Now()

	if status == domain.TransactionStatusCompleted {
		now := time.Now()
		transaction.CompletedAt = &now
	}

	return nil
}
