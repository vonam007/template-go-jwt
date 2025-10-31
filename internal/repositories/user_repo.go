package repositories

import (
	"errors"

	"template-go-jwt/internal/models"
)

// UserRepository defines storage operations for users.
type UserRepository interface {
	GetByID(id string) (*models.User, error)
}

// InMemoryUserRepo is a tiny in-memory implementation for templates and tests.
type InMemoryUserRepo struct {
	users map[string]*models.User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{users: map[string]*models.User{}}
}

func (r *InMemoryUserRepo) Seed(u *models.User) {
	if u == nil {
		return
	}
	r.users[u.ID] = u
}

func (r *InMemoryUserRepo) GetByID(id string) (*models.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}
