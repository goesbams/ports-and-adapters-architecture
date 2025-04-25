package primary

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
)

// UserService defines the contract for user application service
type UserService interface {

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id int) (*domain.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetUserByPhone retrieves a user by phone
	GetUserByPhone(ctx context.Context, phone string) (*domain.User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, fullname, email, phone string) (*domain.User, error)

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, id int, fullname, email, phone string) (*domain.User, error)

	// DeactiveUser deactives a user(sets status to inactive)
	DeactiveUser(ctx context.Context, id int) error

	// ActivateUser activates a user
	ActivateUser(ctx context.Context, id int) error
}
