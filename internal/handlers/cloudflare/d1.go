package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/d1"
	"github.com/cloudflare/cloudflare-go/v7/option"
)

type D1Client struct {
	client    *cloudflare.Client
	accountID string
}

type D1Database struct {
	ID        string    `json:"uuid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewD1Client(accountID, apiToken string) *D1Client {
	client := cloudflare.NewClient(
		option.WithAPIToken(apiToken),
	)

	return &D1Client{
		client:    client,
		accountID: accountID,
	}
}

func (c *D1Client) ListDatabases() ([]D1Database, error) {
	ctx := context.Background()

	// List all D1 databases for the account
	params := d1.DatabaseListParams{
		AccountID: cloudflare.F(c.accountID),
	}

	response, err := c.client.D1.Database.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list D1 databases: %w", err)
	}

	// Convert to our D1Database type
	result := make([]D1Database, 0, len(response.Result))
	for _, db := range response.Result {
		result = append(result, D1Database{
			ID:        db.UUID,
			Name:      db.Name,
			CreatedAt: db.CreatedAt,
		})
	}

	return result, nil
}

func (c *D1Client) CreateDatabase(name string) (*D1Database, error) {
	ctx := context.Background()

	// Create new D1 database
	params := d1.DatabaseNewParams{
		AccountID: cloudflare.F(c.accountID),
		Name:      cloudflare.F(name),
	}

	db, err := c.client.D1.Database.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create D1 database: %w", err)
	}

	return &D1Database{
		ID:        db.UUID,
		Name:      db.Name,
		CreatedAt: db.CreatedAt,
	}, nil
}

func (c *D1Client) ExportDatabase(databaseID, outputPath string) error {
	// Note: The cloudflare-go SDK may not have direct export support yet
	// This would need to be implemented if the SDK adds it, or keep using HTTP API
	return fmt.Errorf("export not yet implemented with cloudflare-go SDK")
}

func (c *D1Client) ImportSQLite(databaseID, sqlitePath string) error {
	// Note: The cloudflare-go SDK may not have direct import support yet
	// This would need to be implemented if the SDK adds it, or keep using HTTP API
	return fmt.Errorf("import not yet implemented with cloudflare-go SDK")
}
