# Cloudflare Dual-Mode Refactoring Design

**Date:** 2026-09-03  
**Status:** Draft  
**Complexity:** Medium

## Summary

Refactor mycart to support dual deployment modes (local development and Cloudflare Workers) with runtime auto-detection, enhanced installation flow for Cloudflare configuration, and upgraded file-based maintenance mode. This enables seamless deployment to both local environments and Cloudflare's edge platform while maintaining backward compatibility.

## Requirements Restatement

### Core Requirements

1. **File Structure**: Rename `Dockerfile` to `Dockerfile.cfworkers` and add `deploy:cf` command to package.json
2. **Local Dev Fix**: Fix `npm run devrun` SQLite error by ensuring `./lc_base/` directory exists
3. **Runtime Detection**: Auto-detect Cloudflare Workers environment via `CLOUDFLARE_APPLICATION_ID` or explicit `RUNTIME=cloudflare`
4. **Installation Enhancement**: Add optional Cloudflare configuration section with D1/R2 project selection/creation
5. **Maintenance Mode Upgrade**:
   - Auto-detect flag path (`/app/maintenance.flag` in Docker, `./maintenance.flag` locally)
   - Support both IP-based allowlist and authenticated admin access
   - Use Cloudflare REST API for D1/R2 operations (not Wrangler at runtime)
   - Remove Wrangler installation management from the app
6. **Cloudflare REST API**: Direct API integration for D1 database and R2 storage operations

### User Choices (from clarifying questions)

- **Database path**: `./lc_base/mycart.db` for local SQLite (Question 1: Option A)
- **Environment detection**: Check `CLOUDFLARE_APPLICATION_ID` + allow `RUNTIME=cloudflare` override (Question 2: Option A and C)
- **Installation flow**: List existing D1/R2 projects with "Create New" button (Question 3: Option B)
- **Maintenance flag path**: Auto-detect based on environment (Question 4: Option D)
- **Admin access**: Both IP-based (localhost) + authenticated admins (Question 5: Option D)
- **Cloudflare API**: Direct REST API calls, Wrangler only for deployment (Question 6: Option B)
- **Deployment flow**: Manual via `npm run deploy:cf` command (Question 7: Option A)
- **API credentials**: Environment variables only (Question 8: Option A)
- **Installation CF config**: Collect Account ID, API Token, D1/R2 selection during install (Question 9: Option B)

## Architecture Overview

### Current State

```
mycart (Go backend)
├── Database: SQLite (./lc_base/mycart.db)
├── Storage: Filesystem (./lc_base/uploads/)
├── Maintenance: .maintenance file at root
└── Deployment: Dockerfile (Cloudflare Container Workers)

Installation: email + password + domain
Maintenance UI: CLI enable/disable, admin UI for backup/restore/switch
```

### Target State

```
mycart (Go backend)
├── Database: SQLite (local) OR D1 (Cloudflare)
│   └── Auto-detect via CLOUDFLARE_APPLICATION_ID or RUNTIME env var
├── Storage: Filesystem (local) OR R2 (Cloudflare)
│   └── Follows database selection
├── Maintenance: Auto-detect flag path
│   ├── Local: ./maintenance.flag
│   └── Cloudflare: /app/maintenance.flag
└── Deployment: Dockerfile.cfworkers

Installation: email + password + domain + [optional CF config]
  └── If CF selected: Account ID, API Token, D1 selection/creation, R2 selection/creation
Maintenance UI: 
  ├── CLI enable/disable (creates flag file)
  ├── Admin access: localhost IPs + authenticated admins
  └── Backup/Restore via Cloudflare REST API (when in CF mode)
```

### Key Architectural Decisions

1. **Runtime Detection Layer** (`pkg/runtime/`): Single source of truth for "are we in Cloudflare?"
2. **No Wrangler Management**: App doesn't install/uninstall Wrangler - user handles this manually
3. **REST API for Operations**: Direct Cloudflare API calls for D1/R2 (no Wrangler at runtime)
4. **Backward Compatible**: Existing local deployments continue working unchanged
5. **Progressive Enhancement**: CF features only activate when credentials configured

### Component Interaction Flow

```
Request → Runtime Detection → Database Factory → SQLite/D1
                            → Storage Factory → Filesystem/R2
                            
Installation → Check CF Credentials → List D1/R2 → Save to .env

Maintenance Mode → Check Flag Path → Check Admin Auth → Allow/Block
                 → Backup/Restore → REST API (CF) or File Copy (local)
```

