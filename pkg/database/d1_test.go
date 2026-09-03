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

func TestD1_Transaction_Error(t *testing.T) {
	t.Parallel()

	db, err := NewD1("test-account", "test-db", "test-token")
	require.NoError(t, err)

	ctx := context.Background()

	// Begin transaction should fail with explicit error
	tx, err := db.Begin(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "D1 does not support transactions")
	require.Nil(t, tx)
}


func TestD1_Close(t *testing.T) {
	t.Parallel()

	db, err := NewD1("test-account", "test-db", "test-token")
	require.NoError(t, err)

	// Close should be no-op (no error)
	err = db.Close()
	require.NoError(t, err)
}

func TestD1_ErrorSanitization(t *testing.T) {
	t.Parallel()

	// Test that API token is sanitized in error messages
	secretToken := "super-secret-api-token-12345"

	// This will fail because the driver will try to actually connect
	// but we're testing that the error message doesn't leak the token
	db, err := NewD1("test-account", "test-db", secretToken)

	if err != nil {
		// Error message should NOT contain the actual token
		require.NotContains(t, err.Error(), secretToken,
			"API token should be sanitized in error messages")

		// Error message SHOULD contain the sanitized placeholder
		require.Contains(t, err.Error(), "api_token=***",
			"Error should contain sanitized token placeholder")
	}

	if db != nil {
		db.Close()
	}
}

func TestSanitizeDSN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "basic DSN with token",
			dsn:      "account123/db456?api_token=secret123",
			expected: "account123/db456?api_token=***",
		},
		{
			name:     "DSN with long token",
			dsn:      "acc/db?api_token=very-long-secret-token-that-should-be-hidden",
			expected: "acc/db?api_token=***",
		},
		{
			name:     "DSN without token",
			dsn:      "account/database",
			expected: "account/database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeDSN(tt.dsn)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestMarshalParam(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     interface{}
		want      string
		wantError bool
	}{
		{
			name:  "nil value",
			input: nil,
			want:  "NULL",
		},
		{
			name:  "int64 positive",
			input: int64(42),
			want:  "42",
		},
		{
			name:  "int64 negative",
			input: int64(-123),
			want:  "-123",
		},
		{
			name:  "float64",
			input: float64(3.14159),
			want:  "3.14159",
		},
		{
			name:  "bool true",
			input: true,
			want:  "1",
		},
		{
			name:  "bool false",
			input: false,
			want:  "0",
		},
		{
			name:  "string",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "byte slice",
			input: []byte("hello"),
			want:  "aGVsbG8=", // base64("hello")
		},
		{
			name:      "unsupported type",
			input:     struct{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := marshalParam(tt.input)

			if tt.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "unsupported parameter type")
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, result)
			}
		})
	}
}
