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
	// Auto-detect storage type if not specified
	if config.Type == "" {
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
		return NewR2(
			config.AccountID,
			config.AccessKeyID,
			config.SecretAccessKey,
			config.BucketName,
		)

	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}