### Cloudflare Environment Variables

When running in Cloudflare Workers, these variables are automatically injected:
- `CLOUDFLARE_APPLICATION_ID`: Unique identifier of the Containers application
- `CLOUDFLARE_DURABLE_OBJECT_ID`: Durable Object instance ID
- `CLOUDFLARE_DEPLOYMENT_ID`: Auto-generated deployment hash
- `CLOUDFLARE_LOCATION`: Data center location name
- `CLOUDFLARE_COUNTRY_A2`: ISO 3166-1 Alpha 2 country code
- `CLOUDFLARE_REGION`: Geographic region string

## File Structure

### Files to Rename

```
Dockerfile → Dockerfile.cfworkers
```

**References to update**:
- `.github/workflows/` (if any CI/CD files reference Dockerfile)
- `README.md` or `docs/` (deployment instructions)
- `.dockerignore` → keep as-is (works for both)

### New Files to Create

```
pkg/
  runtime/
    detect.go              # Runtime detection helper
    detect_test.go         # Tests for detection logic

internal/
  handlers/
    cloudflare/
      d1.go               # D1 REST API client
      r2.go               # R2 REST API client
      cloudflare_test.go  # API client tests
```

### Files to Modify

```
cmd/
  main.go                 # Update maintenance command (flag path)

internal/
  config/
    config.go             # Add CF credentials loading
  
  middleware/
    maintenance.go        # Auto-detect flag path, add admin auth check
  
  handlers/
    maintenance/
      maintenance.go      # Remove Wrangler install/uninstall
                         # Add REST API calls for D1/R2
    install.go           # Add CF configuration handling
  
  app.go                 # Ensure ./lc_base/ creation on startup

pkg/
  database/
    factory.go           # Use runtime detection for auto-selection
  
  storage/
    factory.go           # Use runtime detection for auto-selection

web/admin/src/routes/
  install/
    +page.svelte         # Add CF configuration fields
  _/maintenance/
    +page.svelte         # Remove Wrangler section, update status display

package.json            # Add deploy:cf command

.env.example            # Add CF_ACCOUNT_ID, CF_API_TOKEN, etc.
```

### Configuration Changes (.env.example)

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

## Components

### 1. Runtime Detection (`pkg/runtime/detect.go`)

**Responsibilities**:
- Detect if app is running in Cloudflare Workers environment
- Determine appropriate maintenance flag path
- Provide runtime-specific configuration

**Key Functions**:
```go
func IsCloudflare() bool
func GetMaintenanceFlagPath() string
func GetDatabasePath() string
```

**Detection Logic**:
1. Check for explicit `RUNTIME=cloudflare` environment variable (highest priority)
2. Check for `CLOUDFLARE_APPLICATION_ID` environment variable (auto-detect)
3. Default to local mode if neither is present

**Path Detection**:
1. Try `/app` directory (Docker/Cloudflare) - check if writable
2. Fallback to `./maintenance.flag` (local project root)

### 2. Cloudflare D1 Client (`internal/handlers/cloudflare/d1.go`)

**Responsibilities**:
- List D1 databases in account
- Create new D1 databases
- Export D1 database to SQLite file
- Import SQLite file to D1 database
- Execute SQL statements

**API Endpoints Used**:
- `GET /accounts/:account_id/d1/database` - List databases
- `POST /accounts/:account_id/d1/database` - Create database
- `POST /accounts/:account_id/d1/database/:database_id/export` - Export
- `POST /accounts/:account_id/d1/database/:database_id/query` - Execute SQL

**Authentication**: Bearer token in `Authorization` header

**Timeout**: 5 minutes for export/import operations

### 3. Cloudflare R2 Client (`internal/handlers/cloudflare/r2.go`)

**Responsibilities**:
- List R2 buckets in account
- Create new R2 buckets
- List objects in bucket

**API Endpoints Used**:
- `GET /accounts/:account_id/r2/buckets` - List buckets
- `POST /accounts/:account_id/r2/buckets` - Create bucket
- `GET /accounts/:account_id/r2/buckets/:bucket_name/objects` - List objects

**Authentication**: Bearer token in `Authorization` header

### 4. Updated Maintenance Middleware

**File**: `internal/middleware/maintenance.go`

