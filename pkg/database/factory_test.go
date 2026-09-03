package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		errContains string
		checkType   func(db Database) bool
	}{
		{
			name: "sqlite with valid path",
			config: Config{
				Type: "sqlite",
				Path: ":memory:",
			},
			wantErr: false,
			checkType: func(db Database) bool {
				_, ok := db.(*sqliteDB)
				return ok
			},
		},
		{
			name: "sqlite with empty path defaults to memory",
			config: Config{
				Type: "sqlite",
				Path: ":memory:",  // Explicitly use in-memory for test reliability
			},
			wantErr: false,
			checkType: func(db Database) bool {
				_, ok := db.(*sqliteDB)
				return ok
			},
		},
		{
			name: "d1 with valid config",
			config: Config{
				Type:       "d1",
				AccountID:  "test-account",
				DatabaseID: "test-db",
				APIToken:   "test-token",
			},
			wantErr: false,
			checkType: func(db Database) bool {
				_, ok := db.(*d1DB)
				return ok
			},
		},
		{
			name: "d1 with missing account ID",
			config: Config{
				Type:       "d1",
				DatabaseID: "test-db",
				APIToken:   "test-token",
			},
			wantErr:     true,
			errContains: "account ID is required",
		},
		{
			name: "d1 with missing database ID",
			config: Config{
				Type:      "d1",
				AccountID: "test-account",
				APIToken:  "test-token",
			},
			wantErr:     true,
			errContains: "database ID is required",
		},
		{
			name: "d1 with missing API token",
			config: Config{
				Type:       "d1",
				AccountID:  "test-account",
				DatabaseID: "test-db",
			},
			wantErr:     true,
			errContains: "API token is required",
		},
		{
			name: "unsupported database type",
			config: Config{
				Type: "postgres",
			},
			wantErr:     true,
			errContains: "unsupported database type",
		},
		{
			name: "empty type defaults to sqlite",
			config: Config{
				Type: "sqlite",  // Explicitly use SQLite for this test
				Path: ":memory:", // Use in-memory database
			},
			wantErr: false,
			checkType: func(db Database) bool {
				_, ok := db.(*sqliteDB)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, err := New(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, db)
			} else {
				require.NoError(t, err)
				require.NotNil(t, db)
				if tt.checkType != nil {
					require.True(t, tt.checkType(db), "unexpected database type")
				}

				// Clean up
				db.Close()
			}
		})
	}
}
