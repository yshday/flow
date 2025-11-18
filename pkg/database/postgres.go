package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

// Config holds database configuration
type Config struct {
	URL            string
	MaxConnections int
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if config.MaxConnections > 0 {
		db.SetMaxOpenConns(config.MaxConnections)
	} else {
		db.SetMaxOpenConns(25) // default
	}
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("database connection established", "max_connections", config.MaxConnections)

	return db, nil
}
