package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/cloudflare/cloudflare-go/v7/r2"
)

type R2Client struct {
	client    *cloudflare.Client
	accountID string
}

type R2Bucket struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"creation_date"`
}

type R2Object struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

func NewR2Client(accountID, apiToken string) *R2Client {
	client := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	return &R2Client{
		client:    client,
		accountID: accountID,
	}
}

func (c *R2Client) ListBuckets() ([]R2Bucket, error) {
	ctx := context.Background()

	// List all R2 buckets for the account
	params := r2.BucketListParams{
		AccountID: cloudflare.F(c.accountID),
	}

	buckets, err := c.client.R2.Buckets.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list R2 buckets: %w", err)
	}

	// Convert to our R2Bucket type
	result := make([]R2Bucket, 0, len(buckets.Buckets))
	for _, bucket := range buckets.Buckets {
		createdAt, _ := time.Parse(time.RFC3339, bucket.CreationDate)
		result = append(result, R2Bucket{
			Name:      bucket.Name,
			CreatedAt: createdAt,
		})
	}

	return result, nil
}

func (c *R2Client) CreateBucket(name string) (*R2Bucket, error) {
	ctx := context.Background()

	// Create new R2 bucket
	params := r2.BucketNewParams{
		AccountID: cloudflare.F(c.accountID),
		Name:      cloudflare.F(name),
	}

	bucket, err := c.client.R2.Buckets.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 bucket: %w", err)
	}

	createdAt, _ := time.Parse(time.RFC3339, bucket.CreationDate)

	return &R2Bucket{
		Name:      bucket.Name,
		CreatedAt: createdAt,
	}, nil
}

func (c *R2Client) ListObjects(bucketName string) ([]R2Object, error) {
	// Note: The cloudflare-go SDK may not have direct object listing support yet
	// This would need to be implemented if the SDK adds it, or keep using HTTP API
	return nil, fmt.Errorf("list objects not yet implemented with cloudflare-go SDK")
}
