package database

import "fmt"

// New creates a database instance based on the provided configuration.
// If Type is empty, defaults to "sqlite".
// If Type is "sqlite" and Path is empty, defaults to ":memory:".
func New(config Config) (Database, error) {
	// Default to sqlite if no type specified
	if config.Type == "" {
		config.Type = "sqlite"
	}

	switch config.Type {
	case "sqlite":
		path := config.Path
		if path == "" {
			path = ":memory:"
		}
		return NewSQLite(path)

	case "d1":
		return NewD1(config.AccountID, config.DatabaseID, config.APIToken)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
