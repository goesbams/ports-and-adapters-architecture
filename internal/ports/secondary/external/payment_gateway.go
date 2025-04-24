package external

import "context"

// PaymentGatewayProvider represents different payment gateway providers
type PaymentGatewayProvider string

const (
	ProviderMidtrans PaymentGatewayProvider = "MIDTRANS"
	ProviderDoku     PaymentGatewayProvider = "DOKU"
	ProviderStripe   PaymentGatewayProvider = "STRIPE"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

// PaymentRequest represents a request to a payment gateway
type PaymentRequest struct {
	Amount        int    `json:"amount"`
	Currency      string `json:"currency"`
	Description   string `json:"description"`
	ReferenceID   string `json:"reference_id"`
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
}

// PaymentResponse represents a response from a payment gateway
type PaymentResponse struct {
	TransactionID      string                 `json:"transaction_id"`
	ExternalID         string                 `json:"external_id"`
	Status             PaymentStatus          `json:"status"`
	PaymentURL         string                 `json:"payment_url,omitempty"`
	ProviderResponseID string                 `json:"provider_response_id"`
	Details            map[string]interface{} `json:"details,omitempty"`
}

// PaymentGateway defines the port for payment gateway operations
type PaymentGateway interface {
	// ProcessPayment process a payment throught the payment gateway
	ProcessPayment(ctx context.Context, request PaymentRequest) (*PaymentResponse, error)

	// CheckPaymentStatus checks the status of a payment
	CheckPaymentStatus(ctx context.Context, transactionID string) (*PaymentResponse, error)

	// CancelPayment cancels a payment
	CancelPayment(ctx context.Context, transactionID string) error
}
