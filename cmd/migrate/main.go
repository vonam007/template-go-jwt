package main

import (
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"template-go-jwt/internal/util/config"
)

func main() {
	dbcfg := config.LoadDB()
	dbURL := dbcfg.PostgresURL()

	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		log.Fatalf("migration create failed: %v", err)
	}
	defer func() {
		m.Close()
	}()

	if len(os.Args) > 1 && os.Args[1] == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration down failed: %v", err)
		}
		log.Println("migrations rolled back")
		return
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration up failed: %v", err)
	}
	log.Println("migrations applied")
}