**New Behavior**:
- Use `runtime.GetMaintenanceFlagPath()` instead of hardcoded `.maintenance`
- Check both IP allowlist AND authenticated admin sessions
- Return 503 for unauthorized requests during maintenance

**Authorization Logic**:
1. Check IP allowlist (localhost by default, or `MAINTENANCE_ALLOWED_IPS`)
2. Check for authenticated admin session (JWT token or session cookie)
3. Allow request if either check passes

### 5. Enhanced Installation Flow

**Frontend**: `web/admin/src/routes/install/+page.svelte`

**New Fields**:
- Deployment Target: `<select>` with "Local Development" / "Cloudflare Workers"
- Cloudflare Account ID: `<input type="text">`
- Cloudflare API Token: `<input type="password">`
- D1 Database: `<select>` with existing databases + "Create New"
- R2 Bucket: `<select>` with existing buckets + "Create New"

**UX Flow**:
1. User enters basic fields (email, password, domain)
2. User selects deployment target
3. If Cloudflare selected:
   - User enters Account ID and API Token
   - On blur, frontend fetches D1/R2 resources via API
   - User selects existing or creates new D1 database
   - User selects existing or creates new R2 bucket
4. Submit creates admin user and writes `.env` file

**Backend**: `internal/handlers/install.go`

**New Request Fields**:
```go
type InstallRequest struct {
    Email          string
    Password       string
    Domain         string
    DeployTarget   string // "local" or "cloudflare"
    CFAccountID    string
    CFAPIToken     string
    CFD1DatabaseID string
    CFR2BucketName string
}
```

**Logic**:
- Validate all fields
- Create admin user (existing logic)
- Write `.env` file with appropriate configuration based on `DeployTarget`
- If Cloudflare: set `DB_TYPE=d1`, `STORAGE_TYPE=r2`, and CF credentials
- If local: set `DB_TYPE=sqlite`, `STORAGE_TYPE=filesystem`, and local paths

### 6. Updated Database/Storage Factories

**Database Factory** (`pkg/database/factory.go`):
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
    
    // Create appropriate database instance
}
```

**Storage Factory** (`pkg/storage/factory.go`):
- Similar auto-detection logic
- Choose between Filesystem and R2 based on runtime

### 7. Local Development Setup

**File**: `internal/app.go`

**New Initialization**:
```go
func NewApp(...) error {
    // Ensure local directories exist
    if !runtime.IsCloudflare() {
        os.MkdirAll("./lc_base", 0755)
        os.MkdirAll("./lc_base/uploads", 0755)
        os.MkdirAll("./lc_base/backups", 0755)
    }
    
    // Rest of app initialization
}
```

## Data Flow

### Installation Flow

```
1. User visits /install page
2. User enters email, password, domain
3. User selects "Cloudflare Workers" deployment target
4. Cloudflare config section appears
5. User enters CF Account ID and API Token
6. Frontend triggers: GET /api/cloudflare/d1/list (with credentials in headers)
7. Backend calls Cloudflare API, returns D1 databases
8. Frontend triggers: GET /api/cloudflare/r2/list
9. Backend calls Cloudflare API, returns R2 buckets
10. User selects D1 database (or "Create New")
11. User selects R2 bucket (or "Create New")
12. User clicks Submit
13. Backend validates and creates admin user
14. Backend writes .env file with all configuration
15. Backend returns success
16. Frontend redirects to /signin
```

### Runtime Detection Flow

```
1. App starts
2. runtime.IsCloudflare() checks environment
   → Check RUNTIME env var first
   → Check CLOUDFLARE_APPLICATION_ID
   → Return true/false
3. Database factory uses runtime detection
   → If Cloudflare: NewD1(credentials)
   → If local: NewSQLite(./lc_base/mycart.db)
4. Storage factory uses runtime detection
   → If Cloudflare: NewR2(credentials)
   → If local: NewFilesystem(./lc_base/uploads)
5. App runs with appropriate backend
```

### Maintenance Mode Flow

```
Enable:
1. Admin runs: mycart maintenance enable
2. CLI calls runtime.GetMaintenanceFlagPath()
   → Returns /app/maintenance.flag (Docker) or ./maintenance.flag (local)
3. CLI creates flag file
4. CLI prints success message with access instructions

