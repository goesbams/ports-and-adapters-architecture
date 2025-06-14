package rest

import (
	"ports-and-adapters-architecture/api/rest/handlers"
	"ports-and-adapters-architecture/internal/ports/primary"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CustomValidator implements echo validator interface
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

// SetupRoutes sets up all HTTP routes
func SetupRoutes(
	e *echo.Echo,
	walletService primary.WalletService,
	paymentService primary.PaymentService,
) {
	// Setup validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})

	// API v1 group
	v1 := e.Group("/api/v1")

	// Initialize handlers
	walletHandler := handlers.NewWalletHandler(walletService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)

	// Wallet routes
	wallets := v1.Group("/wallets")
	wallets.POST("", walletHandler.CreateWallet)
	wallets.GET("/:id", walletHandler.GetWallet)
	wallets.POST("/:id/deposit", walletHandler.Deposit)
	wallets.POST("/:id/withdraw", walletHandler.Withdraw)
	wallets.POST("/:id/transfer", walletHandler.Transfer)
	wallets.GET("/:id/transactions", walletHandler.GetTransactionHistory)
	wallets.GET("/:id/balance", walletHandler.GetBalance)

	// User wallet routes
	v1.GET("/users/:user_id/wallets", walletHandler.GetWalletsByUserID)

	// Payment routes
	payments := v1.Group("/payments")
	payments.POST("/process", paymentHandler.ProcessPayment)
	payments.GET("/:id", paymentHandler.GetPayment)
	payments.POST("/:id/verify", paymentHandler.VerifyPayment)
	payments.POST("/:id/cancel", paymentHandler.CancelPayment)
	payments.GET("/transaction/:transaction_id", paymentHandler.GetPaymentsByTransactionID)
	payments.POST("/callback/:provider", paymentHandler.PaymentCallback)
}
