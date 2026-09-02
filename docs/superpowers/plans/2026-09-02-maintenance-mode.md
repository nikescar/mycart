# Maintenance Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add maintenance mode to myCart for safe database operations with IP-restricted access, CLI control, and admin UI.

**Architecture:** File-based toggle (`.maintenance` file) triggers middleware that blocks non-localhost IPs. CLI commands manage state. Admin UI provides backup/restore/switch operations via API handlers that execute wrangler for Cloudflare D1/R2.

**Tech Stack:** Go Fiber v3, Cobra CLI, Svelte, wrangler CLI

**Spec:** `docs/superpowers/specs/2026-09-02-maintenance-mode-design.md`

## Global Constraints

- Go 1.21+ required
- Use existing Fiber v3 graceful shutdown mechanism (no modifications)
- Backup files must be in `./lc_base/backups/` directory only (security)
- Wrangler commands timeout after 5 minutes
- IP allowlist defaults to `127.0.0.1,::1`
- All API errors return JSON format: `{"error": "message", "output": "command output"}`
- Follow existing patterns: middleware in `internal/middleware/`, handlers in `internal/handlers/`

---

### Task 1: Maintenance Mode Middleware

**Files:**
- Create: `internal/middleware/maintenance.go`
- Create: `internal/middleware/maintenance_test.go`

**Interfaces:**
- Produces: `MaintenanceMode() fiber.Handler` - middleware function for app.Use()
- Produces: `MaintenanceFile string = ".maintenance"` - constant for file path

- [ ] **Step 1: Write failing test for middleware blocking external IP**

Create `internal/middleware/maintenance_test.go`:

```go
package middleware

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestMaintenanceMode_BlocksExternalIP(t *testing.T) {
	// Setup: create .maintenance file
	f, err := os.Create(".maintenance")
	assert.NoError(t, err)
	f.Close()
	defer os.Remove(".maintenance")

	// Set default allowed IPs (localhost only)
	os.Unsetenv("MAINTENANCE_ALLOWED_IPS")

	app := fiber.New()
	app.Use(MaintenanceMode())
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test: request from external IP
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/... -v -run TestMaintenanceMode_BlocksExternalIP`  
Expected: FAIL with "undefined: MaintenanceMode"

- [ ] **Step 3: Implement maintenance middleware**

Create `internal/middleware/maintenance.go`:

```go
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
	allowedIPs := loadAllowedIPs()

	return func(c fiber.Ctx) error {
		// Check if maintenance file exists
		if !isMaintenanceMode() {
			return c.Next()
		}

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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/... -v -run TestMaintenanceMode_BlocksExternalIP`  
Expected: PASS

- [ ] **Step 5: Write additional middleware tests**

Add to `internal/middleware/maintenance_test.go`:

```go
func TestMaintenanceMode_AllowsLocalhostWhenEnabled(t *testing.T) {
	// Setup: create .maintenance file
	f, err := os.Create(".maintenance")
	assert.NoError(t, err)
	f.Close()
	defer os.Remove(".maintenance")

	os.Unsetenv("MAINTENANCE_ALLOWED_IPS")

	app := fiber.New()
	app.Use(MaintenanceMode())
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test: request from localhost
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMaintenanceMode_AllowsAllWhenDisabled(t *testing.T) {
	// Ensure .maintenance file doesn't exist
	os.Remove(".maintenance")

	app := fiber.New()
	app.Use(MaintenanceMode())
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test: request from external IP should pass
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestMaintenanceMode_AllowsCustomIPs(t *testing.T) {
	// Setup: create .maintenance file
	f, err := os.Create(".maintenance")
	assert.NoError(t, err)
	f.Close()
	defer os.Remove(".maintenance")

	// Set custom allowed IPs
	os.Setenv("MAINTENANCE_ALLOWED_IPS", "127.0.0.1,192.168.1.100")
	defer os.Unsetenv("MAINTENANCE_ALLOWED_IPS")

	app := fiber.New()
	app.Use(MaintenanceMode())
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test: request from custom allowed IP
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
```

- [ ] **Step 6: Run all middleware tests**

Run: `go test ./internal/middleware/... -v`  
Expected: All tests PASS

- [ ] **Step 7: Commit middleware implementation**

```bash
git add internal/middleware/maintenance.go internal/middleware/maintenance_test.go
git commit -m "feat(middleware): add maintenance mode middleware with IP allowlist

- Block non-localhost IPs when .maintenance file exists
- Support MAINTENANCE_ALLOWED_IPS env var for custom IPs
- Default to 127.0.0.1 and ::1 (localhost)
- Return 503 with JSON error for blocked requests

Claude-Session: https://claude.ai/code/session_01JRXRmaAUbbewBBNq4f1eSr"
```

---

### Task 2: CLI Commands for Maintenance Mode

