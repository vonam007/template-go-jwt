package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"template-go-jwt/internal/util/config"
)

// Connect opens a GORM DB connection using the provided DB config.
// Caller should handle lifecycle and close underlying connection when needed.
func Connect(ctx context.Context, cfg config.DBConfig) (*gorm.DB, error) {
	dsn := cfg.PostgresDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	// Try pinging the database using the generic database/sql DB if available.
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm db(): %w", err)
	}
	// set some sensible defaults
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctxPing); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return db, nil
}

// ConnectWithRetry tries to connect multiple times with backoff.
func ConnectWithRetry(cfg config.DBConfig, attempts int, delay time.Duration) (*gorm.DB, error) {
	var lastErr error
	for i := 0; i < attempts; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		db, err := Connect(ctx, cfg)
		cancel()
		if err == nil {
			return db, nil
		}
		lastErr = err
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("gorm connect attempts=%d failed: %w", attempts, lastErr)
}

// Close closes the underlying sql DB connection obtained from GORM.
func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
