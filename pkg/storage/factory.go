package storage

import (
	"fmt"
	"os"

	"github.com/shurco/mycart/pkg/runtime"
)

// New creates a storage instance based on the provided configuration.
// If Type is empty, auto-detects based on runtime environment.
// If Type is "filesystem" and BasePath is empty, uses runtime default.
func New(config Config) (Storage, error) {
	// Track if type was auto-detected to enable full auto-configuration
	autoDetected := false

	// Auto-detect storage type if not specified
	if config.Type == "" {
		autoDetected = true
		if runtime.IsCloudflare() {
			config.Type = "r2"
		} else {
			config.Type = "filesystem"
		}
	}

	switch config.Type {
	case "filesystem":
		basePath := config.BasePath
		if basePath == "" {
			basePath = os.Getenv("STORAGE_BASE_PATH")
			if basePath == "" {
				basePath = "./lc_base/uploads"
			}
		}
		return NewFilesystem(basePath)

	case "r2":
		accountID := config.AccountID
		accessKeyID := config.AccessKeyID
		secretAccessKey := config.SecretAccessKey
		bucketName := config.BucketName

		// Only auto-populate credentials if type was auto-detected
		if autoDetected {
			if accountID == "" {
				accountID = runtime.GetD1AccountID()
			}
			if accessKeyID == "" {
				accessKeyID = runtime.GetR2AccessKeyID()
			}
			if secretAccessKey == "" {
				secretAccessKey = runtime.GetR2SecretAccessKey()
			}
			if bucketName == "" {
				bucketName = runtime.GetR2BucketName()
			}
		}

		return NewR2(
			accountID,
			accessKeyID,
			secretAccessKey,
			bucketName,
		)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}
