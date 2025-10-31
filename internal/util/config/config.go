package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port      string
	JWTSecret string
	JWTTTL    time.Duration
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func Load() Config {
	port := getenv("PORT", "8080")
	secret := getenv("JWT_SECRET", "dev-secret-change-me")
	ttlStr := getenv("JWT_TTL_SECONDS", "3600")
	ttlInt, _ := strconv.Atoi(ttlStr)
	return Config{
		Port:      port,
		JWTSecret: secret,
		JWTTTL:    time.Duration(ttlInt) * time.Second,
	}
}

// Note: database config is optional for the template. These env vars are
// provided so a future DB integration or docker-compose can wire a Postgres
// instance and the application can construct a DSN from env values.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func LoadDB() DBConfig {
	return DBConfig{
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "5432"),
		User:     getenv("DB_USER", "postgres"),
		Password: getenv("DB_PASSWORD", "postgres"),
		Name:     getenv("DB_NAME", "template"),
	}
}

// PostgresDSN returns a lib/pq / pgx compatible connection string.
func (d DBConfig) PostgresDSN() string {
	// host=... port=... user=... password=... dbname=... sslmode=disable
	return "host=" + d.Host + " port=" + d.Port + " user=" + d.User + " password=" + d.Password + " dbname=" + d.Name + " sslmode=disable"
}

// PostgresURL returns a postgres URL suitable for golang-migrate when using
// the file source (e.g. postgres://user:pass@host:port/dbname?sslmode=disable).
func (d DBConfig) PostgresURL() string {
	// Note: this is a simple builder; if your password contains special
	// characters you should url.QueryEscape it when generating production URLs.
	return "postgres://" + d.User + ":" + d.Password + "@" + d.Host + ":" + d.Port + "/" + d.Name + "?sslmode=disable"
}
