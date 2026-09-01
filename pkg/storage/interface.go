package storage

import (
	"context"
	"io"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Put stores data at the given path
	Put(ctx context.Context, path string, data io.Reader) error

	// Get retrieves data from the given path
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes the file at the given path
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at the given path
	Exists(ctx context.Context, path string) (bool, error)

	// List returns all file paths with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)
}

// Config holds storage configuration
type Config struct {
	Type string // "filesystem" or "r2"

	// Filesystem config
	BasePath string

	// R2 config
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}
