package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

// DetectedDatabase represents a discovered database configuration.
type DetectedDatabase struct {
	Type        string // "sqlite" or "postgresql"
	Source      string // Description of where it was detected from
	Description string // Human-readable description
	Available   bool   // Whether the database is currently accessible
	Config      *Config // Configuration for this database
}

// DetectDatabases discovers available database options.
func DetectDatabases() []DetectedDatabase {
	var databases []DetectedDatabase

	// Check for DATABASE_URL (PostgreSQL)
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		cfg := &Config{
			Type:       "postgresql",
			PostgreSQL: parseConnectionURL(databaseURL),
		}

		available := testDatabaseAvailability(cfg)
		databases = append(databases, DetectedDatabase{
			Type:        "postgresql",
			Source:      "DATABASE_URL",
			Description: fmt.Sprintf("PostgreSQL at %s:%d", cfg.PostgreSQL.Host, cfg.PostgreSQL.Port),
			Available:   available,
			Config:      cfg,
		})
	}

	// Check for individual PostgreSQL environment variables
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg, err := LoadConfig()
		if err == nil && cfg.Type == "postgresql" {
			available := testDatabaseAvailability(cfg)
			databases = append(databases, DetectedDatabase{
				Type:        "postgresql",
				Source:      "DB_HOST",
				Description: fmt.Sprintf("PostgreSQL at %s:%d", cfg.PostgreSQL.Host, cfg.PostgreSQL.Port),
				Available:   available,
				Config:      cfg,
			})
		}
	}

	// Check for existing SQLite database
	sqlitePath := "./lc_base/data.db"
	if _, err := os.Stat(sqlitePath); err == nil {
		cfg := &Config{
			Type: "sqlite",
			SQLite: SQLiteConfig{
				Path: sqlitePath,
			},
		}

		available := testDatabaseAvailability(cfg)
		databases = append(databases, DetectedDatabase{
			Type:        "sqlite",
			Source:      "existing file",
			Description: fmt.Sprintf("Existing SQLite database at %s", sqlitePath),
			Available:   available,
			Config:      cfg,
		})
	}

	// Always offer new SQLite option
	newSqlitePath := "./lc_base/data.db"
	databases = append(databases, DetectedDatabase{
		Type:        "sqlite",
		Source:      "new installation",
		Description: fmt.Sprintf("New SQLite database at %s", newSqlitePath),
		Available:   true, // Always available for new installation
		Config: &Config{
			Type: "sqlite",
			SQLite: SQLiteConfig{
				Path: newSqlitePath,
			},
		},
	})

	return databases
}

// testDatabaseAvailability checks if a database is accessible.
func testDatabaseAvailability(cfg *Config) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var db *sql.DB
	var err error

	switch cfg.Type {
	case "sqlite":
		db, err = sql.Open("sqlite", cfg.SQLite.Path)
	case "postgresql", "postgres":
		db, err = sql.Open("postgres", cfg.PostgreSQL.ConnectionString())
	default:
		return false
	}

	if err != nil {
		return false
	}
	defer db.Close()

	err = db.PingContext(ctx)
	return err == nil
}

// DetectPreferredDatabase returns the best available database option.
// Preference order: existing SQLite > configured PostgreSQL > new SQLite
func DetectPreferredDatabase() *DetectedDatabase {
	databases := DetectDatabases()

	// First, try existing SQLite
	for _, db := range databases {
		if db.Type == "sqlite" && db.Source == "existing file" && db.Available {
			return &db
		}
	}

	// Next, try configured PostgreSQL
	for _, db := range databases {
		if db.Type == "postgresql" && db.Available {
			return &db
		}
	}

	// Finally, default to new SQLite
	for _, db := range databases {
		if db.Type == "sqlite" && db.Source == "new installation" {
			return &db
		}
	}

	return nil
}
