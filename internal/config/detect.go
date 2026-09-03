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
	// Check for Cloudflare environment indicator
	if os.Getenv("CLOUDFLARE_APPLICATION_ID") != "" {
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
			AccountID:  os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
			DatabaseID: os.Getenv("CLOUDFLARE_D1_DATABASE_ID"),
			APIToken:   os.Getenv("CLOUDFLARE_API_TOKEN"),
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
			AccountID:       os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
			AccessKeyID:     os.Getenv("CLOUDFLARE_R2_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("CLOUDFLARE_R2_SECRET_ACCESS_KEY"),
			BucketName:      os.Getenv("CLOUDFLARE_R2_BUCKET_NAME"),
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
