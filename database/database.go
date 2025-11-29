package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sentinel-official/sentinel-dvpnx/database/models"
)

// New initializes a new database connection with the specified file path and configuration.
// It also performs migrations to ensure the database schema is up to date with the models.
func New(file string, cfg *gorm.Config) (*gorm.DB, error) {
	// Build the SQLite DSN
	dsn := file + "?_busy_timeout=5000&_journal_mode=WAL"

	// Open a database connection using the provided filepath and configuration.
	db, err := gorm.Open(sqlite.Open(dsn), cfg)
	if err != nil {
		return nil, fmt.Errorf("opening database file %q: %w", file, err)
	}

	// List of models to be migrated.
	items := []any{
		&models.Session{},
	}

	// Run migrations to apply the schema of the models to the database.
	if err := db.AutoMigrate(items...); err != nil {
		return nil, fmt.Errorf("auto migrating %d model(s): %w", len(items), err)
	}

	// Return the database connection if everything is successful.
	return db, nil
}

// NewDefault uses default configuration settings and calls the New function to initialize the database.
func NewDefault(file string) (*gorm.DB, error) {
	// Define default GORM configuration settings.
	cfg := gorm.Config{
		Logger:         logger.Discard,
		PrepareStmt:    false,
		TranslateError: true,
	}

	// Call New with the default configuration.
	return New(file, &cfg)
}
