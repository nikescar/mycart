package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/shurco/mycart/internal/database/sqlc"
)

// SettingAdapter provides backward-compatible interface for setting queries
type SettingAdapter struct {
	db *Database
}

// NewSettingAdapter creates a new setting adapter
func NewSettingAdapter(db *Database) *SettingAdapter {
	return &SettingAdapter{db: db}
}

// GetSettingByKey retrieves a setting by key and returns it as a map
func (a *SettingAdapter) GetSettingByKey(ctx context.Context, key string) (map[string]struct{ Value any }, error) {
	setting, err := a.db.queries.GetSettingByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	result := map[string]struct{ Value any }{
		key: {Value: setting.Value.String},
	}
	return result, nil
}

// UpdateSettingByKey updates a setting value by key
func (a *SettingAdapter) UpdateSettingByKey(ctx context.Context, key string, value any) error {
	valueStr, err := convertToString(value)
	if err != nil {
		return err
	}

	return a.db.queries.UpdateSettingByKey(ctx, sqlc.UpdateSettingByKeyParams{
		Value: sql.NullString{String: valueStr, Valid: true},
		Key:   key,
	})
}

// convertToString converts various types to string for database storage
func convertToString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
