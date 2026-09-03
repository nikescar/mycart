package testutil

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// SetupTestDBSmart automatically chooses between SQLite (fast) and D1 (real) based on .env.test
//
// Usage:
//   cleanup := testutil.SetupTestDBSmart(t)
//   defer cleanup()
//
// Behavior:
//   - If .env.test exists and has D1 credentials, uses real D1 database
//   - Otherwise, uses in-memory SQLite (fast, isolated)
//
// .env.test example for D1 testing:
//   CLOUDFLARE_TEST_ACCOUNT_ID=...
//   CLOUDFLARE_TEST_API_TOKEN=...
//   CLOUDFLARE_TEST_R2_ACCESS_KEY_ID=...
//   CLOUDFLARE_TEST_R2_SECRET_ACCESS_KEY=...
func SetupTestDBSmart(t *testing.T) func() {
	t.Helper()

	// Try to load .env.test if it exists (non-fatal)
	_ = godotenv.Load(".env.test")

	// Check if D1 credentials are configured
	if isD1Configured() {
		t.Log("Using D1 database for testing (configured in .env.test)")
		return SetupCloudflareTestEnv()
	}

	// Default to fast in-memory SQLite
	t.Log("Using in-memory SQLite for testing (fast mode)")
	return SetupTestDB(t)
}

// isD1Configured checks if all required D1 credentials are present
func isD1Configured() bool {
	required := []string{
		"CLOUDFLARE_TEST_ACCOUNT_ID",
		"CLOUDFLARE_TEST_API_TOKEN",
		"CLOUDFLARE_TEST_R2_ACCESS_KEY_ID",
		"CLOUDFLARE_TEST_R2_SECRET_ACCESS_KEY",
	}

	for _, key := range required {
		if os.Getenv(key) == "" {
			return false
		}
	}
	return true
}