Request Handling:
1. Request arrives
2. Middleware checks if flag file exists
3. If not in maintenance mode → c.Next()
4. If in maintenance mode:
   → Check IP allowlist (localhost by default)
   → Check authenticated admin session
   → If authorized → c.Next()
   → If not authorized → 503 Service Unavailable

Disable:
1. Admin runs: mycart maintenance disable
2. CLI removes flag file
3. All users can access again
```

### Backup/Restore Flow (Cloudflare)

```
Backup:
1. Admin clicks "Create Backup" in /_/maintenance UI
2. Frontend: POST /_/api/maintenance/backup
3. Backend:
   → Check runtime.IsCloudflare()
   → Create D1Client with credentials from env
   → Call client.ExportDatabase(databaseID, outputPath)
   → D1Client: POST /accounts/:id/d1/database/:db_id/export
   → Cloudflare API returns SQLite file
   → Write to ./lc_base/backups/backup-d1-TIMESTAMP.db
   → Return backup path
4. Frontend displays success + backup path

Restore:
1. Admin enters backup path, clicks "Restore"
2. Frontend: POST /_/api/maintenance/restore
   → Body: {"backup_path": "./lc_base/backups/backup-d1-20260903.db"}
3. Backend:
   → Validate path is in allowed directory
   → Check file exists
   → Check runtime.IsCloudflare()
   → Create D1Client
   → Call client.ImportSQLite(databaseID, backupPath)
   → D1Client: POST /accounts/:id/d1/database/:db_id/query
   → Send SQL statements from backup file
   → Return success
4. Frontend displays success message
```

## Error Handling

### Error Categories

1. **Client Errors (4xx)**:
   - Invalid request body → 400 Bad Request
   - Missing backup file → 404 Not Found
   - Path traversal attempt → 403 Forbidden
   - Unauthorized access → 401 Unauthorized

2. **Server Errors (5xx)**:
   - Database connection errors → 500 Internal Server Error
   - File system errors → 500 Internal Server Error
   - Cloudflare API failures → 500 Internal Server Error
   - Timeout errors → 504 Gateway Timeout

### Error Response Format

All API errors return consistent JSON:
```json
{
  "error": "Human-readable error message",
  "details": "Additional error context (optional)",
  "code": "ERROR_CODE"
}
```

### Cloudflare API Error Handling

- Capture HTTP status code and response body
- Log full error details (account ID, database ID, operation)
- Return user-friendly error message
- Include API response in `details` field for debugging

### Path Traversal Prevention

**Validation**:
```go
allowedDir, _ := filepath.Abs("./lc_base/backups")
absBackupPath, _ := filepath.Abs(req.BackupPath)
if !strings.HasPrefix(absBackupPath, allowedDir) {
    return c.Status(403).JSON(fiber.Map{
        "error": "backup file must be in ./lc_base/backups directory",
        "code": "PATH_TRAVERSAL"
    })
}
```

### Logging Strategy

Use structured logging with `zerolog`:
```go
log.Info().
    Str("runtime", runtime.GetRuntime()).
    Str("operation", "backup").
    Str("database_type", dbType).
    Msg("backup started")
