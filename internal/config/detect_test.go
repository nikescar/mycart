package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsCloudflare(t *testing.T) {

	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no cloudflare env vars",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name: "CF_WORKER set",
			envVars: map[string]string{
				"CF_WORKER": "true",
			},
			expected: true,
		},
		{
			name: "CF_PAGES set",
			envVars: map[string]string{
				"CF_PAGES": "1",
			},
			expected: true,
		},
		{
			name: "D1_DATABASE set",
			envVars: map[string]string{
				"D1_DATABASE": "my-db",
			},
			expected: true,
		},
		{
			name: "multiple cloudflare vars",
			envVars: map[string]string{
				"CF_WORKER":   "true",
				"D1_DATABASE": "my-db",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			result := IsCloudflare()
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadDatabaseConfig(t *testing.T) {

	tests := []struct {
		name         string
		envVars      map[string]string
		expectedType string
		checkFields  func(t *testing.T, cfg DatabaseConfig)
	}{
		{
			name:         "local mode - defaults to sqlite",
			envVars:      map[string]string{},
			expectedType: "sqlite",
			checkFields: func(t *testing.T, cfg DatabaseConfig) {
				require.Equal(t, "sqlite", cfg.Type)
				require.NotEmpty(t, cfg.Path)
			},
		},
		{
			name: "local mode - custom sqlite path",
			envVars: map[string]string{
				"DB_PATH": "/custom/path/db.sqlite",
			},
			expectedType: "sqlite",
			checkFields: func(t *testing.T, cfg DatabaseConfig) {
				require.Equal(t, "sqlite", cfg.Type)
				require.Equal(t, "/custom/path/db.sqlite", cfg.Path)
			},
		},
		{
			name: "cloudflare mode - d1 config",
			envVars: map[string]string{
				"CF_WORKER":      "true",
				"D1_ACCOUNT_ID":  "test-account",
				"D1_DATABASE_ID": "test-db-id",
				"D1_API_TOKEN":   "test-token",
			},
			expectedType: "d1",
			checkFields: func(t *testing.T, cfg DatabaseConfig) {
				require.Equal(t, "d1", cfg.Type)
				require.Equal(t, "test-account", cfg.AccountID)
				require.Equal(t, "test-db-id", cfg.DatabaseID)
				require.Equal(t, "test-token", cfg.APIToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg := LoadDatabaseConfig()
			require.Equal(t, tt.expectedType, cfg.Type)
			if tt.checkFields != nil {
				tt.checkFields(t, cfg)
			}
		})
	}
}

func TestLoadStorageConfig(t *testing.T) {

	tests := []struct {
		name         string
		envVars      map[string]string
		expectedType string
		checkFields  func(t *testing.T, cfg StorageConfig)
	}{
		{
			name:         "local mode - defaults to filesystem",
			envVars:      map[string]string{},
			expectedType: "filesystem",
			checkFields: func(t *testing.T, cfg StorageConfig) {
				require.Equal(t, "filesystem", cfg.Type)
				require.NotEmpty(t, cfg.BasePath)
			},
		},
		{
			name: "local mode - custom base path",
			envVars: map[string]string{
				"STORAGE_PATH": "/custom/uploads",
			},
			expectedType: "filesystem",
			checkFields: func(t *testing.T, cfg StorageConfig) {
				require.Equal(t, "filesystem", cfg.Type)
				require.Equal(t, "/custom/uploads", cfg.BasePath)
			},
		},
		{
			name: "cloudflare mode - r2 config",
			envVars: map[string]string{
				"CF_WORKER":            "true",
				"R2_ACCOUNT_ID":        "test-account",
				"R2_ACCESS_KEY_ID":     "test-key",
				"R2_SECRET_ACCESS_KEY": "test-secret",
				"R2_BUCKET_NAME":       "test-bucket",
			},
			expectedType: "r2",
			checkFields: func(t *testing.T, cfg StorageConfig) {
				require.Equal(t, "r2", cfg.Type)
				require.Equal(t, "test-account", cfg.AccountID)
				require.Equal(t, "test-key", cfg.AccessKeyID)
				require.Equal(t, "test-secret", cfg.SecretAccessKey)
				require.Equal(t, "test-bucket", cfg.BucketName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg := LoadStorageConfig()
			require.Equal(t, tt.expectedType, cfg.Type)
			if tt.checkFields != nil {
				tt.checkFields(t, cfg)
			}
		})
	}
}
