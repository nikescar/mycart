package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// RetryConfig configures the retry behavior for database connections.
type RetryConfig struct {
	MaxAttempts  int           // Maximum number of connection attempts
	InitialDelay time.Duration // Initial delay between retries
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Backoff multiplier
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// ConnectWithRetry attempts to connect to the database with exponential backoff retry logic.
func ConnectWithRetry(cfg *Config) (Database, error) {
	return ConnectWithRetryConfig(cfg, DefaultRetryConfig())
}

// ConnectWithRetryConfig attempts to connect with a custom retry configuration.
func ConnectWithRetryConfig(cfg *Config, retry RetryConfig) (Database, error) {
	var db *sql.DB
	var err error
	var driver, connStr string

	// Determine driver and connection string based on database type
	switch cfg.Type {
	case "sqlite":
		driver = "sqlite"
		connStr = cfg.SQLite.Path
	case "postgresql", "postgres":
		driver = "postgres"
		connStr = cfg.PostgreSQL.ConnectionString()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	// Retry with exponential backoff
	delay := retry.InitialDelay
	for attempt := 1; attempt <= retry.MaxAttempts; attempt++ {
		db, err = sql.Open(driver, connStr)
		if err != nil {
			if attempt == retry.MaxAttempts {
				return nil, fmt.Errorf("failed to open database after %d attempts: %w", retry.MaxAttempts, err)
			}
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * retry.Multiplier)
			if delay > retry.MaxDelay {
				delay = retry.MaxDelay
			}
			continue
		}

		// Test the connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			// Connection successful - configure connection pool
			configureConnectionPool(db, cfg)

			// Return the appropriate adapter
			if cfg.Type == "sqlite" {
				return NewSQLiteAdapter(db), nil
			}
			return NewPostgresAdapter(db), nil
		}

		// Connection failed, close and retry
		db.Close()

		if attempt == retry.MaxAttempts {
			return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", retry.MaxAttempts, err)
		}

		time.Sleep(delay)
		delay = time.Duration(float64(delay) * retry.Multiplier)
		if delay > retry.MaxDelay {
			delay = retry.MaxDelay
		}
	}

	return nil, fmt.Errorf("failed to connect to database: %w", err)
}

// configureConnectionPool sets up connection pooling parameters.
func configureConnectionPool(db *sql.DB, cfg *Config) {
	if cfg.Type == "postgresql" || cfg.Type == "postgres" {
		db.SetMaxOpenConns(cfg.PostgreSQL.MaxOpenConns)
		db.SetMaxIdleConns(cfg.PostgreSQL.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.PostgreSQL.ConnMaxLifetime)
	} else {
		// SQLite with WAL mode supports multiple concurrent readers
		// Allow up to 5 connections to support nested queries without deadlock
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(0)
	}
}

// TestConnection tests if a database connection can be established.
func TestConnection(cfg *Config) error {
	db, err := ConnectWithRetry(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.Ping(ctx)
}
