package tests

import (
	"context"
	"ports-and-adapters-architecture/internal/adapters/persistence/memory"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/usecase"
	"testing"
)

func TestWalletService_CreateWallet(t *testing.T) {
	// Setup
	ctx := context.Background()
	userRepo := memory.NewInMemoryUserRepository()
	walletRepo := memory.NewInMemoryWalletRepository()
	transactionRepo := memory.NewInMemoryTransactionRepository()
	
	// Create a test user
	user := domain.NewUser("Test User", "test@example.com", "+1234567890")
	user.ID = 1
	_ = userRepo.Save(ctx, user)

	// Create wallet service
	walletService := usecase.NewWalletService(
		walletRepo,
		userRepo,
		transactionRepo,
		nil, // No event publisher for tests
		nil, // No cache for tests
	)

	// Test cases
	tests := []struct {
		name         string
		userID       int
		currencyCode string
		description  string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "Valid wallet creation",
			userID:       1,
			currencyCode: "USD",
			description:  "My USD wallet",
			wantErr:      false,
		},
		{
			name:         "Invalid user ID",
			userID:       999,
			currencyCode: "USD",
			description:  "Should fail",
			wantErr:      true,
			errMsg:       "user not found",
		},
		{
			name:         "Duplicate currency",
			userID:       1,
			currencyCode: "USD",
			description:  "Another USD wallet",
			wantErr:      true,
			errMsg:       "wallet already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := walletService.CreateWallet(ctx, tt.userID, tt.currencyCode, tt.description)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if wallet == nil {
				t.Error("expected wallet but got nil")
				return
			}

			if wallet.UserID != tt.userID {
				t.Errorf("expected user ID %d, got %d", tt.userID, wallet.UserID)
			}

			if wallet.CurrencyCode != tt.currencyCode {
				t.Errorf("expected currency %s, got %s", tt.currencyCode, wallet.CurrencyCode)
			}

			if wallet.Balance != 0 {
				t.Errorf("expected initial balance 0, got %d", wallet.Balance)
			}
		})
	}
}

func TestWalletService_Deposit(t *testing.T) {
	// Setup
	ctx := context.Background()
	userRepo := memory.NewInMemoryUserRepository()
	walletRepo := memory.NewInMemoryWalletRepository()
	transactionRepo := memory.NewInMemoryTransactionRepository()

	// Create a test user and wallet
	user := domain.NewUser("Test User", "test@example.com", "+1234567890")
	user.ID = 1
	_ = userRepo.Save(ctx, user)

	wallet := domain.NewWallet(1, "USD", "Test wallet")
	wallet.ID = 1
	_ = walletRepo.Save(ctx, wallet)

	// Create wallet service
	walletService := usecase.NewWalletService(
		walletRepo,
		userRepo,
		transactionRepo,
		nil,
		nil,
	)

	// Test deposit
	transaction, err := walletService.Deposit(ctx, 1, 100, "Test deposit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction == nil {
		t.Fatal("expected transaction but got nil")
	}

	if transaction.Amount != 100 {
		t.Errorf("expected amount 100, got %d", transaction.Amount)
	}

	if transaction.Type != domain.TransactionTypeDeposit {
		t.Errorf("expected deposit type, got %s", transaction.Type)
	}

	// Verify wallet balance
	updatedWallet, _ := walletRepo.FindByID(ctx, 1)
	if updatedWallet.Balance != 100 {
		t.Errorf("expected balance 100, got %d", updatedWallet.Balance)
	}
}

func TestWalletService_Transfer(t *testing.T) {
	// Setup
	ctx := context.Background()
	userRepo := memory.NewInMemoryUserRepository()
	walletRepo := memory.NewInMemoryWalletRepository()
	transactionRepo := memory.NewInMemoryTransactionRepository()

	// Create test users
	user1 := domain.NewUser("User 1", "user1@example.com", "+1111111111")
	user1.ID = 1
	_ = userRepo.Save(ctx, user1)

	user2 := domain.NewUser("User 2", "user2@example.com", "+2222222222")
	user2.ID = 2
	_ = userRepo.Save(ctx, user2)

	// Create wallets
	wallet1 := domain.NewWallet(1, "USD", "Wallet 1")
	wallet1.ID = 1
	wallet1.Balance = 200
	_ = walletRepo.Save(ctx, wallet1)

	wallet2 := domain.NewWallet(2, "USD", "Wallet 2")
	wallet2.ID = 2
	_ = walletRepo.Save(ctx, wallet2)

	// Create wallet service
	walletService := usecase.NewWalletService(
		walletRepo,
		userRepo,
		transactionRepo,
		nil,
		nil,
	)

	// Test transfer
	transaction, err := walletService.Transfer(ctx, 1, 2, 50, "Test transfer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction == nil {
		t.Fatal("expected transaction but got nil")
	}

	// Verify balances
	updatedWallet1, _ := walletRepo.FindByID(ctx, 1)
	if updatedWallet1.Balance != 150 {
		t.Errorf("expected sender balance 150, got %d", updatedWallet1.Balance)
	}

	updatedWallet2, _ := walletRepo.FindByID(ctx, 2)
	if updatedWallet2.Balance != 50 {
		t.Errorf("expected receiver balance 50, got %d", updatedWallet2.Balance)
	}
}