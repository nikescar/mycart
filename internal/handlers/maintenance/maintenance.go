package maintenance

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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
	f, err := os.Create(".maintenance")
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npx", "wrangler", "d1", "export", databaseID, "--output", outputPath)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("wrangler export timed out after 5 minutes")
	}

	if err != nil {
		return fmt.Errorf("wrangler export failed: %w\nOutput: %s", err, output)
	}

	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("backup file not created: %w", err)
	}

	return nil
}

func importSQLiteToD1(sqlitePath, databaseID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "npx", "wrangler", "d1", "execute", databaseID, "--file", sqlitePath)
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("wrangler import timed out after 5 minutes")
	}

	if err != nil {
		return fmt.Errorf("wrangler import failed: %w\nOutput: %s", err, output)
	}

	return nil
}
