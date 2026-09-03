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
			name: "CLOUDFLARE_APPLICATION_ID set",
			envVars: map[string]string{
				"CLOUDFLARE_APPLICATION_ID": "true",
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
				"CLOUDFLARE_APPLICATION_ID":      "true",
				"CLOUDFLARE_ACCOUNT_ID":  "test-account",
				"CLOUDFLARE_D1_DATABASE_ID": "test-db-id",
				"CLOUDFLARE_API_TOKEN":   "test-token",
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
				"CLOUDFLARE_APPLICATION_ID":            "true",
				"CLOUDFLARE_ACCOUNT_ID":        "test-account",
				"CLOUDFLARE_R2_ACCESS_KEY_ID":     "test-key",
				"CLOUDFLARE_R2_SECRET_ACCESS_KEY": "test-secret",
				"CLOUDFLARE_R2_BUCKET_NAME":       "test-bucket",
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
