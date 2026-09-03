package runtime

import (
	"os"
	"path/filepath"
)

// IsCloudflare checks if the app is running in Cloudflare Workers environment
func IsCloudflare() bool {
	// Check for explicit override first
	if runtime := os.Getenv("RUNTIME"); runtime == "cloudflare" {
		return true
	}

	// Check for test environment with Cloudflare credentials
	if os.Getenv("CLOUDFLARE_TEST_ACCOUNT_ID") != "" {
		return true
	}

	// Auto-detect Cloudflare environment
	return os.Getenv("CLOUDFLARE_APPLICATION_ID") != ""
}

// isWritable is a variable to allow mocking in tests
var isWritable = func(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	// Try to create a temp file to verify write access
	testFile := filepath.Join(path, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)

	return true
}

// GetMaintenanceFlagPath returns the appropriate maintenance flag path
// based on the runtime environment
func GetMaintenanceFlagPath() string {
	// Try /app first (Docker/Cloudflare)
	if isWritable("/app") {
		return "/app/maintenance.flag"
	}

	// Fallback to local project root
	return "./maintenance.flag"
}

// GetDatabasePath returns the appropriate database path based on runtime
func GetDatabasePath() string {
	if IsCloudflare() {
		// D1 doesn't use file paths, return empty
		return ""
	}

	// Local SQLite path
	path := os.Getenv("DB_PATH")
	if path == "" {
		return "./lc_base/mycart.db"
	}
	return path
}

// GetD1AccountID returns the Cloudflare account ID from environment
func GetD1AccountID() string {
	// Check test credentials first
	if id := os.Getenv("CLOUDFLARE_TEST_ACCOUNT_ID"); id != "" {
		return id
	}
	// Production credentials
	return os.Getenv("CLOUDFLARE_ACCOUNT_ID")
}

// GetD1DatabaseID returns the Cloudflare D1 database ID from environment
func GetD1DatabaseID() string {
	// Check test credentials first
	if id := os.Getenv("CLOUDFLARE_TEST_DATABASE_ID"); id != "" {
		return id
	}
	// Production credentials
	return os.Getenv("CLOUDFLARE_DATABASE_ID")
}

// GetD1APIToken returns the Cloudflare API token from environment
func GetD1APIToken() string {
	// Check test credentials first
	if token := os.Getenv("CLOUDFLARE_TEST_API_TOKEN"); token != "" {
		return token
	}
	// Production credentials
	return os.Getenv("CLOUDFLARE_API_TOKEN")
}
