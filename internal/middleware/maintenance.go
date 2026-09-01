package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
)

const MaintenanceFile = ".maintenance"

// MaintenanceMode middleware blocks requests when in maintenance mode
// except for allowed IPs (localhost by default)
func MaintenanceMode() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Check if maintenance file exists
		if !isMaintenanceMode() {
			return c.Next()
		}

		// Load allowed IPs on each request (for testing flexibility)
		allowedIPs := loadAllowedIPs()

		// Check if request IP is allowed
		clientIP := c.IP()
		if !isIPAllowed(clientIP, allowedIPs) {
			// Return 503 Service Unavailable
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":       "System is in maintenance mode",
				"allowed_ips": allowedIPs,
			})
		}

		return c.Next()
	}
}

func isMaintenanceMode() bool {
	_, err := os.Stat(MaintenanceFile)
	return err == nil // file exists = maintenance mode ON
}

func loadAllowedIPs() []string {
	ips := os.Getenv("MAINTENANCE_ALLOWED_IPS")
	if ips == "" {
		return []string{"127.0.0.1", "::1"} // localhost only by default
	}

	// Split and trim whitespace
	parts := strings.Split(ips, ",")
	result := make([]string, 0, len(parts))
	for _, ip := range parts {
		trimmed := strings.TrimSpace(ip)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func isIPAllowed(clientIP string, allowedIPs []string) bool {
	for _, allowed := range allowedIPs {
		if allowed == clientIP {
			return true
		}
	}
	return false
}
