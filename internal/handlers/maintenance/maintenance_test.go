package maintenance

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	app := fiber.New()
	app.Get("/status", GetStatus)

	req := httptest.NewRequest("GET", "/status", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	assert.Contains(t, result, "database")
	assert.Contains(t, result, "storage")
	assert.Contains(t, result, "wrangler")
}

func TestIsWranglerInstalled(t *testing.T) {
	// This test checks the helper function
	// Actual result depends on local environment
	result := isWranglerInstalled()
	// Just verify it returns a boolean without panicking
	assert.IsType(t, false, result)
}

func TestIsMaintenanceMode(t *testing.T) {
	// Test when file doesn't exist
	result := isMaintenanceMode()
	assert.IsType(t, false, result)
}
