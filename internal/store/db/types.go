package db

import (
	"database/sql"

	"github.com/shurco/mycart/internal/store/db/postgres"
	"github.com/shurco/mycart/internal/store/db/sqlite"
)

// Unified types that abstract postgres/sqlite differences
// Key difference: postgres uses sql.NullInt32 for integers, sqlite uses sql.NullInt64

// Setting is the unified type for database settings
// Compatible with both postgres.Setting and sqlite.Setting
type Setting struct {
	ID    string
	Key   string
	Value sql.NullString
}

// Session is the unified type for sessions
// Uses Int64 for Expires to accommodate both postgres (Int32) and sqlite (Int64)
type Session struct {
	Key     string
	Value   sql.NullString
	Expires sql.NullInt64 // Unified as Int64 (postgres Int32 converts to this)
}

// CreateSettingParams for CreateSetting operation
type CreateSettingParams struct {
	ID    string
	Key   string
	Value sql.NullString
}

// UpdateSettingParams for UpdateSetting operation
type UpdateSettingParams struct {
	Value sql.NullString
	Key   string
}

// CreateSessionParams for CreateSession operation
type CreateSessionParams struct {
	Key     string
	Value   sql.NullString
	Expires sql.NullInt64 // Unified as Int64
}

// UpdateSessionParams for UpdateSession operation
type UpdateSessionParams struct {
	Value   sql.NullString
	Expires sql.NullInt64 // Unified as Int64
	Key     string
}

// ToPostgresSetting converts unified Setting to postgres.Setting
func ToPostgresSetting(s Setting) postgres.Setting {
	return postgres.Setting{
		ID:    s.ID,
		Key:   s.Key,
		Value: s.Value,
	}
}

// ToSQLiteSetting converts unified Setting to sqlite.Setting
func ToSQLiteSetting(s Setting) sqlite.Setting {
	return sqlite.Setting{
		ID:    s.ID,
		Key:   s.Key,
		Value: s.Value,
	}
}

// FromPostgresSetting converts postgres.Setting to unified Setting
func FromPostgresSetting(s postgres.Setting) Setting {
	return Setting{
		ID:    s.ID,
		Key:   s.Key,
		Value: s.Value,
	}
}

// FromSQLiteSetting converts sqlite.Setting to unified Setting
func FromSQLiteSetting(s sqlite.Setting) Setting {
	return Setting{
		ID:    s.ID,
		Key:   s.Key,
		Value: s.Value,
	}
}

// FromPostgresSession converts postgres.Session to unified Session
func FromPostgresSession(s postgres.Session) Session {
	return Session{
		Key:   s.Key,
		Value: s.Value,
		Expires: sql.NullInt64{
			Int64: int64(s.Expires.Int32),
			Valid: s.Expires.Valid,
		},
	}
}

// FromSQLiteSession converts sqlite.Session to unified Session
func FromSQLiteSession(s sqlite.Session) Session {
	return Session{
		Key:     s.Key,
		Value:   s.Value,
		Expires: s.Expires,
	}
}
