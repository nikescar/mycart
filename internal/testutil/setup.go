package testutil

import "os"

// SetupTestEnv sets up test environment based on USE_CLOUDFLARE_TESTS env var.
// Returns cleanup function.
//
// Default (USE_CLOUDFLARE_TESTS not set): in-memory SQLite (fast)
// USE_CLOUDFLARE_TESTS=1: real Cloudflare D1 + R2 (production-accurate)
func SetupTestEnv() func() {
	if os.Getenv("USE_CLOUDFLARE_TESTS") == "1" {
		return SetupCloudflareTestEnv()
	}

	// Default: in-memory SQLite (use existing testdb.go setup)
	// Note: This is a simplified version. Full implementation would call SetupTestDB
	// For now, return a no-op cleanup
	return func() {}
}
