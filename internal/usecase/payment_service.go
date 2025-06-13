package usecase

import (
	"context"
	"errors"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/ports/primary"
	"ports-and-adapters-architecture/internal/ports/secondary/external"
	"ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
	"ports-and-adapters-architecture/internal/ports/secondary/persistence"
	"time"
)

var (
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentGatewayFailed = errors.New("payment gateway failed")
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
)

// PaymentService implements the payment application service
type PaymentService struct {
	paymentRepo     persistence.PaymentRepository
	walletRepo      persistence.WalletRepository
	transactionRepo persistence.TransactionRepository
	gateways        map[domain.PaymentProvider]external.PaymentGateway
	eventPublisher  infrastructure.EventPublisher
	cache           infrastructure.Cache
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo persistence.PaymentRepository,
	walletRepo persistence.WalletRepository,
	transactionRepo persistence.TransactionRepository,
	eventPublisher infrastructure.EventPublisher,
	cache infrastructure.Cache,
) *PaymentService {
	return &PaymentService{
		paymentRepo:     paymentRepo,
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		gateways:        make(map[domain.PaymentProvider]external.PaymentGateway),
		eventPublisher:  eventPublisher,
		cache:           cache,
	}
}

// RegisterGateway registers a payment gateway
func (s *PaymentService) RegisterGateway(provider domain.PaymentProvider, gateway external.PaymentGateway) {
	s.gateways[provider] = gateway
}

// ProcessPayment initiates a payment through a payment gateway
func (s *PaymentService) ProcessPayment(ctx context.Context, req primary.PaymentRequest) (*domain.Payment, error) {
	// Validate request
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// Verify wallet exists
	wallet, err := s.walletRepo.FindByID(ctx, req.WalletID)
	if err != nil {
		return nil, fmt.Errorf("failed to find wallet: %w", err)
	}

	if wallet == nil {
		return nil, ErrWalletNotFound
	}

	// Get payment gateway
	gateway, exists := s.gateways[req.PaymentProvider]
	if !exists {
		return nil, fmt.Errorf("payment provider %s not supported", req.PaymentProvider)
	}

	// Create transaction record
	transaction, err := domain.NewTransaction(req.WalletID, domain.TransactionTypeDeposit, req.Amount, req.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	transaction.Status = domain.TransactionStatusPending

	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Create payment record
	payment, err := domain.NewPayment(transaction.ID, req.Amount, req.PaymentProvider, req.Description)
	if err != nil {
		// Mark transaction as failed
		_ = s.transactionRepo.UpdateStatus(ctx, transaction.ID, domain.TransactionStatusFailed)
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	err = s.paymentRepo.Create(ctx, payment)
	if err != nil {
		// Mark transaction as failed
		_ = s.transactionRepo.UpdateStatus(ctx, transaction.ID, domain.TransactionStatusFailed)
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	// Process payment through gateway
	gatewayReq := external.PaymentRequest{
		Amount:         req.Amount,
		Currency:       wallet.CurrencyCode,
		Description:    req.Description,
		ReferenceID:    fmt.Sprintf("PAY-%d", payment.ID),
		CustomerName:   fmt.Sprintf("User-%d", wallet.UserID),
		CustomerEmail:  fmt.Sprintf("user%d@example.com", wallet.UserID), // In real app, get from user
		PaymentMethod:  mapToExternalPaymentMethod(req.PaymentProvider),
		RedirectURL:    req.RedirectURL,
		CallbackURL:    req.CallbackURL,
		ExpiryDuration: 30, // 30 minutes
	}

	gatewayResp, err := gateway.ProcessPayment(ctx, gatewayReq)
	if err != nil {
		// Mark payment and transaction as failed
		payment.Fail()
		_ = s.paymentRepo.Update(ctx, payment)
		_ = s.transactionRepo.UpdateStatus(ctx, transaction.ID, domain.TransactionStatusFailed)
		return nil, fmt.Errorf("payment gateway error: %w", err)
	}

	// Update payment with gateway response
	payment.SetExternalInfo(gatewayResp.ExternalID, gatewayResp.PaymentURL, gatewayResp.Details)
	err = s.paymentRepo.Update(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish payment initiated event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "payment.initiated",
			Payload: map[string]interface{}{
				"payment_id":     payment.ID,
				"transaction_id": transaction.ID,
				"wallet_id":      wallet.ID,
				"amount":         payment.Amount,
				"provider":       string(payment.Provider),
				"payment_url":    payment.PaymentURL,
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "payments", event)
		}()
	}

	return payment, nil
}

// VerifyPayment checks the status of a payment and updates accordingly
func (s *PaymentService) VerifyPayment(ctx context.Context, paymentID int) (*domain.Payment, error) {
	// Get payment
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	// Skip if payment is already completed or cancelled
	if payment.Status == domain.PaymentStatusCompleted || payment.Status == domain.PaymentStatusCancelled {
		return payment, nil
	}

	// Get payment gateway
	gateway, exists := s.gateways[payment.Provider]
	if !exists {
		return nil, fmt.Errorf("payment provider %s not supported", payment.Provider)
	}

	// Check payment status from gateway
	gatewayResp, err := gateway.CheckPaymentStatus(ctx, payment.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to check payment status: %w", err)
	}

	// Map gateway status to domain status
	newStatus := mapToDomainPaymentStatus(gatewayResp.Status)

	// Update payment status if changed
	if newStatus != payment.Status {
		oldStatus := payment.Status

		switch newStatus {
		case domain.PaymentStatusCompleted:
			err = payment.Complete()
		case domain.PaymentStatusFailed:
			err = payment.Fail()
		case domain.PaymentStatusCancelled:
			err = payment.Cancel()
		default:
			// Status remains pending or unknown
			err = nil
		}

		if err != nil {
			return nil, fmt.Errorf("failed to update payment status: %w", err)
		}

		// Update payment details
		if gatewayResp.Details != nil {
			payment.Details = gatewayResp.Details
		}

		// Save payment
		err = s.paymentRepo.Update(ctx, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to save payment: %w", err)
		}

		// Update related transaction
		transaction, err := s.transactionRepo.FindByID(ctx, payment.TransactionID)
		if err == nil && transaction != nil {
			if newStatus == domain.PaymentStatusCompleted {
				// Credit wallet
				wallet, err := s.walletRepo.FindByID(ctx, transaction.WalletID)
				if err == nil && wallet != nil {
					err = wallet.Credit(transaction.Amount)
					if err == nil {
						_ = s.walletRepo.Save(ctx, wallet)
						_ = s.transactionRepo.UpdateStatus(ctx, transaction.ID, domain.TransactionStatusCompleted)
					}
				}
			} else if newStatus == domain.PaymentStatusFailed || newStatus == domain.PaymentStatusCancelled {
				_ = s.transactionRepo.UpdateStatus(ctx, transaction.ID, domain.TransactionStatusFailed)
			}
		}

		// Publish payment status updated event
		if s.eventPublisher != nil {
			event := infrastructure.Event{
				Type: "payment.status_updated",
				Payload: map[string]interface{}{
					"payment_id": payment.ID,
					"old_status": string(oldStatus),
					"new_status": string(newStatus),
				},
			}

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = s.eventPublisher.Publish(ctx, "payments", event)
			}()
		}
	}

	return payment, nil
}

// CancelPayment cancels a pending payment
func (s *PaymentService) CancelPayment(ctx context.Context, paymentID int) error {
	// Get payment
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to find payment: %w", err)
	}

	if payment == nil {
		return ErrPaymentNotFound
	}

	// Check if payment can be cancelled
	if payment.Status != domain.PaymentStatusPending {
		return fmt.Errorf("payment cannot be cancelled, current status: %s", payment.Status)
	}

	// Get payment gateway
	gateway, exists := s.gateways[payment.Provider]
	if !exists {
		return fmt.Errorf("payment provider %s not supported", payment.Provider)
	}

	// Cancel payment in gateway
	err = gateway.CancelPayment(ctx, payment.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to cancel payment in gateway: %w", err)
	}

	// Update payment status
	err = payment.Cancel()
	if err != nil {
		return err
	}

	// Save payment
	err = s.paymentRepo.Update(ctx, payment)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	// Update related transaction
	_ = s.transactionRepo.UpdateStatus(ctx, payment.TransactionID, domain.TransactionStatusFailed)

	// Publish payment cancelled event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "payment.cancelled",
			Payload: map[string]interface{}{
				"payment_id":     payment.ID,
				"transaction_id": payment.TransactionID,
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "payments", event)
		}()
	}

	return nil
}

