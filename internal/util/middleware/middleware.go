package middleware

import (
	"context"
	"net/http"
	"strings"

	"template-go-jwt/internal/util/auth"
)

type ctxKey string

const (
	UserIDKey ctxKey = "user_id"
	RoleKey   ctxKey = "role"
)

// AuthMiddleware validates JWT from Authorization header and injects user id into context.
func AuthMiddleware(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(authz, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}
		token := parts[1]
		claims, err := auth.ParseToken(token, secret)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks that the request context has the required role.
func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(RoleKey)
		if v == nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		gotRole, _ := v.(string)
		if gotRole != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
