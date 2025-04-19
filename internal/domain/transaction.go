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
