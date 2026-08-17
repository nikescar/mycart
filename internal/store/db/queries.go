package db

import "context"

// Function pointer variables for database operations
// Initialized once at startup by Init() based on database type (postgres or sqlite)
// Provides zero-overhead database abstraction layer

var (
	// Settings operations
	GetSettingByKeyFunc func(ctx context.Context, key string) (Setting, error)
	CreateSettingFunc   func(ctx context.Context, arg CreateSettingParams) (Setting, error)
	UpdateSettingFunc   func(ctx context.Context, arg UpdateSettingParams) error
	DeleteSettingFunc   func(ctx context.Context, key string) error
	ListSettingsFunc    func(ctx context.Context) ([]Setting, error)

	// Session operations
	GetSessionFunc    func(ctx context.Context, key string) (Session, error)
	CreateSessionFunc func(ctx context.Context, arg CreateSessionParams) error
	UpdateSessionFunc func(ctx context.Context, arg UpdateSessionParams) error
	DeleteSessionFunc func(ctx context.Context, key string) error
)

// Additional function pointers can be added here as needed during migration
// Pattern: declare var, then assign in init.go's initPostgres/initSQLite functions