**Files:**
- Modify: `cmd/main.go:70-177` (add cmdMaintenance after existing commands)

**Interfaces:**
- Consumes: `middleware.MaintenanceFile string` - file path constant from Task 1
- Produces: `cmdMaintenance() *cobra.Command` - root maintenance command

- [ ] **Step 1: Write test for enable command**

Create test file `cmd/maintenance_test.go`:

```go
package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaintenanceEnable(t *testing.T) {
	// Ensure clean state
	os.Remove(".maintenance")

	// Execute enable command
	cmd := cmdMaintenanceEnable()
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	assert.NoError(t, err)

	// Verify file created
	_, err = os.Stat(".maintenance")
	assert.NoError(t, err, ".maintenance file should exist")

	// Cleanup
	os.Remove(".maintenance")
}

func TestMaintenanceDisable(t *testing.T) {
	// Setup: create .maintenance file
	f, err := os.Create(".maintenance")
	assert.NoError(t, err)
	f.Close()

	// Execute disable command
	cmd := cmdMaintenanceDisable()
	cmd.SetArgs([]string{})
	err = cmd.Execute()

	assert.NoError(t, err)

	// Verify file removed
	_, err = os.Stat(".maintenance")
	assert.True(t, os.IsNotExist(err), ".maintenance file should not exist")
}

func TestMaintenanceStatus(t *testing.T) {
	// Test when disabled
	os.Remove(".maintenance")
	cmd := cmdMaintenanceStatus()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)

	// Test when enabled
	f, err := os.Create(".maintenance")
	assert.NoError(t, err)
	f.Close()
	defer os.Remove(".maintenance")

	cmd = cmdMaintenanceStatus()
	cmd.SetArgs([]string{})
	err = cmd.Execute()
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -v -run TestMaintenance`  
Expected: FAIL with "undefined: cmdMaintenanceEnable"

- [ ] **Step 3: Implement CLI commands**

Add to `cmd/main.go` after line 68 (after `rootCmd.AddCommand(cmdMigrate())`):

```go
rootCmd.AddCommand(cmdMaintenance())
```

Add before the closing `}` of `main.go` (around line 177):

```go
// cmdMaintenance creates and returns the maintenance command with subcommands.
func cmdMaintenance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance",
		Short: "Manage maintenance mode",
	}

	cmd.AddCommand(cmdMaintenanceEnable())
	cmd.AddCommand(cmdMaintenanceDisable())
	cmd.AddCommand(cmdMaintenanceStatus())

	return cmd
}

// cmdMaintenanceEnable creates and returns the enable subcommand.
func cmdMaintenanceEnable() *cobra.Command {
	var restart bool

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable maintenance mode",
		Run: func(_ *cobra.Command, _ []string) {
			// Check if already in maintenance mode
			if _, err := os.Stat(".maintenance"); err == nil {
				fmt.Println("⚠️  Maintenance mode is already enabled")
				return
			}

			// Create .maintenance file with timestamp
			f, err := os.Create(".maintenance")
			if err != nil {
				handleCommandError(fmt.Errorf("failed to enable maintenance mode: %w", err))
				return
			}
			defer f.Close()

			// Write timestamp for audit trail
			timestamp := time.Now().Format(time.RFC3339)
			if _, err := f.WriteString(timestamp); err != nil {
				fmt.Printf("⚠️  Warning: failed to write timestamp: %v\n", err)
			}

			fmt.Println("✓ Maintenance mode enabled")
			fmt.Println("  Only localhost can access the application")
			fmt.Println("  Access maintenance panel at: http://localhost:8080/_/maintenance")

			if restart {
				fmt.Println("\n⟳ Restarting server...")
				syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			}
		},
	}

	cmd.Flags().BoolVar(&restart, "restart", false, "restart server after enabling")
	return cmd
}

// cmdMaintenanceDisable creates and returns the disable subcommand.
func cmdMaintenanceDisable() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable maintenance mode",
		Run: func(_ *cobra.Command, _ []string) {
			if err := os.Remove(".maintenance"); err != nil {
				if os.IsNotExist(err) {
					fmt.Println("Maintenance mode is not enabled")
					return
				}
				handleCommandError(fmt.Errorf("failed to disable maintenance mode: %w", err))
				return
			}

			fmt.Println("✓ Maintenance mode disabled")
			fmt.Println("  Application is now accessible to all users")
		},
	}
}

// cmdMaintenanceStatus creates and returns the status subcommand.
func cmdMaintenanceStatus() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check maintenance mode status",
		Run: func(_ *cobra.Command, _ []string) {
			_, err := os.Stat(".maintenance")
			if err == nil {
				fmt.Println("Status: MAINTENANCE MODE")
				fmt.Println("  Access restricted to allowed IPs only")
			} else {
				fmt.Println("Status: NORMAL")
				fmt.Println("  Application is publicly accessible")
			}
		},
	}
}
```

