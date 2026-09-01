package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestR2_New(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		accountID       string
		accessKeyID     string
		secretAccessKey string
		bucketName      string
		wantErr         bool
		errContains     string
	}{
		{
			name:            "valid configuration",
			accountID:       "test-account",
			accessKeyID:     "test-access-key",
			secretAccessKey: "test-secret-key",
			bucketName:      "test-bucket",
			wantErr:         false,
		},
		{
			name:            "missing account ID",
			accountID:       "",
			accessKeyID:     "test-access-key",
			secretAccessKey: "test-secret-key",
			bucketName:      "test-bucket",
			wantErr:         true,
			errContains:     "account ID is required",
		},
		{
			name:            "missing access key ID",
			accountID:       "test-account",
			accessKeyID:     "",
			secretAccessKey: "test-secret-key",
			bucketName:      "test-bucket",
			wantErr:         true,
			errContains:     "access key ID is required",
		},
		{
			name:            "missing secret access key",
			accountID:       "test-account",
			accessKeyID:     "test-access-key",
			secretAccessKey: "",
			bucketName:      "test-bucket",
			wantErr:         true,
			errContains:     "secret access key is required",
		},
		{
			name:            "missing bucket name",
			accountID:       "test-account",
			accessKeyID:     "test-access-key",
			secretAccessKey: "test-secret-key",
			bucketName:      "",
			wantErr:         true,
			errContains:     "bucket name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r2, err := NewR2(tt.accountID, tt.accessKeyID, tt.secretAccessKey, tt.bucketName)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				require.Nil(t, r2)
			} else {
				require.NoError(t, err)
				require.NotNil(t, r2)
			}
		})
	}
}

func TestR2_Operations_Stub(t *testing.T) {
	t.Parallel()

	r2, err := NewR2("test-account", "test-key", "test-secret", "test-bucket")
	require.NoError(t, err)

	ctx := context.Background()

	// These are stubs - actual S3 API calls would require real credentials
	// Testing that methods exist and return expected stub errors

	// Put should return not implemented error
	err = r2.Put(ctx, "test.txt", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")

	// Get should return not implemented error
	_, err = r2.Get(ctx, "test.txt")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")

	// Delete should return not implemented error
	err = r2.Delete(ctx, "test.txt")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")

	// Exists should return not implemented error
	_, err = r2.Exists(ctx, "test.txt")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")

	// List should return not implemented error
	_, err = r2.List(ctx, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not yet implemented")
}
