package storage

import "fmt"

// New creates a storage instance based on the provided configuration.
// If Type is empty, defaults to "filesystem".
// If Type is "filesystem" and BasePath is empty, defaults to "./uploads".
func New(config Config) (Storage, error) {
	// Default to filesystem if no type specified
	if config.Type == "" {
		config.Type = "filesystem"
	}

	switch config.Type {
	case "filesystem":
		basePath := config.BasePath
		if basePath == "" {
			basePath = "./uploads"
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
