package config

import (
	"os"

	"github.com/shurco/mycart/pkg/database"
	"github.com/shurco/mycart/pkg/storage"
)

// Type aliases for cleaner API
type (
	DatabaseConfig = database.Config
	StorageConfig  = storage.Config
)

// IsCloudflare detects if running in Cloudflare environment
// by checking for Cloudflare-specific environment variables
func IsCloudflare() bool {
	// Check for common Cloudflare Worker/Pages environment variables
	if os.Getenv("CF_WORKER") != "" {
		return true
	}
	if os.Getenv("CF_PAGES") != "" {
		return true
	}
	if os.Getenv("D1_DATABASE") != "" {
		return true
	}
	return false
}

// LoadDatabaseConfig loads database configuration based on environment
// Local: SQLite
// Cloudflare: D1
func LoadDatabaseConfig() DatabaseConfig {
	if IsCloudflare() {
		// Cloudflare D1 configuration
		return DatabaseConfig{
			Type:       "d1",
			AccountID:  os.Getenv("D1_ACCOUNT_ID"),
			DatabaseID: os.Getenv("D1_DATABASE_ID"),
			APIToken:   os.Getenv("D1_API_TOKEN"),
		}
	}

	// Local SQLite configuration
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "./lc_base/data.db"
	}

	return DatabaseConfig{
		Type: "sqlite",
		Path: path,
	}
}

// LoadStorageConfig loads storage configuration based on environment
// Local: Filesystem
// Cloudflare: R2
func LoadStorageConfig() StorageConfig {
	if IsCloudflare() {
		// Cloudflare R2 configuration
		return StorageConfig{
			Type:            "r2",
			AccountID:       os.Getenv("R2_ACCOUNT_ID"),
			AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
			BucketName:      os.Getenv("R2_BUCKET_NAME"),
		}
	}

	// Local filesystem configuration
	basePath := os.Getenv("STORAGE_PATH")
	if basePath == "" {
		basePath = "./lc_uploads"
	}

	return StorageConfig{
		Type:     "filesystem",
		BasePath: basePath,
	}
}
