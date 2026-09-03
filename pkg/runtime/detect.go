package runtime

import (
	"os"
)

// IsCloudflare checks if the app is running in Cloudflare Workers environment
func IsCloudflare() bool {
	// Check for explicit override first
	if runtime := os.Getenv("RUNTIME"); runtime != "" {
		return runtime == "cloudflare"
	}

	// Auto-detect Cloudflare environment
	// Do NOT check test credentials here - they should only be used
	// when RUNTIME=cloudflare is explicitly set
	return os.Getenv("CLOUDFLARE_APPLICATION_ID") != ""
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
	if id := os.Getenv("CLOUDFLARE_TEST_D1_DATABASE_ID"); id != "" {
		return id
	}
	// Production credentials
	return os.Getenv("CLOUDFLARE_D1_DATABASE_ID")
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

// GetR2BucketName returns the Cloudflare R2 bucket name from environment
func GetR2BucketName() string {
	// Check test credentials first
	if name := os.Getenv("CLOUDFLARE_TEST_R2_BUCKET_NAME"); name != "" {
		return name
	}
	// Production credentials
	return os.Getenv("CLOUDFLARE_R2_BUCKET_NAME")
}

// GetR2AccessKeyID returns the Cloudflare R2 access key ID from environment
func GetR2AccessKeyID() string {
	// Check test credentials first
	if key := os.Getenv("CLOUDFLARE_TEST_R2_ACCESS_KEY_ID"); key != "" {
		return key
	}
	// Production credentials
	return os.Getenv("CLOUDFLARE_R2_ACCESS_KEY_ID")
}

// GetR2SecretAccessKey returns the Cloudflare R2 secret access key from environment
func GetR2SecretAccessKey() string {
	// Check test credentials first
	if secret := os.Getenv("CLOUDFLARE_TEST_R2_SECRET_ACCESS_KEY"); secret != "" {
		return secret
	}
	// Production credentials
	return os.Getenv("CLOUDFLARE_R2_SECRET_ACCESS_KEY")
}