```

## Testing Strategy

### Unit Tests

**Runtime Detection** (`pkg/runtime/detect_test.go`):
- Test `IsCloudflare()` with various environment variable combinations
- Test `GetMaintenanceFlagPath()` with writable/non-writable `/app`
- Test `GetDatabasePath()` for local vs Cloudflare

**Cloudflare Clients** (`internal/handlers/cloudflare/`):
- Mock HTTP server for API responses
- Test successful API calls
- Test error handling (401, 500, timeouts)
- Test JSON parsing and data structures

**Middleware** (`internal/middleware/maintenance_test.go`):
- Test maintenance mode detection with different flag paths
- Test IP allowlist parsing and matching
- Test admin authentication check
- Test combined authorization (IP + auth)

### Integration Tests

**Backup/Restore Cycle**:
1. Create test SQLite database
2. Call backup endpoint
3. Verify backup file created
4. Modify database
5. Call restore endpoint
6. Verify data restored to original state

**Installation Flow**:
1. POST to `/api/install` with local deployment
2. Verify `.env` file created with correct settings
3. POST to `/api/install` with Cloudflare deployment
4. Verify `.env` includes CF credentials

### E2E Tests (Playwright)

**Local Installation**:
- Fill email, password, domain
- Select "Local Development"
- Submit
- Verify redirect to /signin
- Verify app starts successfully

**Cloudflare Installation**:
- Fill email, password, domain
- Select "Cloudflare Workers"
- Enter CF credentials
- Wait for D1/R2 lists to load
- Select or create resources
- Submit
- Verify redirect to /signin

**Maintenance Mode**:
- Enable via CLI
- Access from non-localhost → verify 503
- Login as admin
- Access /_/maintenance → verify dashboard loads
- Create backup
- Restore from backup
- Disable via CLI

### Manual Testing Checklist

**Local Development**:
- [ ] Run `npm run devrun` - should create ./lc_base/ directory
- [ ] Database file created at ./lc_base/mycart.db
- [ ] Installation page loads
- [ ] Install with local deployment works
- [ ] `mycart maintenance enable` creates ./maintenance.flag
- [ ] Access blocked during maintenance (non-localhost)
- [ ] Admin can access /_/maintenance
- [ ] Backup creates file in ./lc_base/backups/
- [ ] Restore from backup works

**Cloudflare Workers** (requires CF account):
- [ ] Set CLOUDFLARE_APPLICATION_ID environment variable
- [ ] Installation shows Cloudflare option
- [ ] CF credentials validation works
- [ ] D1 database list loads
- [ ] R2 bucket list loads
- [ ] Create new D1 database works
- [ ] Create new R2 bucket works
- [ ] `npm run deploy:cf` deploys to Cloudflare
- [ ] App detects Cloudflare runtime
- [ ] D1 backup via REST API works
- [ ] D1 restore via REST API works

## Security Considerations

### API Credentials Storage

- Store CF credentials in environment variables only
- Never commit credentials to git
- Use `.env.example` as template with empty values
- Validate credentials before saving to `.env`

### Path Traversal Prevention

- Validate all file paths are within `./lc_base/backups/`
- Use `filepath.Abs()` and `strings.HasPrefix()` for validation
- Return 403 Forbidden for invalid paths
- Log all path validation failures

### Authentication

- Maintenance mode requires localhost IP OR admin session
- Admin session validation via JWT token or session cookie
- No separate maintenance admin credentials needed
- Reuse existing admin authentication infrastructure

### Cloudflare API Security

- Never log API tokens
- Use HTTPS for all API requests
- Validate API responses before processing
- Set reasonable timeouts (5 minutes)
- Handle rate limiting gracefully

## Deployment Considerations

### Local Development

**First-Time Setup**:
1. Clone repository
2. Run `npm install`
3. Run `npm run devrun`
4. Visit http://localhost:8080/install
5. Choose "Local Development"
6. Complete installation

**Database Location**: `./lc_base/mycart.db`

**Backups**: `./lc_base/backups/`

### Cloudflare Workers

**Prerequisites**:
- Cloudflare account with Workers enabled
- Wrangler CLI installed (`npm install -g wrangler`)
- Authenticated with Cloudflare (`wrangler login`)

**First-Time Setup**:
1. Clone repository
2. Run `npm install`
3. Build: `npm run build`
4. Visit local install page: http://localhost:8080/install
5. Choose "Cloudflare Workers"
6. Enter CF Account ID and API Token
7. Select or create D1 database and R2 bucket
8. Complete installation
9. Deploy: `npm run deploy:cf`

**Environment Variables** (set in Cloudflare dashboard or wrangler.toml):
- `CF_ACCOUNT_ID`
- `CF_API_TOKEN`
- `CF_D1_DATABASE_ID`
- `CF_R2_BUCKET_NAME`

**Deployment Command**: `wrangler deploy` (aliased as `npm run deploy:cf`)

### Migration from Local to Cloudflare

1. Create backup on local: `mycart maintenance enable` → backup via UI
2. Set up Cloudflare resources (D1 database, R2 bucket)
3. Update `.env` with CF credentials
4. Restore backup to D1 via maintenance UI
5. Deploy to Cloudflare: `npm run deploy:cf`
6. Verify deployment works
7. Update DNS to point to Cloudflare

## Future Enhancements

### Potential Improvements

1. **Automatic Backup Sync**: Sync backups to R2 automatically
2. **Multi-Region Deployment**: Deploy to multiple Cloudflare regions
3. **Backup Encryption**: Encrypt backups at rest
4. **Database Migration Tool**: Automated data migration between SQLite and D1
5. **Health Checks**: Endpoint for deployment health monitoring
6. **Rollback Support**: Quick rollback to previous deployment
7. **Configuration Validation**: Pre-deployment validation of CF credentials
8. **Cost Monitoring**: Track Cloudflare usage and costs

### Out of Scope (for this iteration)

- Automatic database synchronization between SQLite and D1
- Live migration without downtime
- Multi-admin coordination
- Backup compression
- Incremental backups
- Backup retention policies

## Dependencies

### Required

- Go 1.26+ (existing)
- Fiber v3 (existing)
- Node.js 24+ (for Svelte build)
- Cloudflare account (for CF deployment only)

### Optional

- Wrangler CLI (for Cloudflare deployment)

### External APIs

- Cloudflare REST API v4
  - D1 Database API: https://developers.cloudflare.com/api/resources/d1/
  - R2 Storage API: https://developers.cloudflare.com/api/resources/r2/

## Acceptance Criteria

### Functionality

- [ ] `Dockerfile` renamed to `Dockerfile.cfworkers`
- [ ] `npm run devrun` creates `./lc_base/` and runs successfully
- [ ] Runtime detection correctly identifies Cloudflare vs local
- [ ] Installation page shows Cloudflare configuration option
- [ ] D1/R2 resource listing works with valid credentials
- [ ] Creating new D1 database works
- [ ] Creating new R2 bucket works
- [ ] Installation saves correct `.env` for local deployment
- [ ] Installation saves correct `.env` for Cloudflare deployment
- [ ] Maintenance flag path auto-detects correctly
- [ ] IP allowlist works for maintenance mode
- [ ] Authenticated admin access works during maintenance
- [ ] Backup works for local SQLite
- [ ] Backup works for Cloudflare D1 (via REST API)
- [ ] Restore works for local SQLite
- [ ] Restore works for Cloudflare D1 (via REST API)
- [ ] `npm run deploy:cf` command exists
- [ ] Path traversal attacks rejected

### Performance

- [ ] Runtime detection adds <1ms overhead
- [ ] Cloudflare API calls timeout after 5 minutes
- [ ] Backup/restore operations log progress

### Security

- [ ] API credentials never logged
- [ ] Path traversal prevented
- [ ] Admin authentication required for maintenance UI
- [ ] HTTPS used for all Cloudflare API calls

### User Experience

- [ ] Clear error messages for all failure scenarios
- [ ] Installation flow is intuitive
- [ ] Maintenance mode explains how to regain access
- [ ] CLI commands provide helpful output

### Testing

- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] E2E tests pass
- [ ] Manual testing checklist completed

## References

### Existing Code Patterns

- **Runtime detection**: Similar to `pkg/database/factory.go` and `pkg/storage/factory.go`
- **Middleware pattern**: `internal/middleware/` (CORS, logging, etc.)
- **API client pattern**: `pkg/database/d1.go` (existing D1 client)
- **Svelte installation**: `web/admin/src/routes/install/+page.svelte`
- **CLI commands**: `cmd/main.go` (serve, install, migrate)

### External Documentation

- [Cloudflare D1 API](https://developers.cloudflare.com/api/resources/d1/)
- [Cloudflare R2 API](https://developers.cloudflare.com/api/resources/r2/)
- [Cloudflare Container Workers](https://developers.cloudflare.com/workers/runtime-apis/container/)
- [Fiber v3 Middleware](https://docs.gofiber.io/guide/middleware/)
- [Svelte Forms](https://svelte.dev/tutorial/form-bindings)

### Design Decisions Log

1. **Why auto-detect runtime instead of explicit config?**
   - Simpler deployment (no manual config changes)
   - Works with Cloudflare's injected environment variables
   - Still allows manual override via `RUNTIME` env var

2. **Why remove Wrangler installation management?**
   - Wrangler is a dev tool, should be installed globally
   - App shouldn't manage its own deployment tools
   - Simplifies codebase (remove install/uninstall handlers)

3. **Why REST API instead of Wrangler at runtime?**
   - Direct control over operations
   - No dependency on Node.js/npm at runtime
   - Better error handling and logging
   - Wrangler is for deployment, not runtime operations

4. **Why auto-detect maintenance flag path?**
   - Works seamlessly in both Docker and local
   - No manual configuration needed
   - Fallback ensures it always works

5. **Why optional Cloudflare config during installation?**
   - Users can start locally and migrate later
   - Not everyone needs Cloudflare deployment
   - Simplifies local development setup

---

**End of Design Document**
