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
	PaymentStatusCancelled PaymentStatus = "CANCELLED"
	PaymentStatusExpired   PaymentStatus = "EXPIRED"
	PaymentStatusRefunded  PaymentStatus = "REFUNDED"
)

// PaymentMethod represents the payment method used
type PaymentMethod string

const (
	PaymentMethodCreditCard   PaymentMethod = "CREDIT_CARD"
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	PaymentMethodEWallet      PaymentMethod = "E_WALLET"
	PaymentMethodDirectDebit  PaymentMethod = "DIRECT_DEBIT"
)

// PaymentRequest represents a request to a payment gateway
type PaymentRequest struct {
	Amount         int               `json:"amount"`
	Currency       string            `json:"currency"`
	Description    string            `json:"description"`
	ReferenceID    string            `json:"reference_id"`
	CustomerName   string            `json:"customer_name"`
	CustomerEmail  string            `json:"customer_email"`
	CustomerPhone  string            `json:"customer_phone,omitempty"`
	PaymentMethod  PaymentMethod     `json:"payment_method,omitempty"`
	RedirectURL    string            `json:"redirect_url,omitempty"`
	CallbackURL    string            `json:"callback_url,omitempty"`
	ExpiryDuration int               `json:"expiry_duration_minutes,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PaymentResponse represents a response from a payment gateway
type PaymentResponse struct {
	TransactionID      string                 `json:"transaction_id"`
	ExternalID         string                 `json:"external_id"`
	Status             PaymentStatus          `json:"status"`
	PaymentURL         string                 `json:"payment_url,omitempty"`
	ProviderResponseID string                 `json:"provider_response_id"`
	PaymentMethod      PaymentMethod          `json:"payment_method,omitempty"`
	ExpiredAt          int64                  `json:"expired_at,omitempty"`
	PaidAt             int64                  `json:"paid_at,omitempty"`
	Amount             int                    `json:"amount"`
	Currency           string                 `json:"currency"`
	Details            map[string]interface{} `json:"details,omitempty"`
}

// RefundRequest represents a request to refund a payment
type RefundRequest struct {
	TransactionID string `json:"transaction_id"`
	Amount        int    `json:"amount"` // if 0, refund full amount
	Reason        string `json:"reason"`
	ReferenceID   string `json:"reference_id"`
}

// RefundResponse represents a response from a refund request
type RefundResponse struct {
	RefundID      string                 `json:"refund_id"`
	TransactionID string                 `json:"transaction_id"`
	Status        string                 `json:"status"`
	Amount        int                    `json:"amount"`
	ProcessAt     int64                  `json:"processed_at,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
}

// PaymentGateway defines the port for payment gateway operations
type PaymentGateway interface {
	// GetProvider returns the payment gateway provider type
	GetProvider() PaymentGatewayProvider

	// GetSupportedPaymentMethods returns supported payment methods
	GetSupportedPaymentMethods() []PaymentMethod

	// ProcessPayment process a payment throught the payment gateway
	ProcessPayment(ctx context.Context, request PaymentRequest) (*PaymentResponse, error)

	// CheckPaymentStatus checks the status of a payment
	CheckPaymentStatus(ctx context.Context, transactionID string) (*PaymentResponse, error)

	// CancelPayment cancels a payment
	CancelPayment(ctx context.Context, transactionID string) error

	// RefundPayment refunds a payment partially or fully
	RefundRepayment(ctx context.Context, request RefundRequest) (*RefundResponse, error)

	// ValidateCallback validates and process a callback from the payment gateway
	ValidateCallback(ctx context.Context, requestBody []byte, headers map[string]string) (*PaymentResponse, error)
}
