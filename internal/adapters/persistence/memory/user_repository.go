package memory

import (
	"context"
	"ports-and-adapters-architecture/internal/domain"
	"sync"
)

// InMemoryUserRepository implements UserRepository interface for testing
type InMemoryUserRepository struct {
	mu     sync.RWMutex
	users  map[int]*domain.User
	nextID int
}

// NewInMemoryUserRepository creates a new in-memory user repository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:  make(map[int]*domain.User),
		nextID: 1,
	}
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, nil
	}

	// Return a copy to prevent external modifications
	userCopy := *user
	return &userCopy, nil
}

func (r *InMemoryUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, nil
}

func (r *InMemoryUserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Phone == phone {
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, nil
}

func (r *InMemoryUserRepository) Save(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == 0 {
		user.ID = r.nextID
		r.nextID++
	}

	userCopy := *user
	r.users[user.ID] = &userCopy

	return nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.users, id)
	return nil
}