package database

import (
	"fmt"

	"github.com/shurco/mycart/pkg/runtime"
)

// New creates a database instance based on the provided configuration.
// If Type is empty, auto-detects based on runtime environment.
// If Type is "sqlite" and Path is empty, uses runtime default or ":memory:".
func New(config Config) (Database, error) {
	// Auto-detect database type if not specified
	if config.Type == "" {
		if runtime.IsCloudflare() {
			config.Type = "d1"
		} else {
			config.Type = "sqlite"
		}
	}

	switch config.Type {
	case "sqlite":
		path := config.Path
		if path == "" {
			dbPath := runtime.GetDatabasePath()
			if dbPath != "" {
				path = dbPath
			} else {
				path = ":memory:"
			}
		}
		return NewSQLite(path)

	case "d1":
		return NewD1(config.AccountID, config.DatabaseID, config.APIToken)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
