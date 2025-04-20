package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidTransactionAmount = errors.New("transaction amount must be greater than zero")
	ErrInvalidTransactionType   = errors.New("invalid transaction type")
	ErrTransactionFailed        = errors.New("transaction failed")
)

type TransactionType string
type TransactionStatus string

// common transaction types
const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
)

// common transaction statuses
const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
)

// transaction represents a financial transaction in the e-wallet system
type Transaction struct {
	ID          int               `json:"id"`
	WalletID    int               `json:"wallet_id"`
	Type        TransactionType   `json:"transaction_type"`
	Amount      int               `json:"amount"`
	Status      TransactionStatus `json:"transaction_status"`
	Reference   string            `json:"reference,omitempty"`
	Description string            `json:"description,omitempty"`
	ToWalletID  *int              `json:"to_wallet_id,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}

// NewTransaction creates a new transaction
func NewTransaction(walletID int, txType TransactionType, amount int, description string) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidTransactionAmount
	}

	// validate transaction type
	if txType != TransactionTypeDeposit &&
		txType != TransactionTypeWithdrawal &&
		txType != TransactionTypeTransfer {
		return nil, ErrInvalidTransactionType
	}

	now := time.Now()
	return &Transaction{
		WalletID:    walletID,
		Type:        txType,
		Amount:      amount,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// NewTransferTransaction creates a new transfer transaction
func NewTransferTransaction(fromWalletId, ToWalletID int, amount int, description string) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidTransactionAmount
	}

	now := time.Now()
	return &Transaction{
		WalletID:    fromWalletId,
		ToWalletID:  &ToWalletID,
		Type:        TransactionTypeTransfer,
		Amount:      amount,
		Status:      TransactionStatusPending,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// complete marks a transaction as completed
func (t *Transaction) Complete() {
	now := time.Now()
	t.Status = TransactionStatusCompleted
	t.CompletedAt = &now
	t.UpdatedAt = now
}

// Fail marks a transaction as failed
func (t *Transaction) Fail() {
	t.Status = TransactionStatusFailed
	t.UpdatedAt = time.Now()
}

// IsPending Checks if a transaction is pending
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending
}

// IsCompleted checks if a transaction is completed
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

// IsFailed checks if a transaction is failed
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}
