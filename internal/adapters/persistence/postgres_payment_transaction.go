package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"time"

	_ "github.com/lib/pq"
)

// PostgresPaymentRepository implements the PaymentRepository interface for PostgreSQL
type PostgresPaymentRepository struct {
	db *sql.DB
}

// NewPostgresPaymentRepository creates a new PostgreSQL payment repository
func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{
		db: db,
	}
}

// FindByID retrieves a payment by its ID
func (r *PostgresPaymentRepository) FindByID(ctx context.Context, id int) (*domain.Payment, error) {
	query := `
		SELECT id, transaction_id, amount, provider, status, external_id, payment_url, 
		       description, details, created_at, updated_at, completed_at
		FROM payments
		WHERE id = $1
	`

	var payment domain.Payment
	var providerStr, statusStr string
	var externalID, paymentURL, description sql.NullString
	var detailsJSON []byte
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&payment.ID,
		&payment.TransactionID,
		&payment.Amount,
		&providerStr,
		&statusStr,
		&externalID,
		&paymentURL,
		&description,
		&detailsJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query payment by ID: %w", err)
	}

	payment.Provider = domain.PaymentProvider(providerStr)
	payment.Status = domain.PaymentStatus(statusStr)

	if externalID.Valid {
		payment.ExternalID = externalID.String
	}

	if paymentURL.Valid {
		payment.PaymentURL = paymentURL.String
	}

	if description.Valid {
		payment.Description = description.String
	}

	if completedAt.Valid {
		payment.CompletedAt = &completedAt.Time
	}

	if len(detailsJSON) > 0 {
		payment.Details = make(map[string]interface{})
		if err := json.Unmarshal(detailsJSON, &payment.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
	}

	return &payment, nil
}

