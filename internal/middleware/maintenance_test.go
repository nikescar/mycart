package middleware

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMaintenanceMode(t *testing.T) {
	// Cleanup first
	os.Remove("maintenance.flag")

	// Test: no file = not in maintenance mode
	assert.False(t, isMaintenanceMode())

	// Test: file exists = in maintenance mode
	f, err := os.Create("maintenance.flag")
	assert.NoError(t, err)
	f.Close()

	assert.True(t, isMaintenanceMode())

	// Cleanup
	os.Remove("maintenance.flag")
}

func TestLoadAllowedIPs_Default(t *testing.T) {
	os.Unsetenv("MAINTENANCE_ALLOWED_IPS")

	ips := loadAllowedIPs()

	assert.Equal(t, []string{"127.0.0.1", "::1"}, ips)
}

func TestLoadAllowedIPs_Custom(t *testing.T) {
	os.Setenv("MAINTENANCE_ALLOWED_IPS", "127.0.0.1,192.168.1.100,10.0.0.1")
	defer os.Unsetenv("MAINTENANCE_ALLOWED_IPS")

	ips := loadAllowedIPs()

	assert.Equal(t, []string{"127.0.0.1", "192.168.1.100", "10.0.0.1"}, ips)
}

func TestLoadAllowedIPs_WithWhitespace(t *testing.T) {
	os.Setenv("MAINTENANCE_ALLOWED_IPS", " 127.0.0.1 , 192.168.1.100 , 10.0.0.1 ")
	defer os.Unsetenv("MAINTENANCE_ALLOWED_IPS")

	ips := loadAllowedIPs()

	assert.Equal(t, []string{"127.0.0.1", "192.168.1.100", "10.0.0.1"}, ips)
}

// TestIsIPAllowed removed - function signature changed to accept fiber.Ctx
// and is tested indirectly through middleware integration tests
