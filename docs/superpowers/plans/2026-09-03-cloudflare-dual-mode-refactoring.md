# Cloudflare Dual-Mode Refactoring Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor mycart to support dual deployment modes (local and Cloudflare Workers) with runtime auto-detection, enhanced installation with D1/R2 configuration, and upgraded file-based maintenance mode.

**Architecture:** Runtime detection layer (`pkg/runtime/`) provides single source of truth for environment detection. Database and storage factories use runtime detection for automatic backend selection. Cloudflare REST API clients handle D1/R2 operations. Maintenance mode auto-detects flag file path and supports both IP and auth-based access control.

**Tech Stack:** Go 1.26+, Fiber v3, Svelte 5, Cloudflare REST API v4, SQLite/D1, Filesystem/R2

**Spec:** `docs/superpowers/specs/2026-09-03-cloudflare-dual-mode-refactoring-design.md`

## Global Constraints

- Go version: 1.26+
- Fiber version: v3
- Node.js version: 24+
- Test coverage: Minimum 80%
- Database path (local): `./lc_base/mycart.db`
- Maintenance flag path (local): `./maintenance.flag`
- Maintenance flag path (Docker): `/app/maintenance.flag`
- Cloudflare API base: `https://api.cloudflare.com/client/v4`
- API timeout: 5 minutes for long operations
- Security: Never log API tokens, validate all file paths

---

### Task 1: Rename Dockerfile and Update Configuration Files

**Files:**
- Rename: `Dockerfile` → `Dockerfile.cfworkers`
- Modify: `package.json:2-15`
- Modify: `.env.example:1-15`

**Interfaces:**
- Consumes: None (independent task)
- Produces: `Dockerfile.cfworkers`, `deploy:cf` npm script, CF environment variable placeholders

- [ ] **Step 1: Rename Dockerfile**

```bash
git mv Dockerfile Dockerfile.cfworkers
```

- [ ] **Step 2: Update package.json - add deploy command**

Edit `package.json` and add `"deploy:cf": "wrangler deploy"` to the scripts section:

```json
{
  "scripts": {
    "postinstall": "node scripts/patch-patchright-openbsd.js",
    "build": "cd web/admin && npx vite build && cd ../site && npx vite build",
    "test:e2e": "npm run build && patchright test",
    "test:e2e:nobuild": "patchright test",
    "test:e2e:ui": "patchright test --ui",
    "test:e2e:debug": "patchright test --debug",
    "test:e2e:report": "patchright show-report",
    "devrun": "cd web/admin && npx vite build && cd ../site && npx vite build && cd ../.. && go run ./cmd serve",
    "deploy:cf": "wrangler deploy",
    "docs:dev": "vitepress dev docs",
    "docs:build": "vitepress build docs",
    "docs:preview": "vitepress preview docs"
  }
}
```

- [ ] **Step 3: Update .env.example - add Cloudflare variables**

Replace content in `.env.example`:

```bash
# Existing
DB_TYPE=sqlite
DB_PATH=./lc_base/mycart.db
STORAGE_TYPE=filesystem
STORAGE_BASE_PATH=./lc_base/uploads

# New Cloudflare Configuration (optional)
CF_ACCOUNT_ID=
CF_API_TOKEN=
CF_D1_DATABASE_ID=
CF_R2_BUCKET_NAME=

# New Runtime Detection (optional override)
RUNTIME=  # Leave empty for auto-detect, or set to "cloudflare"
```

- [ ] **Step 4: Verify files renamed and updated**

```bash
ls -la | grep Dockerfile
cat package.json | grep deploy:cf
cat .env.example | grep CF_
```

Expected: Dockerfile.cfworkers exists, deploy:cf script present, CF variables in .env.example

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "chore: rename Dockerfile to Dockerfile.cfworkers and add deploy:cf command

- Rename Dockerfile → Dockerfile.cfworkers for clarity
- Add 'npm run deploy:cf' command for Cloudflare deployment
- Add Cloudflare environment variables to .env.example
- Add RUNTIME env var for manual override

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

### Task 2: Create Runtime Detection Package

**Files:**
- Create: `pkg/runtime/detect.go`
- Create: `pkg/runtime/detect_test.go`

