package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"time"

	_ "github.com/lib/pq"
)

// PostgresWalletRepository implements the WalletRepository interface for PostgreSQL
type PostgresWalletRepository struct {
	db *sql.DB
}

// NewPostgresWalletRepository creates a new PostgreSQL wallet repository
func NewPostgresWalletRepository(db *sql.DB) *PostgresWalletRepository {
	return &PostgresWalletRepository{
		db: db,
	}
}

// FindByID retrieves a wallet by its ID
func (r *PostgresWalletRepository) FindByID(ctx context.Context, id int) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency_code, description, status, created_at, updated_at
		FROM wallets
		WHERE id = $1
	`

	var wallet domain.Wallet
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.CurrencyCode,
		&wallet.Description,
		&statusStr,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query wallet by ID: %w", err)
	}

	wallet.Status = domain.WalletStatus(statusStr)

	return &wallet, nil
}

// FindByUserID retrieves all wallets for a user
func (r *PostgresWalletRepository) FindByUserID(ctx context.Context, userID int) ([]*domain.Wallet, error) {
	query := `
		SELECT id, user_id, balance, currency_code, description, status, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallets by user ID: %w", err)
	}
	defer rows.Close()

	var wallets []*domain.Wallet

	for rows.Next() {
		var wallet domain.Wallet
		var statusStr string

		err := rows.Scan(
			&wallet.ID,
			&wallet.UserID,
			&wallet.Balance,
			&wallet.CurrencyCode,
			&wallet.Description,
			&statusStr,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet row: %w", err)
		}

		wallet.Status = domain.WalletStatus(statusStr)
		wallets = append(wallets, &wallet)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating wallet rows: %w", err)
	}

	return wallets, nil
}

// Save creates or updates a wallet
func (r *PostgresWalletRepository) Save(ctx context.Context, wallet *domain.Wallet) error {
	if wallet.ID == 0 {
		// Create new wallet
		query := `
			INSERT INTO wallets (user_id, balance, currency_code, description, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
		`

		err := r.db.QueryRowContext(
			ctx,
			query,
			wallet.UserID,
			wallet.Balance,
			wallet.CurrencyCode,
			wallet.Description,
			string(wallet.Status),
			wallet.CreatedAt,
			wallet.UpdatedAt,
		).Scan(&wallet.ID)

		if err != nil {
			return fmt.Errorf("failed to insert wallet: %w", err)
		}

		return nil
	}

	// Update existing wallet
	query := `
		UPDATE wallets
		SET user_id = $1, balance = $2, currency_code = $3, description = $4, status = $5, updated_at = $6
		WHERE id = $7
	`

	wallet.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		wallet.UserID,
		wallet.Balance,
		wallet.CurrencyCode,
		wallet.Description,
		string(wallet.Status),
		wallet.UpdatedAt,
		wallet.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found: %d", wallet.ID)
	}

	return nil
}

// UpdateBalance updates only the wallet balance
func (r *PostgresWalletRepository) UpdateBalance(ctx context.Context, walletID int, newBalance int) error {
	query := `
		UPDATE wallets
		SET balance = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, newBalance, now, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found: %d", walletID)
	}

	return nil
}

// UpdateStatus updates only the wallet status
func (r *PostgresWalletRepository) UpdateStatus(ctx context.Context, walletID int, status domain.WalletStatus) error {
	query := `
		UPDATE wallets
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, string(status), now, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found: %d", walletID)
	}

	return nil
}

// Delete removes a wallet
func (r *PostgresWalletRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM wallets
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found: %d", id)
	}

	return nil
}

// NewPostgresConnection creates a new PostgreSQL database connection
func NewPostgresConnection(host, port, user, password, dbName string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
