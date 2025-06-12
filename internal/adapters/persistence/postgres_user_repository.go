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

// PostgresUserRepository implements the UserRepository interface for PostgreSQL
type PostgresUserRepository struct {
	db *sql.DB
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db: db,
	}
}

// FindByID retrieves a user by ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	query := `
		SELECT id, fullname, email, phone, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Fullname,
		&user.Email,
		&user.Phone,
		&statusStr,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query user by ID: %w", err)
	}

	user.Status = domain.UserStatus(statusStr)

	return &user, nil
}

// FindByEmail retrieves a user by email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, fullname, email, phone, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Fullname,
		&user.Email,
		&user.Phone,
		&statusStr,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	user.Status = domain.UserStatus(statusStr)

	return &user, nil
}

// FindByPhone retrieves a user by phone
func (r *PostgresUserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, fullname, email, phone, status, created_at, updated_at
		FROM users
		WHERE phone = $1
	`

	var user domain.User
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&user.ID,
		&user.Fullname,
		&user.Email,
		&user.Phone,
		&statusStr,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query user by phone: %w", err)
	}

	user.Status = domain.UserStatus(statusStr)

	return &user, nil
}

// Save creates or updates a user
func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	if user.ID == 0 {
		// Create new user
		query := `
			INSERT INTO users (fullname, email, phone, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`

		err := r.db.QueryRowContext(
			ctx,
			query,
			user.Fullname,
			user.Email,
			user.Phone,
			string(user.Status),
			user.CreatedAt,
			user.UpdatedAt,
		).Scan(&user.ID)

		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		return nil
	}

	// Update existing user
	query := `
		UPDATE users
		SET fullname = $1, email = $2, phone = $3, status = $4, updated_at = $5
		WHERE id = $6
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Fullname,
		user.Email,
		user.Phone,
		string(user.Status),
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %d", user.ID)
	}

	return nil
}

// Delete removes a user
func (r *PostgresUserRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %d", id)
	}

	return nil
}