// FindByTransactionID retrieves all payments for a transaction
func (r *PostgresPaymentRepository) FindByTransactionID(ctx context.Context, transactionID int) ([]*domain.Payment, error) {
	query := `
		SELECT id, transaction_id, amount, provider, status, external_id, payment_url, 
		       description, details, created_at, updated_at, completed_at
		FROM payments
		WHERE transaction_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments by transaction ID: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment

	for rows.Next() {
		var payment domain.Payment
		var providerStr, statusStr string
		var externalID, paymentURL, description sql.NullString
		var detailsJSON []byte
		var completedAt sql.NullTime

		err := rows.Scan(
			&payment.ID,
			&payment.TransactionID,
			&payment.Amount,
			&providerStr,
			&statusStr,
			&externalID,
			&paymentURL,
			&description,
			&detailsJSON,
			&payment.CreatedAt,
			&payment.UpdatedAt,
			&completedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan payment row: %w", err)
		}

		payment.Provider = domain.PaymentProvider(providerStr)
		payment.Status = domain.PaymentStatus(statusStr)

		if externalID.Valid {
			payment.ExternalID = externalID.String
		}

		if paymentURL.Valid {
			payment.PaymentURL = paymentURL.String
		}

		if description.Valid {
			payment.Description = description.String
		}

		if completedAt.Valid {
			payment.CompletedAt = &completedAt.Time
		}

		if len(detailsJSON) > 0 {
			payment.Details = make(map[string]interface{})
			if err := json.Unmarshal(detailsJSON, &payment.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal payment details: %w", err)
			}
		}

		payments = append(payments, &payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment rows: %w", err)
	}

	return payments, nil
}

// FindByExternalID retrieves a payment by external ID
func (r *PostgresPaymentRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.Payment, error) {
	query := `
		SELECT id, transaction_id, amount, provider, status, external_id, payment_url, 
		       description, details, created_at, updated_at, completed_at
		FROM payments
		WHERE external_id = $1
	`

	var payment domain.Payment
	var providerStr, statusStr string
	var extID, paymentURL, description sql.NullString
	var detailsJSON []byte
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, externalID).Scan(
		&payment.ID,
		&payment.TransactionID,
		&payment.Amount,
		&providerStr,
		&statusStr,
		&extID,
		&paymentURL,
		&description,
		&detailsJSON,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query payment by external ID: %w", err)
	}

	payment.Provider = domain.PaymentProvider(providerStr)
	payment.Status = domain.PaymentStatus(statusStr)

	if extID.Valid {
		payment.ExternalID = extID.String
	}

	if paymentURL.Valid {
		payment.PaymentURL = paymentURL.String
	}

	if description.Valid {
		payment.Description = description.String
	}

	if completedAt.Valid {
		payment.CompletedAt = &completedAt.Time
	}

	if len(detailsJSON) > 0 {
		payment.Details = make(map[string]interface{})
		if err := json.Unmarshal(detailsJSON, &payment.Details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
	}

	return &payment, nil
}

// FindPendingPayments retrieves all pending payments with optional age limit in minutes
func (r *PostgresPaymentRepository) FindPendingPayments(ctx context.Context, olderThanMinutes int) ([]*domain.Payment, error) {
	query := `
		SELECT id, transaction_id, amount, provider, status, external_id, payment_url, 
		       description, details, created_at, updated_at, completed_at
		FROM payments
		WHERE status = $1
	`

	args := []interface{}{string(domain.PaymentStatusPending)}

	if olderThanMinutes > 0 {
		query += " AND created_at < $2"
		cutoffTime := time.Now().Add(-time.Duration(olderThanMinutes) * time.Minute)
		args = append(args, cutoffTime)
	}

	query += " ORDER BY created_at"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment

	for rows.Next() {
		var payment domain.Payment
		var providerStr, statusStr string
		var externalID, paymentURL, description sql.NullString
		var detailsJSON []byte
		var completedAt sql.NullTime

		err := rows.Scan(
			&payment.ID,
			&payment.TransactionID,
			&payment.Amount,
			&providerStr,
			&statusStr,
			&externalID,
			&paymentURL,
			&description,
			&detailsJSON,
			&payment.CreatedAt,
			&payment.UpdatedAt,
			&completedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan payment row: %w", err)
		}

		payment.Provider = domain.PaymentProvider(providerStr)
		payment.Status = domain.PaymentStatus(statusStr)

		if externalID.Valid {
			payment.ExternalID = externalID.String
		}

		if paymentURL.Valid {
			payment.PaymentURL = paymentURL.String
		}

		if description.Valid {
			payment.Description = description.String
		}

		if completedAt.Valid {
			payment.CompletedAt = &completedAt.Time
		}

		if len(detailsJSON) > 0 {
			payment.Details = make(map[string]interface{})
			if err := json.Unmarshal(detailsJSON, &payment.Details); err != nil {
				return nil, fmt.Errorf("failed to unmarshal payment details: %w", err)
			}
		}

		payments = append(payments, &payment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment rows: %w", err)
	}

	return payments, nil
}

// Create saves a new payment
func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (transaction_id, amount, provider, status, external_id, payment_url, 
		                     description, details, created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var detailsJSON []byte
	var err error

	if payment.Details != nil && len(payment.Details) > 0 {
		detailsJSON, err = json.Marshal(payment.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal payment details: %w", err)
		}
	}

	var completedAt *time.Time
	if payment.Status == domain.PaymentStatusCompleted {
		now := time.Now()
		completedAt = &now
		payment.CompletedAt = completedAt
	}

	err = r.db.QueryRowContext(
		ctx,
		query,
		payment.TransactionID,
		payment.Amount,
		string(payment.Provider),
		string(payment.Status),
		sql.NullString{String: payment.ExternalID, Valid: payment.ExternalID != ""},
		sql.NullString{String: payment.PaymentURL, Valid: payment.PaymentURL != ""},
		sql.NullString{String: payment.Description, Valid: payment.Description != ""},
		detailsJSON,
		payment.CreatedAt,
		payment.UpdatedAt,
		sql.NullTime{Time: safeDerefTime(completedAt), Valid: completedAt != nil},
	).Scan(&payment.ID)

	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	return nil
}

// Update updates an existing payment
func (r *PostgresPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	query := `
		UPDATE payments
		SET status = $1, external_id = $2, payment_url = $3, description = $4, 
		    details = $5, updated_at = $6, completed_at = $7
		WHERE id = $8
	`

	var detailsJSON []byte
	var err error

	if payment.Details != nil && len(payment.Details) > 0 {
		detailsJSON, err = json.Marshal(payment.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal payment details: %w", err)
		}
	}

	payment.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		string(payment.Status),
		sql.NullString{String: payment.ExternalID, Valid: payment.ExternalID != ""},
		sql.NullString{String: payment.PaymentURL, Valid: payment.PaymentURL != ""},
		sql.NullString{String: payment.Description, Valid: payment.Description != ""},
		detailsJSON,
		payment.UpdatedAt,
		sql.NullTime{Time: safeDerefTime(payment.CompletedAt), Valid: payment.CompletedAt != nil},
		payment.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found: %d", payment.ID)
	}

	return nil
}
