package testutil

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shurco/mycart/internal/queries"
	"github.com/shurco/mycart/internal/store"
	"github.com/shurco/mycart/pkg/database"
	"github.com/shurco/mycart/pkg/storage"
)

// SetupCloudflareTestEnv creates a real D1 database and R2 bucket for testing.
// Returns cleanup function that deletes both resources.
func SetupCloudflareTestEnv() func() {
	// 1. Read credentials from environment
	accountID := os.Getenv("CLOUDFLARE_TEST_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_TEST_API_TOKEN")
	r2AccessKey := os.Getenv("CLOUDFLARE_TEST_R2_ACCESS_KEY_ID")
	r2SecretKey := os.Getenv("CLOUDFLARE_TEST_R2_SECRET_ACCESS_KEY")

	if accountID == "" || apiToken == "" {
		panic("CLOUDFLARE_TEST_ACCOUNT_ID and CLOUDFLARE_TEST_API_TOKEN required")
	}
	if r2AccessKey == "" || r2SecretKey == "" {
		panic("CLOUDFLARE_TEST_R2_ACCESS_KEY_ID and CLOUDFLARE_TEST_R2_SECRET_ACCESS_KEY required")
	}

	// 2. Create unique resource names
	timestamp := time.Now().UnixNano()
	dbName := fmt.Sprintf("mycart-test-db-%d", timestamp)
	bucketName := fmt.Sprintf("mycart-test-bucket-%d", timestamp)

	// 3. Create Cloudflare API client
	client := NewCloudflareClient(accountID, apiToken)

	// 4. Create D1 database
	fmt.Printf("Creating D1 database: %s\n", dbName)
	dbID, err := client.CreateD1Database(dbName)
	if err != nil {
		panic(fmt.Sprintf("Failed to create D1 database: %v", err))
	}
	fmt.Printf("Created D1 database: %s (ID: %s)\n", dbName, dbID)

	// 5. Create R2 bucket
	fmt.Printf("Creating R2 bucket: %s\n", bucketName)
	if err := client.CreateR2Bucket(bucketName); err != nil {
		_ = client.DeleteD1Database(dbID)
		panic(fmt.Sprintf("Failed to create R2 bucket: %v", err))
	}
	fmt.Printf("Created R2 bucket: %s\n", bucketName)

	// 6. Connect to D1 database
	d1DB, err := database.NewD1(accountID, dbID, apiToken)
	if err != nil {
		cleanup(client, dbID, bucketName)
		panic(fmt.Sprintf("Failed to connect to D1: %v", err))
	}

	// 7. Run migrations on D1
	fmt.Println("Running migrations...")
	if err := runMigrationsOnD1(client, dbID); err != nil {
		cleanup(client, dbID, bucketName)
		panic(fmt.Sprintf("Failed to run migrations: %v", err))
	}
	fmt.Println("Migrations complete ✓")

	// 8. Run fixtures on D1
	fmt.Println("Running fixtures...")
	if err := runFixturesOnD1(client, dbID); err != nil {
		cleanup(client, dbID, bucketName)
		panic(fmt.Sprintf("Failed to run fixtures: %v", err))
	}
	fmt.Println("Fixtures complete ✓")

	// 9. Set global DB to D1
	queries.NewFromDB(d1DB.DB())

	// 10. Create and set R2 storage
	r2Storage, err := storage.NewR2(accountID, r2AccessKey, r2SecretKey, bucketName)
	if err != nil {
		cleanup(client, dbID, bucketName)
		panic(fmt.Sprintf("Failed to create R2 storage: %v", err))
	}
	store.New(r2Storage)

	// 11. Setup cleanup function
	cleanupFunc := func() {
		fmt.Printf("Cleaning up D1 database: %s\n", dbID)
		_ = client.DeleteD1Database(dbID)

		fmt.Printf("Cleaning up R2 bucket: %s\n", bucketName)
		_ = client.DeleteR2Bucket(bucketName)
	}

	// 12. Setup cleanup on Ctrl+C
	setupCleanupOnInterrupt(cleanupFunc)

	return cleanupFunc
}

// cleanup helper for error cases
func cleanup(client *CloudflareClient, dbID, bucketName string) {
	_ = client.DeleteD1Database(dbID)
	_ = client.DeleteR2Bucket(bucketName)
}

// setupCleanupOnInterrupt registers signal handler for Ctrl+C
func setupCleanupOnInterrupt(cleanup func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nInterrupted! Cleaning up...")
		cleanup()
		os.Exit(1)
	}()
}
