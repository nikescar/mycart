package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shurco/mycart/internal/store/db/postgres"
	"github.com/shurco/mycart/internal/store/db/sqlite"
)

// Init initializes function pointers based on database type
// Call once at application startup after database connection
func Init(sqlDB *sql.DB, dbType string) error {
	if sqlDB == nil {
		return fmt.Errorf("database connection is nil")
	}

	switch dbType {
	case "postgres", "postgresql":
		initPostgres(sqlDB)
	case "sqlite", "sqlite3":
		initSQLite(sqlDB)
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	return nil
}

// initPostgres assigns PostgreSQL sqlc implementations to function pointers
func initPostgres(sqlDB *sql.DB) {
	q := postgres.New(sqlDB)

	// Wrap methods that return postgres-specific types
	GetSettingByKeyFunc = func(ctx context.Context, key string) (Setting, error) {
		pgSetting, err := q.GetSettingByKey(ctx, key)
		if err != nil {
			return Setting{}, err
		}
		return FromPostgresSetting(pgSetting), nil
	}

	DeleteSettingFunc = q.DeleteSetting

	ListSettingsFunc = func(ctx context.Context) ([]Setting, error) {
		pgSettings, err := q.ListSettings(ctx)
		if err != nil {
			return nil, err
		}
		result := make([]Setting, len(pgSettings))
		for i, s := range pgSettings {
			result[i] = FromPostgresSetting(s)
		}
		return result, nil
	}

	GetSessionFunc = func(ctx context.Context, key string) (Session, error) {
		pgSession, err := q.GetSession(ctx, key)
		if err != nil {
			return Session{}, err
		}
		return FromPostgresSession(pgSession), nil
	}

	DeleteSessionFunc = q.DeleteSession

	// Wrap methods that need parameter type conversion
	CreateSettingFunc = func(ctx context.Context, arg CreateSettingParams) (Setting, error) {
		pgParams := postgres.CreateSettingParams{
			ID:    arg.ID,
			Key:   arg.Key,
			Value: arg.Value,
		}
		pgSetting, err := q.CreateSetting(ctx, pgParams)
		if err != nil {
			return Setting{}, err
		}
		return Setting{
			ID:    pgSetting.ID,
			Key:   pgSetting.Key,
			Value: pgSetting.Value,
		}, nil
	}

	UpdateSettingFunc = func(ctx context.Context, arg UpdateSettingParams) error {
		pgParams := postgres.UpdateSettingParams{
			Value: arg.Value,
			Key:   arg.Key,
		}
		return q.UpdateSetting(ctx, pgParams)
	}

	CreateSessionFunc = func(ctx context.Context, arg CreateSessionParams) error {
		pgParams := postgres.CreateSessionParams{
			Key:   arg.Key,
			Value: arg.Value,
			Expires: sql.NullInt32{
				Int32: int32(arg.Expires.Int64),
				Valid: arg.Expires.Valid,
			},
		}
		return q.CreateSession(ctx, pgParams)
	}

	UpdateSessionFunc = func(ctx context.Context, arg UpdateSessionParams) error {
		pgParams := postgres.UpdateSessionParams{
			Value: arg.Value,
			Expires: sql.NullInt32{
				Int32: int32(arg.Expires.Int64),
				Valid: arg.Expires.Valid,
			},
			Key: arg.Key,
		}
		return q.UpdateSession(ctx, pgParams)
	}
}

// initSQLite assigns SQLite sqlc implementations to function pointers
func initSQLite(sqlDB *sql.DB) {
	q := sqlite.New(sqlDB)

	// Wrap methods that return sqlite-specific types
	GetSettingByKeyFunc = func(ctx context.Context, key string) (Setting, error) {
		sqliteSetting, err := q.GetSettingByKey(ctx, key)
		if err != nil {
			return Setting{}, err
		}
		return FromSQLiteSetting(sqliteSetting), nil
	}

	DeleteSettingFunc = q.DeleteSetting

	ListSettingsFunc = func(ctx context.Context) ([]Setting, error) {
		sqliteSettings, err := q.ListSettings(ctx)
		if err != nil {
			return nil, err
		}
		result := make([]Setting, len(sqliteSettings))
		for i, s := range sqliteSettings {
			result[i] = FromSQLiteSetting(s)
		}
		return result, nil
	}

	GetSessionFunc = func(ctx context.Context, key string) (Session, error) {
		sqliteSession, err := q.GetSession(ctx, key)
		if err != nil {
			return Session{}, err
		}
		return FromSQLiteSession(sqliteSession), nil
	}

	DeleteSessionFunc = q.DeleteSession

	// Wrap methods that need parameter type conversion
	CreateSettingFunc = func(ctx context.Context, arg CreateSettingParams) (Setting, error) {
		sqliteParams := sqlite.CreateSettingParams{
			ID:    arg.ID,
			Key:   arg.Key,
			Value: arg.Value,
		}
		sqliteSetting, err := q.CreateSetting(ctx, sqliteParams)
		if err != nil {
			return Setting{}, err
		}
		return Setting{
			ID:    sqliteSetting.ID,
			Key:   sqliteSetting.Key,
			Value: sqliteSetting.Value,
		}, nil
	}

	UpdateSettingFunc = func(ctx context.Context, arg UpdateSettingParams) error {
		sqliteParams := sqlite.UpdateSettingParams{
			Value: arg.Value,
			Key:   arg.Key,
		}
		return q.UpdateSetting(ctx, sqliteParams)
	}

	CreateSessionFunc = func(ctx context.Context, arg CreateSessionParams) error {
		sqliteParams := sqlite.CreateSessionParams{
			Key:     arg.Key,
			Value:   arg.Value,
			Expires: arg.Expires, // sqlite uses Int64, matches our unified type
		}
		return q.CreateSession(ctx, sqliteParams)
	}

	UpdateSessionFunc = func(ctx context.Context, arg UpdateSessionParams) error {
		sqliteParams := sqlite.UpdateSessionParams{
			Value:   arg.Value,
			Expires: arg.Expires, // sqlite uses Int64, matches our unified type
			Key:     arg.Key,
		}
		return q.UpdateSession(ctx, sqliteParams)
	}
}
