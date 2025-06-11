package usecase

import (
	"context"
	"errors"
	"fmt"
	"ports-and-adapters-architecture/internal/domain"
	"ports-and-adapters-architecture/internal/ports/secondary/infrastructure"
	"ports-and-adapters-architecture/internal/ports/secondary/persistence"
	"time"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrPhoneAlreadyExists = errors.New("phone already exists")
)

// UserService implements the user application service
type UserService struct {
	userRepo       persistence.UserRepository
	eventPublisher infrastructure.EventPublisher
	cache          infrastructure.Cache
}

// NewUserService creates a new user service
func NewUserService(
	userRepo persistence.UserRepository,
	eventPublisher infrastructure.EventPublisher,
	cache infrastructure.Cache,
) *UserService {
	return &UserService{
		userRepo:       userRepo,
		eventPublisher: eventPublisher,
		cache:          cache,
	}
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id int) (*domain.User, error) {
	// Try to get from cache first
	var user *domain.User

	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		err := s.cache.GetObject(ctx, cacheKey, &user)
		if err == nil && user != nil {
			return user, nil
		}
	}

	// Fetch from database
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	// Cache the user for future requests
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		_ = s.cache.SetObject(ctx, cacheKey, user, 5*time.Minute)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetUserByPhone retrieves a user by phone
func (s *UserService) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	user, err := s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by phone: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, fullname, email, phone string) (*domain.User, error) {
	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}

	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Check if phone already exists
	existingUser, err = s.userRepo.FindByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing phone: %w", err)
	}

	if existingUser != nil {
		return nil, ErrPhoneAlreadyExists
	}

	// Create new user
	user := domain.NewUser(fullname, email, phone)

	// Save user
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Publish user created event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "user.created",
			Payload: map[string]interface{}{
				"user_id":  user.ID,
				"email":    user.Email,
				"fullname": user.Fullname,
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "users", event)
		}()
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserService) UpdateUser(ctx context.Context, id int, fullname, email, phone string) (*domain.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if email changed and if it's already taken
	if email != user.Email {
		existingUser, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}

		if existingUser != nil && existingUser.ID != id {
			return nil, ErrEmailAlreadyExists
		}
	}

	// Check if phone changed and if it's already taken
	if phone != user.Phone {
		existingUser, err := s.userRepo.FindByPhone(ctx, phone)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing phone: %w", err)
		}

		if existingUser != nil && existingUser.ID != id {
			return nil, ErrPhoneAlreadyExists
		}
	}

	// Update user fields
	user.Fullname = fullname
	user.Email = email
	user.Phone = phone
	user.UpdatedAt = time.Now()

	// Save changes
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish user updated event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "user.updated",
			Payload: map[string]interface{}{
				"user_id":  user.ID,
				"email":    user.Email,
				"fullname": user.Fullname,
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "users", event)
		}()
	}

	return user, nil
}

// DeactiveUser deactivates a user
func (s *UserService) DeactiveUser(ctx context.Context, id int) error {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return ErrUserNotFound
	}

	// Set status to inactive
	user.Status = domain.UserStatusInactive
	user.UpdatedAt = time.Now()

	// Save changes
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish user deactivated event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "user.deactivated",
			Payload: map[string]interface{}{
				"user_id": user.ID,
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "users", event)
		}()
	}

	return nil
}

// ActivateUser activates a user
func (s *UserService) ActivateUser(ctx context.Context, id int) error {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user == nil {
		return ErrUserNotFound
	}

	// Set status to active
	user.Status = domain.UserStatusActive
	user.UpdatedAt = time.Now()

	// Save changes
	err = s.userRepo.Save(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		cacheKey := fmt.Sprintf("user:%d", id)
		_ = s.cache.Delete(ctx, cacheKey)
	}

	// Publish user activated event
	if s.eventPublisher != nil {
		event := infrastructure.Event{
			Type: "user.activated",
			Payload: map[string]interface{}{
				"user_id": user.ID,
			},
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.eventPublisher.Publish(ctx, "users", event)
		}()
	}

	return nil
}
