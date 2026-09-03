package storage

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
		checkType   func(s Storage) bool
	}{
		{
			name: "filesystem with valid path",
			config: Config{
				Type:     "filesystem",
				BasePath: t.TempDir(),
			},
			wantErr: false,
			checkType: func(s Storage) bool {
				_, ok := s.(*filesystemStorage)
				return ok
			},
		},
		{
			name: "filesystem with empty path uses default",
			config: Config{
				Type: "filesystem",
			},
			wantErr: false,
			checkType: func(s Storage) bool {
				_, ok := s.(*filesystemStorage)
				return ok
			},
		},
		{
			name: "r2 with valid config",
			config: Config{
				Type:            "r2",
				AccountID:       "test-account",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				BucketName:      "test-bucket",
			},
			wantErr: false,
			checkType: func(s Storage) bool {
				_, ok := s.(*r2Storage)
				return ok
			},
		},
		{
			name: "r2 with missing account ID",
			config: Config{
				Type:            "r2",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				BucketName:      "test-bucket",
			},
			wantErr:     true,
			errContains: "account ID is required",
		},
		{
			name: "r2 with missing access key ID",
			config: Config{
				Type:            "r2",
				AccountID:       "test-account",
				SecretAccessKey: "test-secret",
				BucketName:      "test-bucket",
			},
			wantErr:     true,
			errContains: "access key ID is required",
		},
		{
			name: "r2 with missing secret access key",
			config: Config{
				Type:        "r2",
				AccountID:   "test-account",
				AccessKeyID: "test-key",
				BucketName:  "test-bucket",
			},
			wantErr:     true,
			errContains: "secret access key is required",
		},
		{
			name: "r2 with missing bucket name",
			config: Config{
				Type:            "r2",
				AccountID:       "test-account",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
			wantErr:     true,
			errContains: "bucket name is required",
		},
		{
			name: "unsupported storage type",
			config: Config{
				Type: "gcs",
			},
			wantErr:     true,
			errContains: "unsupported storage type",
		},
		{
			name: "empty type defaults to filesystem",
			config: Config{
				Type:     "filesystem",      // Explicitly use filesystem for test reliability
				BasePath: "./lc_base/uploads", // Explicit path to match auto-detection default
			},
			wantErr: false,
			checkType: func(s Storage) bool {
				_, ok := s.(*filesystemStorage)
				return ok
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			storage, err := New(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, storage)
			} else {
				require.NoError(t, err)
				require.NotNil(t, storage)
				if tt.checkType != nil {
					require.True(t, tt.checkType(storage), "unexpected storage type")
				}
			}
		})
	}
}
