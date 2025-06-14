package handlers

import (
	"net/http"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/ports/primary"
	"strconv"

	"github.com/labstack/echo/v4"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentService primary.PaymentService
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(paymentService primary.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// ProcessPaymentRequest represents the request to process a payment
type ProcessPaymentRequest struct {
	WalletID        int                    `json:"wallet_id" validate:"required,min=1"`
	Amount          int                    `json:"amount" validate:"required,min=1"`
	Description     string                 `json:"description"`
	PaymentProvider domain.PaymentProvider `json:"payment_provider" validate:"required"`
	RedirectURL     string                 `json:"redirect_url"`
	CallbackURL     string                 `json:"callback_url"`
}

// ProcessPayment handles POST /api/v1/payments/process
func (h *PaymentHandler) ProcessPayment(c echo.Context) error {
	var req ProcessPaymentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	paymentReq := primary.PaymentRequest{
		WalletID:        req.WalletID,
		Amount:          req.Amount,
		Description:     req.Description,
		PaymentProvider: req.PaymentProvider,
		RedirectURL:     req.RedirectURL,
		CallbackURL:     req.CallbackURL,
	}

	payment, err := h.paymentService.ProcessPayment(c.Request().Context(), paymentReq)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   payment,
	})
}

// GetPayment handles GET /api/v1/payments/:id
func (h *PaymentHandler) GetPayment(c echo.Context) error {
	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment ID")
	}

	payment, err := h.paymentService.GetPaymentID(c.Request().Context(), paymentID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   payment,
	})
}

// VerifyPayment handles POST /api/v1/payments/:id/verify
func (h *PaymentHandler) VerifyPayment(c echo.Context) error {
	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment ID")
	}

	payment, err := h.paymentService.VerifyPayment(c.Request().Context(), paymentID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   payment,
	})
}

// CancelPayment handles POST /api/v1/payments/:id/cancel
func (h *PaymentHandler) CancelPayment(c echo.Context) error {
	paymentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment ID")
	}

	err = h.paymentService.CancelPayment(c.Request().Context(), paymentID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Payment cancelled successfully",
	})
}

// GetPaymentsByTransactionID handles GET /api/v1/payments/transaction/:transaction_id
func (h *PaymentHandler) GetPaymentsByTransactionID(c echo.Context) error {
	transactionID, err := strconv.Atoi(c.Param("transaction_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid transaction ID")
	}

	payments, err := h.paymentService.GetPaymentsByTransactionID(c.Request().Context(), transactionID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   payments,
	})
}

// PaymentCallback handles POST /api/v1/payments/callback/:provider
func (h *PaymentHandler) PaymentCallback(c echo.Context) error {
	provider := c.Param("provider")

	// Get request body
	body, err := c.Request().GetBody()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}

	// Get headers
	headers := make(map[string]string)
	for key, values := range c.Request().Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// TODO: Process callback based on provider
	// This would involve looking up the payment by external ID
	// and verifying the callback signature

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Callback received",
	})
}
