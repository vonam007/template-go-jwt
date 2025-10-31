package db

import (
	"time"

	"gorm.io/gorm"

	"template-go-jwt/internal/models"
)

// MigrationRecord records applied migrations.
type MigrationRecord struct {
	Name      string    `gorm:"primaryKey;size:255"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

// Migration is a single named migration.
type Migration struct {
	Name string
	Up   func(*gorm.DB) error
}

// RunMigrations runs all provided migrations in order. It records applied
// migrations in the `migration_records` table so the same migration won't
// run twice.
func RunMigrations(db *gorm.DB, migrations []Migration) error {
	// ensure the migration records table exists
	if err := db.AutoMigrate(&MigrationRecord{}); err != nil {
		return err
	}

	for _, m := range migrations {
		var rec MigrationRecord
		if err := db.First(&rec, "name = ?", m.Name).Error; err == nil {
			// migration already applied
			continue
		}
		// run migration
		if err := m.Up(db); err != nil {
			return err
		}
		// insert record
		rec = MigrationRecord{Name: m.Name, AppliedAt: time.Now()}
		if err := db.Create(&rec).Error; err != nil {
			return err
		}
	}
	return nil
}

// DefaultMigrations returns a slice of migrations to run for this template.
func DefaultMigrations() []Migration {
	return []Migration{
		{
			Name: "20251031_create_users",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&models.User{})
			},
		},
	}
}
