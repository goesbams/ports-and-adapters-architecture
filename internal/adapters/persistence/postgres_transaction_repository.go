package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"time"
)

// PostgresTransactionRepository implements the TransactionRepository interface for PostgreSQL
type PostgresTransactionRepository struct {
	db *sql.DB
}

// NewPostgresTransactionRepository creates a new PostgreSQL transaction repository
func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{
		db: db,
	}
}

// FindByID retrieves a transaction by its ID
func (r *PostgresTransactionRepository) FindByID(ctx context.Context, id int) (*domain.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, status, reference, description, to_wallet_id, 
		       created_at, updated_at, completed_at
		FROM transactions
		WHERE id = $1
	`

	var transaction domain.Transaction
	var typeStr, statusStr string
	var reference, description sql.NullString
	var toWalletID sql.NullInt64
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID,
		&transaction.WalletID,
		&typeStr,
		&transaction.Amount,
		&statusStr,
		&reference,
		&description,
		&toWalletID,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query transaction by ID: %w", err)
	}

	transaction.Type = domain.TransactionType(typeStr)
	transaction.Status = domain.TransactionStatus(statusStr)

	if reference.Valid {
		transaction.Reference = reference.String
	}

	if description.Valid {
		transaction.Description = description.String
	}

	if toWalletID.Valid {
		walletID := int(toWalletID.Int64)
		transaction.ToWalletID = &walletID
	}

	if completedAt.Valid {
		transaction.CompletedAt = &completedAt.Time
	}

	return &transaction, nil
}

// FindByWalletID retrieves all transactions for a wallet
func (r *PostgresTransactionRepository) FindByWalletID(ctx context.Context, walletID int, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, status, reference, description, to_wallet_id, 
		       created_at, updated_at, completed_at
		FROM transactions
		WHERE wallet_id = $1 OR to_wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by wallet ID: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction

	for rows.Next() {
		var transaction domain.Transaction
		var typeStr, statusStr string
		var reference, description sql.NullString
		var toWalletID sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&transaction.ID,
			&transaction.WalletID,
			&typeStr,
			&transaction.Amount,
			&statusStr,
			&reference,
			&description,
			&toWalletID,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&completedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		transaction.Type = domain.TransactionType(typeStr)
		transaction.Status = domain.TransactionStatus(statusStr)

		if reference.Valid {
			transaction.Reference = reference.String
		}

		if description.Valid {
			transaction.Description = description.String
		}

		if toWalletID.Valid {
			id := int(toWalletID.Int64)
			transaction.ToWalletID = &id
		}

		if completedAt.Valid {
			transaction.CompletedAt = &completedAt.Time
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, nil
}

// FindByStatus retrieves transactions by status with optional pagination
func (r *PostgresTransactionRepository) FindByStatus(ctx context.Context, status domain.TransactionStatus, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, status, reference, description, to_wallet_id, 
		       created_at, updated_at, completed_at
		FROM transactions
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions by status: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction

	for rows.Next() {
		var transaction domain.Transaction
		var typeStr, statusStr string
		var reference, description sql.NullString
		var toWalletID sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&transaction.ID,
			&transaction.WalletID,
			&typeStr,
			&transaction.Amount,
			&statusStr,
			&reference,
			&description,
			&toWalletID,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&completedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		transaction.Type = domain.TransactionType(typeStr)
		transaction.Status = domain.TransactionStatus(statusStr)

		if reference.Valid {
			transaction.Reference = reference.String
		}

		if description.Valid {
			transaction.Description = description.String
		}

		if toWalletID.Valid {
			id := int(toWalletID.Int64)
			transaction.ToWalletID = &id
		}

		if completedAt.Valid {
			transaction.CompletedAt = &completedAt.Time
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, nil
}

// FindPendingTransactions retrieves pending transactions older than a specified time
func (r *PostgresTransactionRepository) FindPendingTransactions(ctx context.Context, olderThan time.Time) ([]*domain.Transaction, error) {
	query := `
		SELECT id, wallet_id, type, amount, status, reference, description, to_wallet_id, 
		       created_at, updated_at, completed_at
		FROM transactions
		WHERE status = $1 AND created_at < $2
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, string(domain.TransactionStatusPending), olderThan)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*domain.Transaction

	for rows.Next() {
		var transaction domain.Transaction
		var typeStr, statusStr string
		var reference, description sql.NullString
		var toWalletID sql.NullInt64
		var completedAt sql.NullTime

		err := rows.Scan(
			&transaction.ID,
			&transaction.WalletID,
			&typeStr,
			&transaction.Amount,
			&statusStr,
			&reference,
			&description,
			&toWalletID,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&completedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		transaction.Type = domain.TransactionType(typeStr)
		transaction.Status = domain.TransactionStatus(statusStr)

		if reference.Valid {
			transaction.Reference = reference.String
		}

		if description.Valid {
			transaction.Description = description.String
		}

		if toWalletID.Valid {
			id := int(toWalletID.Int64)
			transaction.ToWalletID = &id
		}

		if completedAt.Valid {
			transaction.CompletedAt = &completedAt.Time
		}

		transactions = append(transactions, &transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, nil
}

// CountByWalletID counts all transactions for a wallet
func (r *PostgresTransactionRepository) CountByWalletID(ctx context.Context, walletID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM transactions
		WHERE wallet_id = $1 OR to_wallet_id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, walletID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions by wallet ID: %w", err)
	}

	return count, nil
}

// Create saves a new transaction
func (r *PostgresTransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (wallet_id, type, amount, status, reference, description, to_wallet_id, 
		                         created_at, updated_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var completedAt *time.Time
	if transaction.Status == domain.TransactionStatusCompleted {
		now := time.Now()
		completedAt = &now
		transaction.CompletedAt = completedAt
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		transaction.WalletID,
		string(transaction.Type),
		transaction.Amount,
		string(transaction.Status),
		sql.NullString{String: transaction.Reference, Valid: transaction.Reference != ""},
		sql.NullString{String: transaction.Description, Valid: transaction.Description != ""},
		sql.NullInt64{Int64: int64(safeDeref(transaction.ToWalletID)), Valid: transaction.ToWalletID != nil},
		transaction.CreatedAt,
		transaction.UpdatedAt,
		sql.NullTime{Time: safeDerefTime(completedAt), Valid: completedAt != nil},
	).Scan(&transaction.ID)

	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	return nil
}

// Update updates an existing transaction
func (r *PostgresTransactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		UPDATE transactions
		SET status = $1, reference = $2, description = $3, updated_at = $4, completed_at = $5
		WHERE id = $6
	`

	transaction.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		string(transaction.Status),
		sql.NullString{String: transaction.Reference, Valid: transaction.Reference != ""},
		sql.NullString{String: transaction.Description, Valid: transaction.Description != ""},
		transaction.UpdatedAt,
		sql.NullTime{Time: safeDerefTime(transaction.CompletedAt), Valid: transaction.CompletedAt != nil},
		transaction.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found: %d", transaction.ID)
	}

	return nil
}

// UpdateStatus updates only the status of a transaction
func (r *PostgresTransactionRepository) UpdateStatus(ctx context.Context, id int, status domain.TransactionStatus) error {
	query := `
		UPDATE transactions
		SET status = $1, updated_at = $2, 
		    completed_at = CASE WHEN $1 = 'COMPLETED' AND completed_at IS NULL THEN $3 ELSE completed_at END
		WHERE id = $4
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		string(status),
		now,
		sql.NullTime{Time: now, Valid: status == domain.TransactionStatusCompleted},
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found: %d", id)
	}

	return nil
}

// Helper functions for handling nil pointers
func safeDeref(ptr *int) int {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func safeDerefTime(ptr *time.Time) time.Time {
	if ptr == nil {
		return time.Time{}
	}
	return *ptr
}