Add required imports at top of `cmd/main.go`:

```go
import (
	// ... existing imports ...
	"syscall"
	"time"
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/... -v -run TestMaintenance`  
Expected: All tests PASS

- [ ] **Step 5: Test CLI manually**

```bash
# Build
go build -o mycart ./cmd

# Test enable
./mycart maintenance enable
# Expected: ✓ Maintenance mode enabled

# Verify file exists
ls -la .maintenance
# Expected: file exists with timestamp

# Test status
./mycart maintenance status
# Expected: Status: MAINTENANCE MODE

# Test disable
./mycart maintenance disable
# Expected: ✓ Maintenance mode disabled

# Verify file removed
ls -la .maintenance
# Expected: file not found

# Cleanup
rm mycart
```

- [ ] **Step 6: Commit CLI commands**

```bash
git add cmd/main.go cmd/maintenance_test.go
git commit -m "feat(cli): add maintenance mode commands (enable/disable/status)

- Add 'mycart maintenance enable [--restart]' command
- Add 'mycart maintenance disable' command
- Add 'mycart maintenance status' command
- Write timestamp to .maintenance file for audit trail
- Support --restart flag to trigger graceful shutdown

Claude-Session: https://claude.ai/code/session_01JRXRmaAUbbewBBNq4f1eSr"
```

---

---

### Task 3: API Handlers - Status & Wrangler Operations

**Files:**
- Create: `internal/handlers/maintenance/maintenance.go`
- Create: `internal/handlers/maintenance/maintenance_test.go`

**Interfaces:**
- Consumes: `config.LoadDatabaseConfig()`, `config.LoadStorageConfig()` - from existing config package
- Produces: `GetStatus(c fiber.Ctx) error` - GET /_/api/maintenance/status
- Produces: `InstallWrangler(c fiber.Ctx) error` - POST /_/api/maintenance/wrangler/install
- Produces: `UninstallWrangler(c fiber.Ctx) error` - POST /_/api/maintenance/wrangler/uninstall
- Produces: `isWranglerInstalled() bool`, `isMaintenanceMode() bool` - helper functions

**Implementation:** Create API handlers for status endpoint and wrangler operations. Follow TDD: write tests → run (fail) → implement → run (pass) → commit.

---

### Task 4: API Handlers - Backup & Restore Operations

**Files:**
- Modify: `internal/handlers/maintenance/maintenance.go` (add backup/restore functions)
- Modify: `internal/handlers/maintenance/maintenance_test.go` (add tests)

**Interfaces:**
- Produces: `BackupDatabase(c fiber.Ctx) error` - POST /_/api/maintenance/backup
- Produces: `RestoreDatabase(c fiber.Ctx) error` - POST /_/api/maintenance/restore
- Produces: `copyFile(src, dst string) error`, `exportD1ToSQLite(databaseID, outputPath string) error`, `importSQLiteToD1(sqlitePath, databaseID string) error` - helper functions

**Implementation:** Add backup/restore handlers with path validation and wrangler integration. Follow TDD pattern.

---

### Task 5: API Handlers - Database Switch & Restart

**Files:**
- Modify: `internal/handlers/maintenance/maintenance.go` (add switch/restart functions)
- Modify: `internal/handlers/maintenance/maintenance_test.go` (add tests)

**Interfaces:**
- Produces: `SwitchDatabase(c fiber.Ctx) error` - POST /_/api/maintenance/switch
- Produces: `EnableMaintenanceAndRestart(c fiber.Ctx) error` - POST /_/api/maintenance/enable-and-restart

**Implementation:** Add database switch and restart handlers. Follow TDD pattern.

---

### Task 6: Integrate Middleware and Routes

**Files:**
- Modify: `internal/app.go` (add middleware)
- Modify: `internal/routes/routes.go` (add API routes)

**Interfaces:**
- Consumes: `middleware.MaintenanceMode() fiber.Handler` - from Task 1
- Consumes: All handler functions from Tasks 3-5

**Implementation:** Wire middleware and routes into application. Test manually.

---

### Task 7: Maintenance UI Panel

**Files:**
- Create: `web/admin/src/routes/_/maintenance/+page.svelte`

**Interfaces:**
- Consumes: All API endpoints from Tasks 3-5

**Implementation:** Create Svelte UI for maintenance operations. Test manually in browser.

---

### Task 8: Settings Restart Button

**Files:**
- Create: `web/admin/src/routes/_/settings/maintenance/+page.svelte`

**Interfaces:**
- Consumes: `EnableMaintenanceAndRestart(c fiber.Ctx) error` - from Task 5

**Implementation:** Create settings page with restart button. Test manually in browser.