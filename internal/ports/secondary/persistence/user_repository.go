package ports

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// UserRepository defines the port for user data operation
type UserRepository interface {
	// FindByID retrieves a user by ID
	FindByID(ctx context.Context, id int) (*domain.User, error)

	// FindByEmail retrieves a user by email
	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	// FindByPhone retrieves a user by phone
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)

	// Save creates or updates a user
	Save(ctx context.Context, user *domain.User) error

	// Delete removes a User
	Delete(ctx context.Context, id int) error
}