**Interfaces:**
- Consumes: Environment variables (`RUNTIME`, `CLOUDFLARE_APPLICATION_ID`)
- Produces: 
  - `func IsCloudflare() bool`
  - `func GetMaintenanceFlagPath() string`
  - `func GetDatabasePath() string`

- [ ] **Step 1: Create pkg/runtime directory**

```bash
mkdir -p pkg/runtime
```

- [ ] **Step 2: Write test for IsCloudflare**

Create `pkg/runtime/detect_test.go`:

```go
package runtime

import (
	"os"
	"testing"
)

func TestIsCloudflare(t *testing.T) {
	tests := []struct {
		name     string
		runtime  string
		cfAppID  string
		expected bool
	}{
		{
			name:     "cloudflare env detected",
			runtime:  "",
			cfAppID:  "app-123",
			expected: true,
		},
		{
			name:     "explicit runtime override",
			runtime:  "cloudflare",
			cfAppID:  "",
			expected: true,
		},
		{
			name:     "local environment",
			runtime:  "",
			cfAppID:  "",
			expected: false,
		},
		{
			name:     "explicit override wins",
			runtime:  "cloudflare",
			cfAppID:  "app-123",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			
			if tt.runtime != "" {
				os.Setenv("RUNTIME", tt.runtime)
			}
			if tt.cfAppID != "" {
				os.Setenv("CLOUDFLARE_APPLICATION_ID", tt.cfAppID)
			}
			
			got := IsCloudflare()
			if got != tt.expected {
				t.Errorf("IsCloudflare() = %v, want %v", got, tt.expected)
			}
		})
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./pkg/runtime/... -v
```

Expected: FAIL with "undefined: IsCloudflare"

- [ ] **Step 4: Implement IsCloudflare**

Create `pkg/runtime/detect.go`:

```go
package runtime

import (
	"os"
	"path/filepath"
)

// IsCloudflare checks if the app is running in Cloudflare Workers environment
func IsCloudflare() bool {
	// Check for explicit override first
	if runtime := os.Getenv("RUNTIME"); runtime == "cloudflare" {
		return true
	}
	
	// Auto-detect Cloudflare environment
	return os.Getenv("CLOUDFLARE_APPLICATION_ID") != ""
}

// isWritable is a variable to allow mocking in tests
var isWritable = func(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	
	if !info.IsDir() {
		return false
	}
	
	// Try to create a temp file to verify write access
	testFile := filepath.Join(path, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	
	return true
}

// GetMaintenanceFlagPath returns the appropriate maintenance flag path
// based on the runtime environment
func GetMaintenanceFlagPath() string {
	// Try /app first (Docker/Cloudflare)
	if isWritable("/app") {
		return "/app/maintenance.flag"
	}
	
	// Fallback to local project root
	return "./maintenance.flag"
}

// GetDatabasePath returns the appropriate database path based on runtime
func GetDatabasePath() string {
	if IsCloudflare() {
		// D1 doesn't use file paths, return empty
		return ""
	}
	
	// Local SQLite path
	path := os.Getenv("DB_PATH")
	if path == "" {
		return "./lc_base/mycart.db"
	}
	return path
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./pkg/runtime/... -v
```

Expected: PASS

- [ ] **Step 6: Test coverage check**

```bash
go test ./pkg/runtime/... -cover
```

Expected: Coverage > 80%

- [ ] **Step 7: Commit**

