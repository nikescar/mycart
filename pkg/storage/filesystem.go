package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type filesystemStorage struct {
	basePath string
}

// NewFilesystem creates a new filesystem-based storage
func NewFilesystem(basePath string) (Storage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	return &filesystemStorage{
		basePath: basePath,
	}, nil
}

// Put stores data at the given path
func (fs *filesystemStorage) Put(ctx context.Context, path string, data io.Reader) error {
	fullPath := filepath.Join(fs.basePath, path)

	// Create parent directories if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	if _, err := io.Copy(file, data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Get retrieves data from the given path
func (fs *filesystemStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(fs.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes the file at the given path
func (fs *filesystemStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(fs.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists at the given path
func (fs *filesystemStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(fs.basePath, path)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// List returns all file paths with the given prefix
func (fs *filesystemStorage) List(ctx context.Context, prefix string) ([]string, error) {
	var files []string

	err := filepath.Walk(fs.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(fs.basePath, path)
		if err != nil {
			return err
		}

		// Normalize path separators to forward slashes
		relPath = filepath.ToSlash(relPath)

		// Filter by prefix
		if prefix == "" || strings.HasPrefix(relPath, prefix) {
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}
