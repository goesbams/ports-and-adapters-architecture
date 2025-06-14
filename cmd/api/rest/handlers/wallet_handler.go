package handlers

import (
	"net/http"
	"ports-and-adapters-architecture/internal/ports/primary"
	"strconv"

	"github.com/labstack/echo/v4"
)

// WalletHandler handles wallet-related HTTP requests
type WalletHandler struct {
	walletService primary.WalletService
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(walletService primary.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// CreateWalletRequest represents the request to create a wallet
type CreateWalletRequest struct {
	UserID       int    `json:"user_id" validate:"required,min=1"`
	CurrencyCode string `json:"currency_code" validate:"required,len=3"`
	Description  string `json:"description"`
}

// DepositRequest represents the request to deposit funds
type DepositRequest struct {
	Amount      int    `json:"amount" validate:"required,min=1"`
	Description string `json:"description"`
}

// WithdrawRequest represents the request to withdraw funds
type WithdrawRequest struct {
	Amount      int    `json:"amount" validate:"required,min=1"`
	Description string `json:"description"`
}

// TransferRequest represents the request to transfer funds
type TransferRequest struct {
	ToWalletID  int    `json:"to_wallet_id" validate:"required,min=1"`
	Amount      int    `json:"amount" validate:"required,min=1"`
	Description string `json:"description"`
}

// CreateWallet handles POST /api/v1/wallets
func (h *WalletHandler) CreateWallet(c echo.Context) error {
	var req CreateWalletRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	wallet, err := h.walletService.CreateWallet(c.Request().Context(), req.UserID, req.CurrencyCode, req.Description)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   wallet,
	})
}

// GetWallet handles GET /api/v1/wallets/:id
func (h *WalletHandler) GetWallet(c echo.Context) error {
	walletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid wallet ID")
	}

	wallet, err := h.walletService.GetWallet(c.Request().Context(), walletID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   wallet,
	})
}

// GetWalletsByUserID handles GET /api/v1/users/:user_id/wallets
func (h *WalletHandler) GetWalletsByUserID(c echo.Context) error {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	wallets, err := h.walletService.GetWalletsByUserID(c.Request().Context(), userID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   wallets,
	})
}

// Deposit handles POST /api/v1/wallets/:id/deposit
func (h *WalletHandler) Deposit(c echo.Context) error {
	walletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid wallet ID")
	}

	var req DepositRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	transaction, err := h.walletService.Deposit(c.Request().Context(), walletID, req.Amount, req.Description)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   transaction,
	})
}

// Withdraw handles POST /api/v1/wallets/:id/withdraw
func (h *WalletHandler) Withdraw(c echo.Context) error {
	walletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid wallet ID")
	}

	var req WithdrawRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	transaction, err := h.walletService.Withdraw(c.Request().Context(), walletID, req.Amount, req.Description)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   transaction,
	})
}

// Transfer handles POST /api/v1/wallets/:id/transfer
func (h *WalletHandler) Transfer(c echo.Context) error {
	fromWalletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid wallet ID")
	}

	var req TransferRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	transaction, err := h.walletService.Transfer(
		c.Request().Context(),
		fromWalletID,
		req.ToWalletID,
		req.Amount,
		req.Description,
	)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   transaction,
	})
}

// GetTransactionHistory handles GET /api/v1/wallets/:id/transactions
func (h *WalletHandler) GetTransactionHistory(c echo.Context) error {
	walletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid wallet ID")
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	transactions, total, err := h.walletService.GetTransactionHistory(
		c.Request().Context(),
		walletID,
		limit,
		offset,
	)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"transactions": transactions,
			"total":        total,
			"limit":        limit,
			"offset":       offset,
		},
	})
}

// GetBalance handles GET /api/v1/wallets/:id/balance
func (h *WalletHandler) GetBalance(c echo.Context) error {
	walletID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid wallet ID")
	}

	balance, currency, err := h.walletService.GetBalance(c.Request().Context(), walletID)
	if err != nil {
		return handleServiceError(err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"wallet_id": walletID,
			"balance":   balance,
			"currency":  currency,
		},
	})
}
