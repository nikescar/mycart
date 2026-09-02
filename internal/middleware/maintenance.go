package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/mycart/pkg/runtime"
)

// MaintenanceMode middleware blocks requests when in maintenance mode
// except for localhost IPs and authenticated admin users
func MaintenanceMode() fiber.Handler {
	return func(c fiber.Ctx) error {
		if !isMaintenanceMode() {
			return c.Next()
		}

		if isAuthorized(c) {
			return c.Next()
		}

		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "System is in maintenance mode",
			"message": "Please try again later",
		})
	}
}

func isMaintenanceMode() bool {
	flagPath := runtime.GetMaintenanceFlagPath()
	_, err := os.Stat(flagPath)
	return err == nil
}

func isAuthorized(c fiber.Ctx) bool {
	return isIPAllowed(c) || isAuthenticatedAdmin(c)
}

func isIPAllowed(c fiber.Ctx) bool {
	allowedIPs := loadAllowedIPs()
	clientIP := c.IP()

	for _, allowed := range allowedIPs {
		if allowed == clientIP {
			return true
		}
	}
	return false
}

func loadAllowedIPs() []string {
	ips := os.Getenv("MAINTENANCE_ALLOWED_IPS")
	if ips == "" {
		return []string{"127.0.0.1", "::1"}
	}

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

func isAuthenticatedAdmin(c fiber.Ctx) bool {
	user := c.Locals("user")
	if user == nil {
		return false
	}

	userMap, ok := user.(map[string]interface{})
	if !ok {
		return false
	}

	role, ok := userMap["role"].(string)
	return ok && role == "admin"
}
