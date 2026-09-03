package cloudflare

import (
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/mycart/pkg/webutil"
)

// ListD1Databases returns available D1 databases
func ListD1Databases(c fiber.Ctx) error {
	// Accept credentials from query params (for installer) or env (for post-install)
	accountID := c.Query("account_id")
	apiToken := c.Query("api_token")

	// Check if request credentials are incomplete
	hasRequestCreds := accountID != "" || apiToken != ""
	if hasRequestCreds {
		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "incomplete Cloudflare credentials in request (both account_id and api_token required)")
		}
	} else {
		// Fall back to environment variables
		accountID = os.Getenv("CF_ACCOUNT_ID")
		apiToken = os.Getenv("CLOUDFLARE_API_TOKEN")

		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CLOUDFLARE_API_TOKEN environment variables required")
		}
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
		Name      string `json:"name"`
		AccountID string `json:"account_id"`
		APIToken  string `json:"api_token"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	// Use credentials from request body or fall back to env
	accountID := req.AccountID
	apiToken := req.APIToken

	// Check if request credentials are incomplete
	hasRequestCreds := accountID != "" || apiToken != ""
	if hasRequestCreds {
		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "incomplete Cloudflare credentials in request (both account_id and api_token required)")
		}
	} else {
		// Fall back to environment variables
		accountID = os.Getenv("CF_ACCOUNT_ID")
		apiToken = os.Getenv("CLOUDFLARE_API_TOKEN")

		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CLOUDFLARE_API_TOKEN environment variables required")
		}
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
	// Accept credentials from query params (for installer) or env (for post-install)
	accountID := c.Query("account_id")
	apiToken := c.Query("api_token")

	// Check if request credentials are incomplete
	hasRequestCreds := accountID != "" || apiToken != ""
	if hasRequestCreds {
		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "incomplete Cloudflare credentials in request (both account_id and api_token required)")
		}
	} else {
		// Fall back to environment variables
		accountID = os.Getenv("CF_ACCOUNT_ID")
		apiToken = os.Getenv("CLOUDFLARE_API_TOKEN")

		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CLOUDFLARE_API_TOKEN environment variables required")
		}
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
		Name      string `json:"name"`
		AccountID string `json:"account_id"`
		APIToken  string `json:"api_token"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return webutil.StatusBadRequest(c, err.Error())
	}

	// Use credentials from request body or fall back to env
	accountID := req.AccountID
	apiToken := req.APIToken

	// Check if request credentials are incomplete
	hasRequestCreds := accountID != "" || apiToken != ""
	if hasRequestCreds {
		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "incomplete Cloudflare credentials in request (both account_id and api_token required)")
		}
	} else {
		// Fall back to environment variables
		accountID = os.Getenv("CF_ACCOUNT_ID")
		apiToken = os.Getenv("CLOUDFLARE_API_TOKEN")

		if accountID == "" || apiToken == "" {
			return webutil.StatusBadRequest(c, "CF_ACCOUNT_ID and CLOUDFLARE_API_TOKEN environment variables required")
		}
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
