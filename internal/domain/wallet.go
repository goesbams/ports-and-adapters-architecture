package domain

import (
	"errors"
	"time"
)

// common errors for wallet operations
var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrWalletNotActive     = errors.New("wallet is not active")
)

type Wallet struct {
	ID           int          `json:"id"`
	UserID       int          `json:"user_id"`
	Balance      int          `json:"balance"`
	CurrencyCode string       `json:"currency_code"`
	Description  string       `json:"description"`
	Status       WalletStatus `json:"Status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type WalletStatus string

// common status for wallet
const (
	WalletStatusActive   WalletStatus = "ACTIVE"
	WalletStatusInactive WalletStatus = "INACTIVE"
)

func NewWallet(userID int, currencyCode, description string) *Wallet {
	now := time.Now()

	return &Wallet{
		UserID:       userID,
		Balance:      0,
		CurrencyCode: currencyCode,
		Description:  description,
		Status:       WalletStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Credit adds funds to wallet
func (w *Wallet) Credit(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if w.Status != WalletStatusActive {
		return ErrWalletNotActive
	}

	if w.Balance < amount {
		return ErrInsufficientBalance
	}

	w.Balance += amount
	w.UpdatedAt = time.Now()

	return nil
}

// Debit remove funds from wallet
func (w *Wallet) Debit(amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if w.Status != WalletStatusActive {
		return ErrWalletNotActive
	}

	if w.Balance < amount {
		return ErrInsufficientBalance
	}

	w.Balance -= amount
	w.UpdatedAt = time.Now()

	return nil
}

func (w *Wallet) IsActive() bool {
	return w.Status == WalletStatusActive
}
