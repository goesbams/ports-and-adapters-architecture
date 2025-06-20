package memory

import (
	"context"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"sync"
)

// InMemoryWalletRepository implements WalletRepository interface for testing
type InMemoryWalletRepository struct {
	mu      sync.RWMutex
	wallets map[int]*domain.Wallet
	nextID  int
}

// NewInMemoryWalletRepository creates a new in-memory wallet repository
func NewInMemoryWalletRepository() *InMemoryWalletRepository {
	return &InMemoryWalletRepository{
		wallets: make(map[int]*domain.Wallet),
		nextID:  1,
	}
}

func (r *InMemoryWalletRepository) FindByID(ctx context.Context, id int) (*domain.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wallet, exists := r.wallets[id]
	if !exists {
		return nil, nil
	}

	walletCopy := *wallet
	return &walletCopy, nil
}

func (r *InMemoryWalletRepository) FindByUserID(ctx context.Context, userID int) ([]*domain.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var wallets []*domain.Wallet
	for _, wallet := range r.wallets {
		if wallet.UserID == userID {
			walletCopy := *wallet
			wallets = append(wallets, &walletCopy)
		}
	}

	return wallets, nil
}

func (r *InMemoryWalletRepository) Save(ctx context.Context, wallet *domain.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if wallet.ID == 0 {
		wallet.ID = r.nextID
		r.nextID++
	}

	walletCopy := *wallet
	r.wallets[wallet.ID] = &walletCopy

	return nil
}

func (r *InMemoryWalletRepository) UpdateBalance(ctx context.Context, walletID int, newBalance int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	wallet, exists := r.wallets[walletID]
	if !exists {
		return fmt.Errorf("wallet not found: %d", walletID)
	}

	wallet.Balance = newBalance
	return nil
}

func (r *InMemoryWalletRepository) UpdateStatus(ctx context.Context, walletID int, status domain.WalletStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	wallet, exists := r.wallets[walletID]
	if !exists {
		return fmt.Errorf("wallet not found: %d", walletID)
	}

	wallet.Status = status
	return nil
}

func (r *InMemoryWalletRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.wallets, id)
	return nil
}
