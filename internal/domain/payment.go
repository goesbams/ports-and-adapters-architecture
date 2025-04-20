package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidPaymentAmount    = errors.New("payment amount must be greater than zero")
	ErrInvalidPaymentStatus    = errors.New("invalid payment status")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed")
)

// PaymentProvider represents different payment providers
type PaymentProvider string

// common payment providers
const (
	PaymentProviderMidtrans PaymentProvider = "MIDTRANS"
	PaymentProviderDoku     PaymentProvider = "DOKU"
	PaymentProviderStripe   PaymentProvider = "STRIPE"
)

type PaymentStatus string

// common payment statuses
const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
)

// payment represents a payment entitiy in e-wallet system
type Payment struct {
	ID            int                    `json:"id"`
	TransactionID int                    `json:"transaction_id"`
	Amount        int                    `json:"amount"`
	Provider      PaymentProvider        `json:"provider"`
	Status        PaymentStatus          `json:"status"`
	ExternalID    string                 `json:"external_id,omitempty"`
	PaymentURL    string                 `json:"payment_url,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

// NewPayment creates a new payment
func NewPayment(transactionID int, amount int, provider PaymentProvider, description string) (*Payment, error) {
	if amount <= 0 {
		return nil, ErrInvalidPaymentAmount
	}

	now := time.Now()
	return &Payment{
		TransactionID: transactionID,
		Amount:        amount,
		Provider:      provider,
		Status:        PaymentStatusPending,
		Description:   description,
		Details:       make(map[string]interface{}),
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Complete marks a payment as completed
func (p *Payment) Complete() error {
	if p.Status != PaymentStatusPending {
		return ErrPaymentAlreadyProcessed
	}

	now := time.Now()
	p.Status = PaymentStatusCompleted
	p.CompletedAt = &now
	p.UpdatedAt = now

	return nil
}

// Fail marks a payment as Failed
func (p *Payment) Fail() error {
	if p.Status != PaymentStatusPending {
		return ErrPaymentAlreadyProcessed
	}

	p.Status = PaymentStatusFailed
	p.UpdatedAt = time.Now()

	return nil
}

// Cancel marks a payment as cancelled
func (p *Payment) Cancel() error {
	if p.Status != PaymentStatusPending {
		return ErrPaymentAlreadyProcessed
	}

	p.Status = PaymentStatusCancelled
	p.UpdatedAt = time.Now()

	return nil
}
