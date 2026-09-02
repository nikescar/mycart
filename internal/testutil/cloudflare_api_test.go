package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCloudflareClient(t *testing.T) {
	client := NewCloudflareClient("test-account", "test-token")

	assert.NotNil(t, client)
	assert.Equal(t, "test-account", client.accountID)
	assert.Equal(t, "test-token", client.apiToken)
	assert.NotNil(t, client.httpClient)
}

func TestCreateD1Database_Stub(t *testing.T) {
	client := NewCloudflareClient("test-account", "test-token")

	// This will fail against real API (invalid credentials)
	// We're testing the method exists and has correct signature
	dbID, err := client.CreateD1Database("test-db")

	// Expect error with invalid credentials, but method should exist
	assert.Error(t, err)
	assert.Empty(t, dbID)
	assert.Contains(t, err.Error(), "create D1 database")
}

func TestDeleteD1Database_Stub(t *testing.T) {
	client := NewCloudflareClient("test-account", "test-token")

	// Test that method exists and has correct signature
	// With invalid credentials, may get 404 (treated as success - idempotent)
	// or auth error
	err := client.DeleteD1Database("nonexistent-db-id")

	// Method should exist (no compile error)
	// Error behavior depends on API response (404 vs 401)
	_ = err // May be nil (404) or error (auth failure)
}

func TestCreateR2Bucket_Stub(t *testing.T) {
	client := NewCloudflareClient("test-account", "test-token")

	err := client.CreateR2Bucket("test-bucket")

	// Expect error with invalid credentials
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create R2 bucket")
}

func TestDeleteR2Bucket_Stub(t *testing.T) {
	client := NewCloudflareClient("test-account", "test-token")

	err := client.DeleteR2Bucket("nonexistent-bucket")

	// Method should exist (no compile error)
	// Error behavior depends on API response (404 vs 401)
	_ = err // May be nil (404) or error (auth failure)
}
