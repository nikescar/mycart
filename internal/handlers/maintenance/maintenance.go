package maintenance

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/mycart/internal/config"
)

// GetStatus returns current maintenance mode status
func GetStatus(c fiber.Ctx) error {
	dbConfig := config.LoadDatabaseConfig()
	storageConfig := config.LoadStorageConfig()

	wranglerInstalled := isWranglerInstalled()

	return c.JSON(fiber.Map{
		"maintenance_mode": isMaintenanceMode(),
		"database": fiber.Map{
			"type": dbConfig.Type,
			"path": dbConfig.Path,
		},
		"storage": fiber.Map{
			"type":      storageConfig.Type,
			"base_path": storageConfig.BasePath,
		},
		"wrangler": fiber.Map{
			"installed": wranglerInstalled,
		},
	})
}

// InstallWrangler installs wrangler via npm
func InstallWrangler(c fiber.Ctx) error {
	cmd := exec.Command("npm", "i", "-D", "wrangler@latest")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  fmt.Sprintf("wrangler installation failed: %v", err),
			"output": string(output),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Wrangler installed successfully",
		"output":  string(output),
	})
}

// UninstallWrangler uninstalls wrangler via npm
func UninstallWrangler(c fiber.Ctx) error {
	cmd := exec.Command("npm", "uninstall", "wrangler")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  fmt.Sprintf("wrangler uninstall failed: %v", err),
			"output": string(output),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Wrangler uninstalled successfully",
		"output":  string(output),
	})
}

// Helper functions

func isMaintenanceMode() bool {
	_, err := os.Stat(".maintenance")
	return err == nil
}

func isWranglerInstalled() bool {
	cmd := exec.Command("npx", "wrangler", "--version")
	return cmd.Run() == nil
}
