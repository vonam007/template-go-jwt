package services

import (
	"time"

	"template-go-jwt/internal/repositories"
	"template-go-jwt/internal/util/auth"
)

// AuthService coordinates auth-related operations and storage checks.
type AuthService struct {
	repo   repositories.UserRepository
	secret string
	ttl    time.Duration
}

func NewAuthService(repo repositories.UserRepository, secret string, ttl time.Duration) *AuthService {
	return &AuthService{repo: repo, secret: secret, ttl: ttl}
}

// Login creates a JWT for a user if repository recognizes the user.
// For template simplicity, if repo is nil we allow token creation.
func (s *AuthService) Login(userID string) (string, error) {
	var role string
	if s.repo != nil {
		u, err := s.repo.GetByID(userID)
		if err != nil {
			return "", err
		}
		role = u.Role
	}
	return auth.GenerateToken(userID, role, s.secret, s.ttl)
}

// ValidateToken parses the token and optionally verifies user existence.
func (s *AuthService) ValidateToken(token string) (string, error) {
	claims, err := auth.ParseToken(token, s.secret)
	if err != nil {
		return "", err
	}
	if s.repo != nil {
		if _, err := s.repo.GetByID(claims.UserID); err != nil {
			return "", err
		}
	}
	return claims.UserID, nil
}
