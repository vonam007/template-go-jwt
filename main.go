package main

import (
	"log"
	"net/http"
	"time"

	"template-go-jwt/internal/controllers"
	"template-go-jwt/internal/models"
	"template-go-jwt/internal/repositories"
	"template-go-jwt/internal/server"
	"template-go-jwt/internal/services"

	"template-go-jwt/internal/util/config"
)

func main() {
	cfg := config.Load()

	// Wire MVC components: repo -> service -> controller
	repo := repositories.NewInMemoryUserRepo()
	// seed sample users so login works out-of-the-box
	repo.Seed(&models.User{ID: "alice", Name: "Alice", Role: "admin"})
	repo.Seed(&models.User{ID: "bob", Name: "Bob", Role: "user"})

	svc := services.NewAuthService(repo, cfg.JWTSecret, cfg.JWTTTL)
	ctrl := controllers.NewAuthController(svc)

	// Build router in separate package so main stays small.
	router := server.NewRouter(cfg, ctrl)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("starting server on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
