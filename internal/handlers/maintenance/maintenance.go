package maintenance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/mycart/internal/config"
	"github.com/shurco/mycart/internal/handlers/cloudflare"
	"github.com/shurco/mycart/pkg/runtime"
)

// GetStatus returns current maintenance mode status
func GetStatus(c fiber.Ctx) error {
	dbConfig := config.LoadDatabaseConfig()
	storageConfig := config.LoadStorageConfig()

	return c.JSON(fiber.Map{
		"maintenance_mode": isMaintenanceMode(),
		"runtime":          map[string]bool{"cloudflare": runtime.IsCloudflare()},
		"database": fiber.Map{
			"type": dbConfig.Type,
			"path": dbConfig.Path,
		},
		"storage": fiber.Map{
			"type":      storageConfig.Type,
			"base_path": storageConfig.BasePath,
		},
	})
}

// Helper functions

func isMaintenanceMode() bool {
	flagPath := runtime.GetMaintenanceFlagPath()
	_, err := os.Stat(flagPath)
	return err == nil
}

// BackupDatabase creates a backup of the current database
func BackupDatabase(c fiber.Ctx) error {
	dbConfig := config.LoadDatabaseConfig()
	timestamp := time.Now().Format("20060102-150405")
	backupDir := "./lc_base/backups"

	os.MkdirAll(backupDir, 0755)

	var backupPath string
	var err error

	switch dbConfig.Type {
	case "sqlite":
		backupPath = filepath.Join(backupDir, fmt.Sprintf("backup-%s.db", timestamp))
		err = copyFile(dbConfig.Path, backupPath)
	case "d1":
		backupPath = filepath.Join(backupDir, fmt.Sprintf("backup-d1-%s.db", timestamp))
		err = exportD1ToSQLite(dbConfig.DatabaseID, backupPath)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("unsupported database type: %s", dbConfig.Type),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("backup failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message":     "Backup created successfully",
		"backup_path": backupPath,
	})
}

// RestoreDatabase restores from a backup file
func RestoreDatabase(c fiber.Ctx) error {
	var req struct {
		BackupPath string `json:"backup_path"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if _, err := os.Stat(req.BackupPath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("backup file not found: %s", req.BackupPath),
		})
	}

	allowedDir, _ := filepath.Abs("./lc_base/backups")
	absBackupPath, _ := filepath.Abs(req.BackupPath)
	if !strings.HasPrefix(absBackupPath, allowedDir) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "backup file must be in ./lc_base/backups directory",
		})
	}

	dbConfig := config.LoadDatabaseConfig()

	var err error
	switch dbConfig.Type {
	case "sqlite":
		err = copyFile(req.BackupPath, dbConfig.Path)
	case "d1":
		err = importSQLiteToD1(req.BackupPath, dbConfig.DatabaseID)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("unsupported database type: %s", dbConfig.Type),
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("restore failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Database restored successfully",
	})
}

// SwitchDatabase switches between SQLite and D1
func SwitchDatabase(c fiber.Ctx) error {
	var req struct {
		TargetType string `json:"target_type"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	switch req.TargetType {
	case "sqlite":
		os.Setenv("DB_TYPE", "sqlite")
		os.Unsetenv("CF_WORKER")
	case "d1":
		os.Setenv("DB_TYPE", "d1")
		os.Setenv("CF_WORKER", "true")
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "target_type must be 'sqlite' or 'd1'",
		})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Database switched to %s (restart required to take effect)", req.TargetType),
		"note":    "Update your .env file to persist this change",
	})
}

// EnableMaintenanceAndRestart enables maintenance mode and triggers graceful restart
func EnableMaintenanceAndRestart(c fiber.Ctx) error {
	flagPath := runtime.GetMaintenanceFlagPath()

	f, err := os.Create(flagPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to enable maintenance mode: %v", err),
		})
	}
	f.Close()

	go func() {
		time.Sleep(1 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	return c.JSON(fiber.Map{
		"message": "Maintenance mode enabled. Server restarting...",
	})
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	return os.WriteFile(dst, input, 0644)
}

func exportD1ToSQLite(databaseID, outputPath string) error {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	if accountID == "" || apiToken == "" {
		return fmt.Errorf("CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN must be set for D1 operations")
	}

	client := cloudflare.NewD1Client(accountID, apiToken)
	return client.ExportDatabase(databaseID, outputPath)
}

func importSQLiteToD1(sqlitePath, databaseID string) error {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	if accountID == "" || apiToken == "" {
		return fmt.Errorf("CLOUDFLARE_ACCOUNT_ID and CLOUDFLARE_API_TOKEN must be set for D1 operations")
	}

	client := cloudflare.NewD1Client(accountID, apiToken)
	return client.ImportSQLite(databaseID, sqlitePath)
}
