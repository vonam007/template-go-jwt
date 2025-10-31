package auth

import (
	"testing"
	"time"
)

func TestGenerateParseToken(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"
	role := "user"
	token, err := GenerateToken(userID, role, secret, time.Minute)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected user id %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Fatalf("expected role %s, got %s", role, claims.Role)
	}
}
