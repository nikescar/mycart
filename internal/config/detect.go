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
// Priority: DB_TYPE env var > IsCloudflare() detection > default SQLite
func LoadDatabaseConfig() DatabaseConfig {
	// 1. Check DB_TYPE env var first (explicit configuration)
	dbType := os.Getenv("DB_TYPE")

	// 2. Fall back to IsCloudflare() detection if not set
	if dbType == "" && IsCloudflare() {
		dbType = "d1"
	}

	// 3. Default to sqlite if still not set
	if dbType == "" {
		dbType = "sqlite"
	}

	// 4. Return appropriate config based on type
	if dbType == "d1" {
		return DatabaseConfig{
			Type:       "d1",
			AccountID:  os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
			DatabaseID: os.Getenv("CLOUDFLARE_D1_DATABASE_ID"),
			APIToken:   os.Getenv("CLOUDFLARE_API_TOKEN"),
		}
	}

	// Default SQLite configuration
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
// Priority: STORAGE_TYPE env var > IsCloudflare() detection > default filesystem
func LoadStorageConfig() StorageConfig {
	// 1. Check STORAGE_TYPE env var first (explicit configuration)
	storageType := os.Getenv("STORAGE_TYPE")

	// 2. Fall back to IsCloudflare() detection if not set
	if storageType == "" && IsCloudflare() {
		storageType = "r2"
	}

	// 3. Default to filesystem if still not set
	if storageType == "" {
		storageType = "filesystem"
	}

	// 4. Return appropriate config based on type
	if storageType == "r2" {
		return StorageConfig{
			Type:            "r2",
			AccountID:       os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
			AccessKeyID:     os.Getenv("CLOUDFLARE_R2_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("CLOUDFLARE_R2_SECRET_ACCESS_KEY"),
			BucketName:      os.Getenv("CLOUDFLARE_R2_BUCKET_NAME"),
		}
	}

	// Default filesystem configuration
	basePath := os.Getenv("STORAGE_PATH")
	if basePath == "" {
		basePath = "./lc_uploads"
	}

	return StorageConfig{
		Type:     "filesystem",
		BasePath: basePath,
	}
}