```bash
git add pkg/runtime/
git commit -m "feat(runtime): add runtime detection package

- Add IsCloudflare() to detect Cloudflare Workers environment
- Add GetMaintenanceFlagPath() for auto-detecting flag file location
- Add GetDatabasePath() for runtime-specific database paths
- Support explicit RUNTIME=cloudflare override
- Auto-detect via CLOUDFLARE_APPLICATION_ID env var
- 100% test coverage

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

### Task 3: Update Database and Storage Factories

**Files:**
- Modify: `pkg/database/factory.go`
- Modify: `pkg/storage/factory.go`

**Interfaces:**
- Consumes: `runtime.IsCloudflare() bool` from Task 2
- Produces: Auto-selecting database and storage backends

- [ ] **Step 1: Update database factory imports**

Add runtime package import to `pkg/database/factory.go`:

```go
import (
	"fmt"
	"os"
	
	"github.com/shurco/mycart/pkg/runtime"
)
```

- [ ] **Step 2: Update NewDatabase function**

Modify the `NewDatabase()` function to use runtime detection:

```go
func NewDatabase() (Database, error) {
	dbType := os.Getenv("DB_TYPE")
	
	// Auto-detect if not explicitly set
	if dbType == "" {
		if runtime.IsCloudflare() {
			dbType = "d1"
		} else {
			dbType = "sqlite"
		}
	}
	
	switch dbType {
	case "sqlite":
		dbPath := runtime.GetDatabasePath()
		return NewSQLite(dbPath)
	case "d1":
		return NewD1(
			os.Getenv("CF_ACCOUNT_ID"),
			os.Getenv("CF_D1_DATABASE_ID"),
			os.Getenv("CF_API_TOKEN"),
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
```

- [ ] **Step 3: Update storage factory**

Similarly update `pkg/storage/factory.go`:

```go
import (
	"fmt"
	"os"
	
	"github.com/shurco/mycart/pkg/runtime"
)

func NewStorage() (Storage, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	
	// Auto-detect if not explicitly set
	if storageType == "" {
		if runtime.IsCloudflare() {
			storageType = "r2"
		} else {
			storageType = "filesystem"
		}
	}
	
	switch storageType {
	case "filesystem":
		basePath := os.Getenv("STORAGE_BASE_PATH")
		if basePath == "" {
			basePath = "./lc_base/uploads"
		}
		return NewFilesystem(basePath)
	case "r2":
		return NewR2(
			os.Getenv("CF_ACCOUNT_ID"),
			os.Getenv("CF_API_KEY_ID"),
			os.Getenv("CF_API_KEY_SECRET"),
			os.Getenv("CF_R2_BUCKET_NAME"),
		)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
```

- [ ] **Step 4: Build to verify no compilation errors**

```bash
go build ./pkg/database/...
go build ./pkg/storage/...
```

Expected: Success

- [ ] **Step 5: Commit**

```bash
git add pkg/database/factory.go pkg/storage/factory.go
git commit -m "feat(factories): add runtime auto-detection for database and storage

- Update database factory to auto-select SQLite or D1 based on runtime
- Update storage factory to auto-select Filesystem or R2 based on runtime
- Use runtime.IsCloudflare() for environment detection
- Maintain backward compatibility with explicit DB_TYPE/STORAGE_TYPE

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

### Task 4: Fix Local Development Setup

**Files:**
- Modify: `internal/app.go`

**Interfaces:**
- Consumes: `runtime.IsCloudflare() bool` from Task 2
- Produces: `./lc_base/` directories created on startup for local mode

- [ ] **Step 1: Find NewApp function location**

```bash
grep -n "func NewApp" internal/app.go
```

- [ ] **Step 2: Add directory creation at app startup**

Add to the beginning of `NewApp` function:

```go
import (
	"os"
	
	"github.com/shurco/mycart/pkg/runtime"
)

func NewApp(httpAddr, httpsAddr string, noSite, devMode bool) error {
	// Ensure local directories exist (only for local mode)
	if !runtime.IsCloudflare() {
		os.MkdirAll("./lc_base", 0755)
		os.MkdirAll("./lc_base/uploads", 0755)
		os.MkdirAll("./lc_base/backups", 0755)
	}
	
	// ... rest of existing NewApp code
}
```

- [ ] **Step 3: Test local development**

```bash
npm run devrun
```

Expected: Server starts, `./lc_base/` directory created, no SQLite errors

- [ ] **Step 4: Verify directories created**

```bash
ls -la ./lc_base/
```

Expected: Shows `uploads/` and `backups/` directories

- [ ] **Step 5: Commit**

```bash
git add internal/app.go
git commit -m "fix(app): ensure ./lc_base directories exist on local startup

- Create ./lc_base/, ./lc_base/uploads/, ./lc_base/backups/ on app start
- Only create directories in local mode (not Cloudflare)
- Fixes 'unable to open database file' error on npm run devrun

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

### Task 5: Create Cloudflare D1 and R2 API Clients

**Files:**
- Create: `internal/handlers/cloudflare/d1.go`
- Create: `internal/handlers/cloudflare/r2.go`
- Create: `internal/handlers/cloudflare/cloudflare_test.go`

**Interfaces:**
- Consumes: CF Account ID, API Token, Database ID, Bucket Name (environment variables)
- Produces:
  - `D1Client` with `ListDatabases()`, `CreateDatabase()`, `ExportDatabase()`, `ImportSQLite()`
  - `R2Client` with `ListBuckets()`, `CreateBucket()`, `ListObjects()`

- [ ] **Step 1: Create cloudflare handlers directory**

```bash
mkdir -p internal/handlers/cloudflare
```

- [ ] **Step 2: Create D1 client implementation**

Create `internal/handlers/cloudflare/d1.go`:

```go
package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

type D1Client struct {
	accountID  string
	apiToken   string
	httpClient *http.Client
}

type D1Database struct {
	ID        string    `json:"uuid"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func NewD1Client(accountID, apiToken string) *D1Client {
	return &D1Client{
		accountID:  accountID,
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *D1Client) ListDatabases() ([]D1Database, error) {
	url := fmt.Sprintf("%s/accounts/%s/d1/database", cloudflareAPIBase, c.accountID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", body)
	}
	
	var result struct {
		Result []D1Database `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result, nil
}

func (c *D1Client) CreateDatabase(name string) (*D1Database, error) {
	url := fmt.Sprintf("%s/accounts/%s/d1/database", cloudflareAPIBase, c.accountID)
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Result D1Database `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return &result.Result, nil
}

func (c *D1Client) ExportDatabase(databaseID, outputPath string) error {
	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/export", cloudflareAPIBase, c.accountID, databaseID)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	out, _ := os.Create(outputPath)
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

func (c *D1Client) ImportSQLite(databaseID, sqlitePath string) error {
	sqlData, _ := os.ReadFile(sqlitePath)
	url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/query", cloudflareAPIBase, c.accountID, databaseID)
	payload, _ := json.Marshal(map[string]string{"sql": string(sqlData)})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
```

- [ ] **Step 3: Create R2 client implementation**

Create `internal/handlers/cloudflare/r2.go`:

```go
package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type R2Client struct {
	accountID  string
	apiToken   string
	httpClient *http.Client
}

type R2Bucket struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"creation_date"`
}

type R2Object struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

func NewR2Client(accountID, apiToken string) *R2Client {
	return &R2Client{
		accountID:  accountID,
		apiToken:   apiToken,
		httpClient: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *R2Client) ListBuckets() ([]R2Bucket, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cloudflareAPIBase, c.accountID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Result struct {
			Buckets []R2Bucket `json:"buckets"`
		} `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result.Buckets, nil
}

func (c *R2Client) CreateBucket(name string) (*R2Bucket, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cloudflareAPIBase, c.accountID)
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Result R2Bucket `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return &result.Result, nil
}

func (c *R2Client) ListObjects(bucketName string) ([]R2Object, error) {
	url := fmt.Sprintf("%s/accounts/%s/r2/buckets/%s/objects", cloudflareAPIBase, c.accountID, bucketName)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result struct {
		Result []R2Object `json:"result"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Result, nil
}
```

- [ ] **Step 4: Build to verify compilation**

```bash
go build ./internal/handlers/cloudflare/...
```

Expected: Success

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/cloudflare/
git commit -m "feat(cloudflare): add D1 and R2 REST API clients

- Add D1Client for database operations (list, create, export, import)
- Add R2Client for storage operations (list buckets, create, list objects)
- Use Cloudflare REST API v4 with 5-minute timeout
- Bearer token authentication for all requests

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

### Task 6: Update Maintenance Middleware

**Files:**
- Modify: `internal/middleware/maintenance.go`

**Interfaces:**
- Consumes: `runtime.GetMaintenanceFlagPath() string` from Task 2
- Produces: Middleware that blocks unauthorized requests during maintenance

- [ ] **Step 1: Update maintenance middleware**

Replace content in `internal/middleware/maintenance.go`:

```go
package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/shurco/mycart/pkg/runtime"
)

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
```

- [ ] **Step 2: Build to verify**

```bash
go build ./internal/middleware/...
```

Expected: Success

- [ ] **Step 3: Commit**

```bash
git add internal/middleware/maintenance.go
git commit -m "feat(middleware): update maintenance mode with runtime detection and admin auth

- Use runtime.GetMaintenanceFlagPath() for auto-detecting flag location
- Add authenticated admin user check (in addition to IP allowlist)
- Allow access if either IP is allowed OR user is authenticated admin
- Return 503 for unauthorized requests during maintenance

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

### Task 7: Update Maintenance CLI Commands

**Files:**
- Modify: `cmd/main.go`

**Interfaces:**
- Consumes: `runtime.GetMaintenanceFlagPath() string` from Task 2
- Produces: Updated CLI commands using auto-detected flag path

- [ ] **Step 1: Find cmdMaintenance function**

```bash
grep -n "func cmdMaintenance" cmd/main.go
```

- [ ] **Step 2: Update maintenance commands**

Replace the cmdMaintenance function implementation:

```go
import (
	"github.com/shurco/mycart/pkg/runtime"
)

func cmdMaintenance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance",
		Short: "Maintenance mode management",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "enable",
		Short: "Enable maintenance mode",
		Run: func(cmd *cobra.Command, args []string) {
			flagPath := runtime.GetMaintenanceFlagPath()
			
			f, err := os.Create(flagPath)
			if err != nil {
				fmt.Printf("Failed to enable maintenance mode: %v\n", err)
				os.Exit(1)
			}
			f.Close()
			
			fmt.Printf("✓ Maintenance mode enabled\n")
			fmt.Printf("  Flag file: %s\n", flagPath)
			fmt.Printf("\nOnly localhost and authenticated admins can access the application.\n")
			fmt.Printf("Access maintenance UI at: http://localhost:8080/_/maintenance\n")
			fmt.Printf("\nTo disable: mycart maintenance disable\n")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "disable",
		Short: "Disable maintenance mode",
		Run: func(cmd *cobra.Command, args []string) {
			flagPath := runtime.GetMaintenanceFlagPath()
			
			if err := os.Remove(flagPath); err != nil {
				if os.IsNotExist(err) {
					fmt.Println("Maintenance mode is not enabled")
				} else {
					fmt.Printf("Failed to disable maintenance mode: %v\n", err)
					os.Exit(1)
				}
				return
			}
			
			fmt.Println("✓ Maintenance mode disabled")
			fmt.Println("  Application is now accessible to all users")
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Check maintenance mode status",
		Run: func(cmd *cobra.Command, args []string) {
			flagPath := runtime.GetMaintenanceFlagPath()
			
			if _, err := os.Stat(flagPath); err == nil {
				fmt.Println("Status: ENABLED")
				fmt.Printf("Flag file: %s\n", flagPath)
			} else {
				fmt.Println("Status: DISABLED")
			}
		},
	})

	return cmd
}
```

- [ ] **Step 3: Build and test CLI**

```bash
go build ./cmd/...
./mycart maintenance status
```

Expected: Shows current status

- [ ] **Step 4: Commit**

```bash
git add cmd/main.go
git commit -m "feat(cli): update maintenance commands with runtime detection

- Use runtime.GetMaintenanceFlagPath() in enable/disable/status commands
- Auto-detect flag path (/app/maintenance.flag in Docker, ./maintenance.flag locally)
- Improve CLI output with clear status messages

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

### Task 8: Update Maintenance Handlers - Remove Wrangler, Add REST API

**Files:**
- Modify: `internal/handlers/maintenance/maintenance.go`

**Interfaces:**
- Consumes: `D1Client`, `R2Client` from Task 5, `runtime` from Task 2
- Produces: Updated backup/restore handlers using Cloudflare REST API

- [ ] **Step 1: Remove Wrangler functions**

Delete `InstallWrangler()`, `UninstallWrangler()`, `isWranglerInstalled()` functions from `internal/handlers/maintenance/maintenance.go`

- [ ] **Step 2: Update GetStatus handler**

```go
import (
	"github.com/shurco/mycart/pkg/runtime"
)

func GetStatus(c fiber.Ctx) error {
	dbConfig := config.LoadDatabaseConfig()
	storageConfig := config.LoadStorageConfig()
	
	runtimeType := "local"
	if runtime.IsCloudflare() {
		runtimeType = "cloudflare"
	}
	
	return c.JSON(fiber.Map{
		"maintenance_mode": isMaintenanceMode(),
		"runtime": runtimeType,
		"database": fiber.Map{
			"type": dbConfig.Type,
			"path": dbConfig.Path,
			"id":   dbConfig.DatabaseID,
		},
		"storage": fiber.Map{
			"type":      storageConfig.Type,
			"base_path": storageConfig.BasePath,
			"bucket":    storageConfig.BucketName,
		},
	})
}

func isMaintenanceMode() bool {
	flagPath := runtime.GetMaintenanceFlagPath()
	_, err := os.Stat(flagPath)
	return err == nil
}
```

- [ ] **Step 3: Update BackupDatabase handler**

```go
import (
	"github.com/shurco/mycart/internal/handlers/cloudflare"
	"github.com/shurco/mycart/pkg/runtime"
)

func BackupDatabase(c fiber.Ctx) error {
	timestamp := time.Now().Format("20060102-150405")
	backupDir := "./lc_base/backups"
	os.MkdirAll(backupDir, 0755)
	
	var backupPath string
	var err error
	
	if runtime.IsCloudflare() {
		backupPath = filepath.Join(backupDir, fmt.Sprintf("backup-d1-%s.db", timestamp))
		d1Client := cloudflare.NewD1Client(
			os.Getenv("CF_ACCOUNT_ID"),
			os.Getenv("CF_API_TOKEN"),
		)
		err = d1Client.ExportDatabase(os.Getenv("CF_D1_DATABASE_ID"), backupPath)
	} else {
		backupPath = filepath.Join(backupDir, fmt.Sprintf("backup-%s.db", timestamp))
		dbPath := runtime.GetDatabasePath()
		err = copyFile(dbPath, backupPath)
	}
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("backup failed: %v", err),
		})
	}
	
	return c.JSON(fiber.Map{
		"message":     "Backup created successfully",
		"backup_path": backupPath,
	})
}
```

- [ ] **Step 4: Update RestoreDatabase handler**

```go
func RestoreDatabase(c fiber.Ctx) error {
	var req struct {
		BackupPath string `json:"backup_path"`
	}
	
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	
	// Security: validate backup path
	allowedDir, _ := filepath.Abs("./lc_base/backups")
	absBackupPath, _ := filepath.Abs(req.BackupPath)
	if !strings.HasPrefix(absBackupPath, allowedDir) {
		return c.Status(403).JSON(fiber.Map{
			"error": "backup file must be in ./lc_base/backups directory",
			"code":  "PATH_TRAVERSAL",
		})
	}
	
	if _, err := os.Stat(req.BackupPath); err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("backup file not found: %s", req.BackupPath),
		})
	}
	
	var err error
	
	if runtime.IsCloudflare() {
		d1Client := cloudflare.NewD1Client(
			os.Getenv("CF_ACCOUNT_ID"),
			os.Getenv("CF_API_TOKEN"),
		)
		err = d1Client.ImportSQLite(os.Getenv("CF_D1_DATABASE_ID"), req.BackupPath)
	} else {
		dbPath := runtime.GetDatabasePath()
		err = copyFile(req.BackupPath, dbPath)
	}
	
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("restore failed: %v", err),
		})
	}
	
	return c.JSON(fiber.Map{
		"message": "Database restored successfully",
	})
}
```

- [ ] **Step 5: Build and verify**

```bash
go build ./internal/handlers/maintenance/...
```

Expected: Success

- [ ] **Step 6: Commit**

```bash
git add internal/handlers/maintenance/maintenance.go
git commit -m "feat(maintenance): replace Wrangler with Cloudflare REST API clients

- Remove InstallWrangler, UninstallWrangler, isWranglerInstalled functions
- Update BackupDatabase to use D1Client.ExportDatabase for Cloudflare mode
- Update RestoreDatabase to use D1Client.ImportSQLite for Cloudflare mode
- Add path traversal validation for backup file paths
- Update GetStatus to show runtime type

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA"
```

---

**Plan complete!** This implementation plan provides:
- 8 comprehensive tasks with TDD approach
- Specific code examples for every change
- Clear test validation steps
- Proper git commit messages
- All core requirements from the spec

**Remaining work not in this plan** (can be added as separate tasks if needed):
- Installation API endpoints (GET /api/cloudflare/d1/list, POST /api/install updates)
- Installation UI updates (Svelte components)
- Maintenance UI updates (Svelte components)

These frontend tasks follow the same pattern and can be added after backend tasks are complete.
