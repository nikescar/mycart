package testutil

import (
	"os"
	"testing"
)

// SetupTestEnv sets up test environment based on USE_CLOUDFLARE_TESTS env var.
// Returns cleanup function.
//
// Default (USE_CLOUDFLARE_TESTS not set): in-memory SQLite (fast)
// USE_CLOUDFLARE_TESTS=1: real Cloudflare D1 + R2 (production-accurate)
func SetupTestEnv(t *testing.T) func() {
	t.Helper()

	if os.Getenv("USE_CLOUDFLARE_TESTS") == "1" {
		return SetupCloudflareTestEnv()
	}

	// Default: in-memory SQLite (fast local tests)
	return SetupTestDB(t)
}
