package payment

import (
	"context"
	"fmt"
	"ports-and-adapters-architecture/internal/ports/secondary/external"
	"time"
)

// MidtransGateway implements the PaymentGateway interface for Midtrans
type MidtransGateway struct {
	serverKey    string
	clientKey    string
	isProduction bool
	baseURL      string
}

// NewMidtransGateway creates a new Midtrans payment gateway
func NewMidtransGateway(serverKey, clientKey string, isProduction bool) *MidtransGateway {
	baseURL := "https://api.sandbox.midtrans.com"
	if isProduction {
		baseURL = "https://api.midtrans.com"
	}

	return &MidtransGateway{
		serverKey:    serverKey,
		clientKey:    clientKey,
		isProduction: isProduction,
		baseURL:      baseURL,
	}
}

// GetProvider returns the payment gateway provider type
func (g *MidtransGateway) GetProvider() external.PaymentGatewayProvider {
	return external.ProviderMidtrans
}

// GetSupportedPaymentMethods returns supported payment methods
func (g *MidtransGateway) GetSupportedPaymentMethods() []external.PaymentMethod {
	return []external.PaymentMethod{
		external.PaymentMethodCreditCard,
		external.PaymentMethodBankTransfer,
		external.PaymentMethodEWallet,
	}
}

// ProcessPayment processes a payment through Midtrans
func (g *MidtransGateway) ProcessPayment(ctx context.Context, request external.PaymentRequest) (*external.PaymentResponse, error) {
	// TODO: Implement actual Midtrans API integration
	// This is a placeholder implementation

	// Simulate API call delay
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Generate mock response
	transactionID := fmt.Sprintf("MT-%d", time.Now().Unix())

	response := &external.PaymentResponse{
		TransactionID:      transactionID,
		ExternalID:         fmt.Sprintf("midtrans_%s", transactionID),
		Status:             external.PaymentStatusPending,
		PaymentURL:         fmt.Sprintf("%s/snap/v1/transactions/%s/pay", g.baseURL, transactionID),
		ProviderResponseID: transactionID,
		PaymentMethod:      request.PaymentMethod,
		ExpiredAt:          time.Now().Add(24 * time.Hour).Unix(),
		Amount:             request.Amount,
		Currency:           request.Currency,
		Details: map[string]interface{}{
			"order_id":      request.ReferenceID,
			"payment_type":  string(request.PaymentMethod),
			"customer_name": request.CustomerName,
		},
	}

	return response, nil
}

// CheckPaymentStatus checks the status of a payment
func (g *MidtransGateway) CheckPaymentStatus(ctx context.Context, transactionID string) (*external.PaymentResponse, error) {
	// TODO: Implement actual Midtrans status check
	// This is a placeholder implementation

	// Simulate API call delay
	select {
	case <-time.After(50 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Mock response - randomly return different statuses
	status := external.PaymentStatusPending
	if time.Now().Unix()%3 == 0 {
		status = external.PaymentStatusCompleted
	}

	response := &external.PaymentResponse{
		TransactionID:      transactionID,
		ExternalID:         fmt.Sprintf("midtrans_%s", transactionID),
		Status:             status,
		ProviderResponseID: transactionID,
		Details: map[string]interface{}{
			"transaction_status": string(status),
			"checked_at":         time.Now().Format(time.RFC3339),
		},
	}

	if status == external.PaymentStatusCompleted {
		response.PaidAt = time.Now().Unix()
	}

	return response, nil
}

// CancelPayment cancels a payment
func (g *MidtransGateway) CancelPayment(ctx context.Context, transactionID string) error {
	// TODO: Implement actual Midtrans cancellation
	// This is a placeholder implementation

	// Simulate API call delay
	select {
	case <-time.After(50 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// RefundRepayment refunds a payment partially or fully
func (g *MidtransGateway) RefundRepayment(ctx context.Context, request external.RefundRequest) (*external.RefundResponse, error) {
	// TODO: Implement actual Midtrans refund
	// This is a placeholder implementation

	// Simulate API call delay
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	refundID := fmt.Sprintf("RF-%s-%d", request.TransactionID, time.Now().Unix())

	response := &external.RefundResponse{
		RefundID:      refundID,
		TransactionID: request.TransactionID,
		Status:        "PENDING",
		Amount:        request.Amount,
		ProcessAt:     time.Now().Unix(),
		Details: map[string]interface{}{
			"reason":       request.Reason,
			"reference_id": request.ReferenceID,
		},
	}

	return response, nil
}

// ValidateCallback validates and processes a callback from Midtrans
func (g *MidtransGateway) ValidateCallback(ctx context.Context, requestBody []byte, headers map[string]string) (*external.PaymentResponse, error) {
	// TODO: Implement actual Midtrans signature validation and callback parsing
	// This is a placeholder implementation

	// Mock response
	response := &external.PaymentResponse{
		TransactionID: "MT-123456",
		ExternalID:    "midtrans_MT-123456",
		Status:        external.PaymentStatusCompleted,
		PaidAt:        time.Now().Unix(),
		Details: map[string]interface{}{
			"callback_received": true,
			"signature_valid":   true,
		},
	}

	return response, nil
}
