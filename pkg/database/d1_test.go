package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestD1_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		accountID   string
		databaseID  string
		apiToken    string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid configuration",
			accountID:  "test-account-id",
			databaseID: "test-database-id",
			apiToken:   "test-api-token",
			wantErr:    false,
		},
		{
			name:        "missing account ID",
			accountID:   "",
			databaseID:  "test-database-id",
			apiToken:    "test-api-token",
			wantErr:     true,
			errContains: "account ID is required",
		},
		{
			name:        "missing database ID",
			accountID:   "test-account-id",
			databaseID:  "",
			apiToken:    "test-api-token",
			wantErr:     true,
			errContains: "database ID is required",
		},
		{
			name:        "missing API token",
			accountID:   "test-account-id",
			databaseID:  "test-database-id",
			apiToken:    "",
			wantErr:     true,
			errContains: "API token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, err := NewD1(tt.accountID, tt.databaseID, tt.apiToken)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)
			}
		})
	}
}

func TestD1_Transaction_NoOp(t *testing.T) {
	t.Parallel()

	db, err := NewD1("test-account", "test-db", "test-token")
	require.NoError(t, err)

	ctx := context.Background()

	// Begin transaction should succeed
	tx, err := db.Begin(ctx)
	require.NoError(t, err)
	require.NotNil(t, tx)

	// Commit should be no-op (no error)
	err = tx.Commit()
	require.NoError(t, err)
}

func TestD1_Transaction_Rollback(t *testing.T) {
	t.Parallel()

	db, err := NewD1("test-account", "test-db", "test-token")
	require.NoError(t, err)

	ctx := context.Background()

	tx, err := db.Begin(ctx)
	require.NoError(t, err)

	// Rollback should be no-op (no error)
	err = tx.Rollback()
	require.NoError(t, err)
}

func TestD1_Close(t *testing.T) {
	t.Parallel()

	db, err := NewD1("test-account", "test-db", "test-token")
	require.NoError(t, err)

	// Close should be no-op (no error)
	err = db.Close()
	require.NoError(t, err)
}
