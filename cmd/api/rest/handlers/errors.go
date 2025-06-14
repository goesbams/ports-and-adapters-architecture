package handlers

import (
	"errors"
	"net/http"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/usecase"

	"github.com/labstack/echo/v4"
)

// handleServiceError converts service errors to HTTP errors
func handleServiceError(err error) error {
	if err == nil {
		return nil
	}

	// Domain errors
	if errors.Is(err, domain.ErrInsufficientBalance) {
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient balance")
	}
	if errors.Is(err, domain.ErrInvalidAmount) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid amount")
	}
	if errors.Is(err, domain.ErrWalletNotActive) {
		return echo.NewHTTPError(http.StatusBadRequest, "Wallet is not active")
	}

	// Use case errors
	if errors.Is(err, usecase.ErrWalletNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "Wallet not found")
	}
	if errors.Is(err, usecase.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	if errors.Is(err, usecase.ErrInsufficientBalance) {
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient balance")
	}
	if errors.Is(err, usecase.ErrInvalidAmount) {
		return echo.NewHTTPError(http.StatusBadRequest, "Amount must be greater than zero")
	}
	if errors.Is(err, usecase.ErrTransactionFailed) {
		return echo.NewHTTPError(http.StatusInternalServerError, "Transaction failed")
	}
	if errors.Is(err, usecase.ErrTransferFailed) {
		return echo.NewHTTPError(http.StatusInternalServerError, "Transfer failed")
	}
	if errors.Is(err, usecase.ErrWalletAlreadyExists) {
		return echo.NewHTTPError(http.StatusConflict, "Wallet already exists for this user and currency")
	}
	if errors.Is(err, usecase.ErrEmailAlreadyExists) {
		return echo.NewHTTPError(http.StatusConflict, "Email already exists")
	}
	if errors.Is(err, usecase.ErrPhoneAlreadyExists) {
		return echo.NewHTTPError(http.StatusConflict, "Phone already exists")
	}
	if errors.Is(err, usecase.ErrTransactionNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "Transaction not found")
	}
	if errors.Is(err, usecase.ErrPaymentNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
	}
	if errors.Is(err, usecase.ErrPaymentGatewayFailed) {
		return echo.NewHTTPError(http.StatusBadGateway, "Payment gateway failed")
	}
	if errors.Is(err, usecase.ErrInvalidPaymentStatus) {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment status")
	}

	// Default error
	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