// GetPaymentByID retrieves a payment by its ID
func (s *PaymentService) GetPaymentID(ctx context.Context, paymentID int) (*domain.Payment, error) {
	// Try to get from cache first
	var payment *domain.Payment

	if s.cache != nil {
		cacheKey := fmt.Sprintf("payment:%d", paymentID)
		err := s.cache.GetObject(ctx, cacheKey, &payment)
		if err == nil && payment != nil {
			return payment, nil
		}
	}

	// Fetch from database
	payment, err := s.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	if payment == nil {
		return nil, ErrPaymentNotFound
	}

	// Cache the payment for future requests
	if s.cache != nil {
		cacheKey := fmt.Sprintf("payment:%d", paymentID)
		_ = s.cache.SetObject(ctx, cacheKey, payment, 5*time.Minute)
	}

	return payment, nil
}

// GetPaymentsByTransactionID retrieves all payments for a transaction
func (s *PaymentService) GetPaymentsByTransactionID(ctx context.Context, transactionID int) ([]*domain.Payment, error) {
	payments, err := s.paymentRepo.FindByTransactionID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}

	return payments, nil
}

// Helper functions

func mapToExternalPaymentMethod(provider domain.PaymentProvider) external.PaymentMethod {
	switch provider {
	case domain.PaymentProviderMidtrans:
		return external.PaymentMethodBankTransfer
	case domain.PaymentProviderStripe:
		return external.PaymentMethodCreditCard
	default:
		return external.PaymentMethodEWallet
	}
}

func mapToDomainPaymentStatus(status external.PaymentStatus) domain.PaymentStatus {
	switch status {
	case external.PaymentStatusCompleted:
		return domain.PaymentStatusCompleted
	case external.PaymentStatusFailed:
		return domain.PaymentStatusFailed
	case external.PaymentStatusCancelled:
		return domain.PaymentStatusCancelled
	case external.PaymentStatusExpired:
		return domain.PaymentStatusFailed
	case external.PaymentStatusRefunded:
		return domain.PaymentStatusCancelled
	default:
		return domain.PaymentStatusPending
	}
}
