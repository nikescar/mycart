package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type r2Storage struct {
	client     *s3.Client
	bucketName string
}

// NewR2 creates a new Cloudflare R2 storage instance
func NewR2(accountID, accessKeyID, secretAccessKey, bucketName string) (Storage, error) {
	if accountID == "" {
		return nil, fmt.Errorf("account ID is required")
	}
	if accessKeyID == "" {
		return nil, fmt.Errorf("access key ID is required")
	}
	if secretAccessKey == "" {
		return nil, fmt.Errorf("secret access key is required")
	}
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	// R2 endpoint format: https://<accountID>.r2.cloudflarestorage.com
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	// Create S3 client configured for R2
	client := s3.New(s3.Options{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		),
		BaseEndpoint: aws.String(endpoint),
	})

	return &r2Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// Put stores data at the given path (stub - S3 API integration deferred)
func (r *r2Storage) Put(ctx context.Context, path string, data io.Reader) error {
	return fmt.Errorf("R2 Put not yet implemented")
}

// Get retrieves data from the given path (stub - S3 API integration deferred)
func (r *r2Storage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("R2 Get not yet implemented")
}

// Delete removes the file at the given path (stub - S3 API integration deferred)
func (r *r2Storage) Delete(ctx context.Context, path string) error {
	return fmt.Errorf("R2 Delete not yet implemented")
}

// Exists checks if a file exists at the given path (stub - S3 API integration deferred)
func (r *r2Storage) Exists(ctx context.Context, path string) (bool, error) {
	return false, fmt.Errorf("R2 Exists not yet implemented")
}

// List returns all file paths with the given prefix (stub - S3 API integration deferred)
func (r *r2Storage) List(ctx context.Context, prefix string) ([]string, error) {
	return nil, fmt.Errorf("R2 List not yet implemented")
}
