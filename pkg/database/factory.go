package database

import (
	"fmt"

	"github.com/shurco/mycart/pkg/runtime"
)

// New creates a database instance based on the provided configuration.
// If Type is empty, auto-detects based on runtime environment.
// If Type is "sqlite" and Path is empty, uses runtime default or ":memory:".
func New(config Config) (Database, error) {
	// Track if type was auto-detected to enable full auto-configuration
	autoDetected := false

	// Auto-detect database type if not specified
	if config.Type == "" {
		autoDetected = true
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
		accountID := config.AccountID
		databaseID := config.DatabaseID
		apiToken := config.APIToken

		// Only auto-populate credentials if type was auto-detected
		if autoDetected {
			if accountID == "" {
				accountID = runtime.GetD1AccountID()
			}
			if databaseID == "" {
				databaseID = runtime.GetD1DatabaseID()
			}
			if apiToken == "" {
				apiToken = runtime.GetD1APIToken()
			}
		}

		return NewD1(accountID, databaseID, apiToken)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
