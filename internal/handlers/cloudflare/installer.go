package cloudflare

import (
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/mycart/pkg/webutil"
)

// ListD1Databases returns available D1 databases
func ListD1Databases(c fiber.Ctx) error {
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")

	if accountID == "" || apiToken == "" {
		return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CF_API_TOKEN required")
	}

	client := NewD1Client(accountID, apiToken)
	databases, err := client.ListDatabases()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return webutil.Response(c, fiber.StatusOK, "D1 databases retrieved", databases)
}

// CreateD1Database creates a new D1 database
func CreateD1Database(c fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")

	if accountID == "" || apiToken == "" {
		return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CF_API_TOKEN required")
	}

	client := NewD1Client(accountID, apiToken)
	database, err := client.CreateDatabase(req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return webutil.Response(c, fiber.StatusOK, "D1 database created", database)
}

// ListR2Buckets returns available R2 buckets
func ListR2Buckets(c fiber.Ctx) error {
	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")

	if accountID == "" || apiToken == "" {
		return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CF_API_TOKEN required")
	}

	client := NewR2Client(accountID, apiToken)
	buckets, err := client.ListBuckets()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return webutil.Response(c, fiber.StatusOK, "R2 buckets retrieved", buckets)
}

// CreateR2Bucket creates a new R2 bucket
func CreateR2Bucket(c fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	accountID := os.Getenv("CF_ACCOUNT_ID")
	apiToken := os.Getenv("CF_API_TOKEN")

	if accountID == "" || apiToken == "" {
		return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CF_API_TOKEN required")
	}

	client := NewR2Client(accountID, apiToken)
	bucket, err := client.CreateBucket(req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return webutil.Response(c, fiber.StatusOK, "R2 bucket created", bucket)
}
