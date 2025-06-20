package tests

import (
	"context"
	"ports-and-adapters-architecture/internal/adapters/persistence/memory"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/usecase"
	"testing"
)

func TestTransactionService_CreateTransaction(t *testing.T) {
	// Setup
	ctx := context.Background()
	walletRepo := memory.NewInMemoryWalletRepository()
	transactionRepo := memory.NewInMemoryTransactionRepository()

	// Create a test wallet
	wallet := domain.NewWallet(1, "USD", "Test wallet")
	wallet.ID = 1
	_ = walletRepo.Save(ctx, wallet)

	// Create transaction service
	transactionService := usecase.NewTransactionService(
		transactionRepo,
		walletRepo,
		nil,
		nil,
	)

	// Test transaction creation
	transaction := &domain.Transaction{
		WalletID:    1,
		Type:        domain.TransactionTypeDeposit,
		Amount:      100,
		Description: "Test transaction",
	}

	err := transactionService.CreateTransaction(ctx, transaction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if transaction.ID == 0 {
		t.Error("expected transaction ID to be set")
	}

	if transaction.Status != domain.TransactionStatusPending {
		t.Errorf("expected pending status, got %s", transaction.Status)
	}
}

func TestTransactionService_UpdateTransactionStatus(t *testing.T) {
	// Setup
	ctx := context.Background()
	walletRepo := memory.NewInMemoryWalletRepository()
	transactionRepo := memory.NewInMemoryTransactionRepository()

	// Create a test wallet and transaction
	wallet := domain.NewWallet(1, "USD", "Test wallet")
	wallet.ID = 1
	_ = walletRepo.Save(ctx, wallet)

	transaction, _ := domain.NewTransaction(1, domain.TransactionTypeDeposit, 100, "Test")
	transaction.ID = 1
	transaction.Status = domain.TransactionStatusPending
	_ = transactionRepo.Create(ctx, transaction)

	// Create transaction service
	transactionService := usecase.NewTransactionService(
		transactionRepo,
		walletRepo,
		nil,
		nil,
	)

	// Test status update
	err := transactionService.UpdateTransactionStatus(ctx, 1, domain.TransactionStatusCompleted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify status
	updatedTx, _ := transactionRepo.FindByID(ctx, 1)
	if updatedTx.Status != domain.TransactionStatusCompleted {
		t.Errorf("expected completed status, got %s", updatedTx.Status)
	}
}