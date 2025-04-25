package primary

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// PaymentRequest represent a request to proccess a payment
type PaymentRequest struct {
	WalletID        int                    `json:"wallet_id"`
	Amount          int                    `json:"amount"`
	Description     string                 `json:"description"`
	PaymentProvider domain.PaymentProvider `json:"payment_provider"`
	RedirectURL     string                 `json:"redirect_url"`
	CallbackURL     string                 `json:"callback_url"`
}

// PaymentService defines the contract for payment application service
type PaymentService interface {

	// ProcessPayment inititates a payment through a payment gateway
	ProcessPayment(ctx context.Context, req PaymentRequest) (*domain.Payment, error)

	// VerifyPayment check the status of a payment and updates
	VerifyPayment(ctx context.Context, paymentID int) (*domain.Payment, error)

	// CancelPayment cancel a pending payment
	CancelPayment(ctx context.Context, paymentID int) error

	// GetPaymentID retrieves a payment by its ID
	GetPaymentID(ctx context.Context, paymentID int) (*domain.Payment, error)

	// GetPaymentsByTransactionID retrieves all payments for a transaction
	GetPaymentsByTransactionID(ctx context.Context, transactionID int) ([]*domain.Payment, error)
}
