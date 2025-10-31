package main

import (
	"log"
	"net/http"
	"time"

	"template-go-jwt/internal/controllers"
	"template-go-jwt/internal/db"
	"template-go-jwt/internal/models"
	"template-go-jwt/internal/repositories"
	"template-go-jwt/internal/server"
	"template-go-jwt/internal/services"

	"template-go-jwt/internal/util/config"
)

func main() {
	cfg := config.Load()
	// attempt to connect to DB (if available) — used for later DB wiring
	dbCfg := config.LoadDB()
	gdb, err := db.ConnectWithRetry(dbCfg, 5, 2*time.Second)
	if err != nil {
		log.Printf("warning: could not connect to DB: %v\n", err)
		gdb = nil
	} else {
		defer func() {
			if err := db.Close(gdb); err != nil {
				log.Printf("error closing db: %v", err)
			}
		}()
	}

	// Wire MVC components: repo -> service -> controller
	// NOTE: default is in-memory repo. Replace with Postgres repo implementation when ready.
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
