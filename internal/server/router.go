package server

import (
	"net/http"

	"template-go-jwt/internal/controllers"

	"template-go-jwt/internal/util/config"
	"template-go-jwt/internal/util/middleware"
	"template-go-jwt/internal/util/response"
)

// NewRouter registers routes and returns an http.Handler.
func NewRouter(cfg config.Config, ctrl *controllers.AuthController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"message": "hello"})
	})

	mux.HandleFunc("/login", ctrl.Login)

	mux.Handle("/protected", middleware.AuthMiddleware(cfg.JWTSecret, http.HandlerFunc(ctrl.Protected)))

	// Admin routes — only accessible to role == "admin"
	mux.Handle("/admin/", middleware.AuthMiddleware(cfg.JWTSecret, middleware.RequireRole("admin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"message": "admin area"})
	}))))

	return mux
}
