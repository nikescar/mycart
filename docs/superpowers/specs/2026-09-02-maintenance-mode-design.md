# Maintenance Mode Feature Design

**Date:** 2026-09-02  
**Status:** Draft  
**Complexity:** Medium

## Summary

Add a maintenance mode feature to myCart that allows administrators to safely perform database operations (backup, restore, switching between SQLite and Cloudflare D1/R2) while blocking public access. Maintenance mode is triggered by a `.maintenance` file in the project root, enforces IP-based access restrictions, and integrates with the existing graceful shutdown mechanism.

## Requirements Restatement

### Core Requirements
1. **File-based toggle**: Touching `.maintenance` file enables maintenance mode (persistent across restarts)
2. **Access control**: Only localhost (or configured IPs) can access the application in maintenance mode
3. **Database operations**: Admin can switch between SQLite ↔ D1 and perform backup/restore operations
4. **Storage operations**: Implicit switching between Filesystem ↔ R2 when database switches
5. **CLI-only enable/disable**: Maintenance mode can only be enabled/disabled via terminal/SSH
6. **UI-driven operations**: Database backup/restore/switch operations accessible through Svelte admin UI
7. **Wrangler integration**: Backend invokes wrangler CLI for Cloudflare D1/R2 operations
8. **Graceful shutdown**: Integrate with existing Fiber graceful shutdown when restarting into maintenance mode

### User Choices (from clarifying questions)
- **State management**: Persistent via `.maintenance` file (Option A)
- **Access control**: IP-restricted to localhost/specific IPs (Option C)
- **Database switching**: Manual backup/restore (two separate operations) (Option B)
- **Backup format**: Binary SQLite format for efficiency (Option C)
- **Wrangler integration**: Shell execution via `exec.Command` with install/uninstall support (Option A)
- **Graceful shutdown**: Standard `app.Shutdown()` with default timeout (Option A)

## Architecture Overview

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ Request arrives                                              │
└─────────────┬───────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│ Maintenance Middleware (internal/middleware/maintenance.go) │
│  • Check .maintenance file exists                           │
│  • Check request IP against MAINTENANCE_ALLOWED_IPS          │
│  • If blocked: return 503 with maintenance page              │
│  • If allowed: c.Next()                                      │
└─────────────┬───────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────────────┐
│ Route Handler                                                │
│  • Normal routes: /_/, /api/*, /                            │
│  • Maintenance routes: /_/maintenance (API + UI)            │
└─────────────────────────────────────────────────────────────┘
```

### Key Components

1. **Maintenance Mode State**: `.maintenance` file in project root
2. **IP Allowlist**: `MAINTENANCE_ALLOWED_IPS` environment variable (defaults to `127.0.0.1,::1`)
3. **Middleware**: `internal/middleware/maintenance.go` - intercepts all requests
4. **CLI Commands**: `cmd/main.go` - adds `mycart maintenance enable|disable|status`
5. **API Handlers**: `internal/handlers/maintenance/` - backup/restore/switch/wrangler operations
6. **UI**: `web/admin/src/routes/_/maintenance/+page.svelte` - maintenance control panel
7. **Settings UI**: `web/admin/src/routes/_/settings/maintenance/+page.svelte` - restart button

### Data Flow

#### Enabling Maintenance Mode
```
Admin (SSH/Terminal)
  → mycart maintenance enable [--restart]
  → Creates .maintenance file
  → (Optional) Sends SIGTERM for graceful restart
  → Server restarts in maintenance mode
  → Middleware blocks all non-localhost requests
```

#### Database Backup Flow
```
Admin (localhost browser)
  → /_/maintenance UI
  → Click "Create Backup" button
  → POST /_/api/maintenance/backup
  → Handler checks database type:
     - SQLite: Copy .db file to backups/
     - D1: Execute `wrangler d1 export` to SQLite file
  → Return backup file path
  → UI displays backup path
```

#### Database Restore Flow
```
Admin (localhost browser)
  → /_/maintenance UI
  → Enter backup path, click "Restore" button
  → POST /_/api/maintenance/restore
  → Handler validates backup file exists and is in allowed directory
  → Handler checks database type:
     - SQLite: Copy backup file to active database
     - D1: Execute `wrangler d1 execute` with SQLite file
  → Return success message
```

#### Database Switch Flow
```
Admin (localhost browser)
  → /_/maintenance UI
  → Click "Switch to D1" or "Switch to SQLite"
  → POST /_/api/maintenance/switch
  → Handler updates environment variables (runtime only)
  → Returns message to update .env file manually
  → Admin restarts server to apply changes
```

## File Structure

### New Files

```
internal/
  middleware/
    maintenance.go          # Maintenance mode middleware
  handlers/
    maintenance/
      maintenance.go        # API handlers for maintenance operations

cmd/
  main.go                   # Add maintenance CLI commands (modify existing)

web/admin/src/routes/_/
  maintenance/
    +page.svelte           # Maintenance control panel UI
  settings/maintenance/
    +page.svelte           # Settings page with restart button

docs/superpowers/specs/
  2026-09-02-maintenance-mode-design.md  # This design document
```

### Modified Files

```
internal/
  app.go                    # Integrate maintenance middleware
  routes/routes.go          # Add maintenance API routes

cmd/
  main.go                   # Add cmdMaintenance() command
```

## Components

### 1. Maintenance Middleware

**File**: `internal/middleware/maintenance.go`

**Responsibilities**:
- Check `.maintenance` file existence on every request
- Parse `MAINTENANCE_ALLOWED_IPS` environment variable
- Block requests from non-allowed IPs with 503 status
- Allow requests from allowed IPs to proceed normally

**Key functions**:
- `MaintenanceMode() fiber.Handler` - Main middleware handler
- `isMaintenanceMode() bool` - Check if `.maintenance` file exists
- `loadAllowedIPs() []string` - Parse environment variable (default: `127.0.0.1,::1`)
- `isIPAllowed(clientIP string, allowedIPs []string) bool` - Check if IP is in allowlist

**Performance considerations**:
- `os.Stat()` on `.maintenance` file is fast (cached by OS)
- IP allowlist is loaded once at middleware creation time
- String comparison for IP matching is O(n) where n is number of allowed IPs (typically 1-5)

### 2. CLI Commands

**File**: `cmd/main.go` (modifications)

**Commands**:
- `mycart maintenance enable [--restart]` - Create `.maintenance` file, optionally trigger restart
- `mycart maintenance disable` - Remove `.maintenance` file
- `mycart maintenance status` - Check if maintenance mode is active

**Behavior**:
- `enable`: Creates `.maintenance` file with timestamp, prints access instructions
- `enable --restart`: Additionally sends SIGTERM to trigger graceful shutdown/restart
- `disable`: Removes `.maintenance` file, prints confirmation
- `status`: Checks file existence and prints current state

### 3. API Handlers

**File**: `internal/handlers/maintenance/maintenance.go`

**Endpoints**:
- `GET /_/api/maintenance/status` - Get current database/storage type and wrangler status
- `POST /_/api/maintenance/backup` - Create database backup
- `POST /_/api/maintenance/restore` - Restore from backup file
- `POST /_/api/maintenance/switch` - Switch database type (requires restart)
- `POST /_/api/maintenance/wrangler/install` - Install wrangler via npm
- `POST /_/api/maintenance/wrangler/uninstall` - Uninstall wrangler
- `POST /_/api/maintenance/enable-and-restart` - Enable maintenance mode and restart server

**Backup logic**:
- SQLite: Copy `.db` file to `./lc_base/backups/backup-TIMESTAMP.db`
- D1: Execute `npx wrangler d1 export <database_id> --output <backup_path>`

**Restore logic**:
- SQLite: Copy backup file to active database path
- D1: Execute `npx wrangler d1 execute <database_id> --file <backup_path>`

**Switch logic**:
- Update `DB_TYPE` environment variable (runtime only)
- Set/unset `CF_WORKER` environment variable
- Return message to update `.env` file manually for persistence

**Error handling**:
- Validate backup paths are within `./lc_base/backups/` directory (security)
- Timeout wrangler commands after 5 minutes
- Return detailed error messages with wrangler output
- Check file existence before operations

### 4. Frontend UI

**Maintenance Panel**: `web/admin/src/routes/_/maintenance/+page.svelte`

**Features**:
- Status display: Current database type, storage type, wrangler installation status
- Backup section: "Create Backup" button, displays last backup path
- Restore section: "Restore from Backup" button (disabled until backup created)
- Switch section: "Switch to SQLite" / "Switch to D1" buttons (disabled if already on that type)
- Wrangler section: Install/Uninstall buttons based on current state
- Exit section: Reminder to use CLI command to exit maintenance mode

**State management**:
- Load status on mount via `GET /_/api/maintenance/status`
- Update status after operations complete
- Show loading states on buttons during operations
- Display toast notifications for success/error

**Settings Page**: `web/admin/src/routes/_/settings/maintenance/+page.svelte`

**Features**:
- "Enable Maintenance Mode & Restart" button
- Help text explaining what happens during restart
- Redirects to `/_/maintenance` after server restarts

### 5. Wrangler Integration

**Installation**:
- Execute `npm i -D wrangler@latest` via `exec.Command`
- Check installation status via `npx wrangler --version`

**D1 Operations**:
- Export: `npx wrangler d1 export <database_id> --output <path>`
- Import: `npx wrangler d1 execute <database_id> --file <path>`

**Timeout handling**:
- All wrangler commands have 5-minute timeout via `context.WithTimeout`
- Return timeout error if exceeded

**Error handling**:
- Capture stdout/stderr from wrangler commands
- Return detailed error messages with command output
- Check for file creation after export operation

### 6. Graceful Shutdown Integration

**Existing infrastructure** (in `internal/app.go`):
- `handleShutdown()` - Listens for SIGTERM, calls `app.Shutdown()`
- `app.Shutdown()` - Stops accepting new connections, waits for active requests to complete

**New integration**:
- Settings UI triggers `POST /_/api/maintenance/enable-and-restart`
- Handler creates `.maintenance` file, then sends SIGTERM to self
- `handleShutdown()` executes graceful shutdown
- Process manager (systemd/supervisor/docker) restarts server
- Server starts in maintenance mode (`.maintenance` file exists)

**Timeout**: Uses existing default timeout from Fiber (no changes needed)

## Error Handling

### Error Categories

1. **User errors** (400 Bad Request)
   - Invalid request body
   - Missing backup file
   - Invalid database type

2. **Permission errors** (403 Forbidden)
   - Backup path outside allowed directory (path traversal attempt)

3. **Not found errors** (404 Not Found)
   - Backup file doesn't exist

4. **External tool errors** (500 Internal Server Error)
   - Wrangler execution failures
   - File system errors
   - Database connection errors

5. **Timeout errors** (500 Internal Server Error)
   - Wrangler operations exceeding 5-minute timeout

### Error Response Format

All API errors return JSON:
```json
{
  "error": "Human-readable error message",
  "output": "Command output (for wrangler errors)"
}
```

### Frontend Error Handling

- Network errors: Show "Network error: Check server connection"
- HTTP errors: Parse JSON error field and display in toast
- Timeout errors: Show "Operation timed out" message
- Log all errors to console for debugging

### CLI Error Handling

- Check file existence before operations
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Exit with status code 1 on errors
- Print warnings for non-critical errors

## Testing Strategy

### Unit Tests

**Middleware tests** (`internal/middleware/maintenance_test.go`):
- Maintenance mode off allows all requests
- Maintenance mode on blocks external IPs
- Maintenance mode on allows localhost
- Maintenance mode on allows custom IPs from environment variable
- Invalid IP format handling

**CLI tests** (`cmd/main_test.go`):
- `enable` creates `.maintenance` file
- `enable` with existing file shows warning
- `disable` removes `.maintenance` file
- `disable` with no file shows message
- `status` correctly reports state

**API handler tests** (`internal/handlers/maintenance/maintenance_test.go`):
- Backup SQLite database succeeds
- Backup D1 database calls wrangler correctly
- Restore validates backup path
- Restore rejects path traversal attempts
- Switch updates environment variables
- Wrangler install/uninstall execute correct commands

### Integration Tests

**Full backup/restore cycle**:
1. Create test database with sample data
2. Call backup endpoint
3. Verify backup file created
4. Modify database
5. Call restore endpoint
6. Verify data restored to original state

**Database switching**:
1. Start with SQLite
2. Call switch endpoint to D1
3. Verify environment variables updated
4. Restart server (mocked)
5. Verify new database type active

### E2E Tests

**Maintenance mode workflow**:
1. Enable maintenance mode via CLI
2. Visit app from external IP - verify 503 response
3. Access from localhost - verify maintenance UI loads
4. Create backup via UI
5. Verify backup file created on filesystem
6. Restore from backup via UI
7. Disable maintenance mode via CLI
8. Verify app accessible to all users

**Restart workflow**:
1. Access settings page from localhost
2. Click "Enable Maintenance Mode & Restart" button
3. Verify `.maintenance` file created
4. Verify graceful shutdown triggered
5. Verify server restarts in maintenance mode
6. Access `/_/maintenance` from localhost

### Manual Testing Checklist

- [ ] Enable maintenance mode via `mycart maintenance enable`
- [ ] Verify external requests blocked with 503
- [ ] Verify localhost requests allowed
- [ ] Access `/_/maintenance` from localhost
- [ ] Check status display shows correct database/storage types
- [ ] Create SQLite backup
- [ ] Verify backup file exists in `./lc_base/backups/`
- [ ] Restore from backup
- [ ] Install wrangler via UI
- [ ] Verify wrangler installed with `npx wrangler --version`
- [ ] Switch to D1 (requires Cloudflare credentials configured)
- [ ] Export D1 to SQLite backup
- [ ] Import SQLite to D1
- [ ] Uninstall wrangler via UI
- [ ] Use restart button in settings
- [ ] Verify graceful shutdown occurs
- [ ] Verify server restarts in maintenance mode
- [ ] Disable maintenance mode via `mycart maintenance disable`
- [ ] Verify app accessible to all users

## Security Considerations

### Path Traversal Prevention

**Issue**: User-provided backup paths could reference files outside the allowed directory.

**Mitigation**:
```go
// Validate backup file is in allowed directory
allowedDir, _ := filepath.Abs("./lc_base/backups")
absBackupPath, _ := filepath.Abs(req.BackupPath)
if !strings.HasPrefix(absBackupPath, allowedDir) {
    return fiber.StatusForbidden, "backup file must be in ./lc_base/backups directory"
}
```

### IP Spoofing Prevention

**Issue**: X-Forwarded-For header can be spoofed by clients.

**Mitigation**:
- Fiber's `TrustProxy` config is already set to trust only loopback/private proxies
- `c.IP()` returns the real client IP, ignoring untrusted headers
- For extra security, maintenance mode typically used on servers without public load balancers

### Command Injection Prevention

**Issue**: Wrangler commands execute shell commands.

**Mitigation**:
- Use `exec.Command()` with separate arguments (not shell execution)
- Database IDs and paths are from config/filesystem, not user input
- Backup paths validated to be within allowed directory

### Environment Variable Exposure

**Issue**: Wrangler requires Cloudflare credentials in environment variables.

**Mitigation**:
- Credentials loaded from `.env` file (not committed to git)
- API responses don't include credential values
- Status endpoint shows only database type, not connection strings

## Deployment Considerations

### First-Time Setup

1. Ensure process manager (systemd/supervisor/docker) is configured to auto-restart
2. Set `MAINTENANCE_ALLOWED_IPS` if accessing from non-localhost
3. For Cloudflare operations, configure D1/R2 credentials in `.env`
4. Create `./lc_base/backups/` directory (auto-created on first backup)

### Production Deployment

1. Test maintenance mode on staging environment first
2. Schedule maintenance window for database operations
3. Enable maintenance mode via SSH: `mycart maintenance enable`
4. Perform backup before any database switch
5. Test backup file integrity before proceeding
6. Switch database type and restart: `mycart maintenance enable --restart`
7. Verify application health after restart
8. Disable maintenance mode: `mycart maintenance disable`

### Monitoring

**Metrics to track**:
- Maintenance mode enable/disable events (log to audit trail)
- Backup operation duration
- Restore operation duration
- Wrangler operation failures
- Graceful shutdown duration

**Alerts**:
- Maintenance mode enabled for >1 hour (possible admin forgot to disable)
- Wrangler operations failing repeatedly
- Backup files not created in expected directory

## Future Enhancements

### Potential Improvements

1. **Scheduled backups**: Cron job to automatically backup database daily
2. **Backup retention**: Automatically delete backups older than N days
3. **Database migration**: Automatic data migration when switching databases
4. **Multi-region support**: Handle database replication across regions
5. **Backup compression**: Compress backup files to save disk space
6. **Web-based file browser**: Browse and select backup files from UI
7. **Backup verification**: Automatically verify backup integrity after creation
8. **Rollback capability**: Quick rollback to previous database state if restore fails
9. **Audit log**: Track all maintenance operations with timestamps and admin users
10. **Email notifications**: Send email when entering/exiting maintenance mode

### Out of Scope (for this iteration)

- Automatic database synchronization between SQLite and D1
- Live migration without downtime
- Backup encryption at rest
- Multi-admin coordination (locking mechanism)
- Progress bars for long-running operations
- Backup diff/comparison tools

## Dependencies

### Required

- Go 1.21+ (existing)
- Fiber v3 (existing)
- Cobra CLI (existing)
- Database abstraction layer (existing: `pkg/database`)
- Storage abstraction layer (existing: `pkg/storage`)

### Optional

- Wrangler CLI (installed via npm on demand for Cloudflare operations)
- Cloudflare account with D1/R2 configured (for production Cloudflare features)

### Development

- Testing frameworks (existing: `testing`, `httptest`)
- Playwright (existing: for E2E tests)

## Acceptance Criteria

### Functionality

- [ ] CLI commands `enable`, `disable`, `status` work correctly
- [ ] Middleware blocks external IPs when maintenance mode enabled
- [ ] Middleware allows localhost when maintenance mode enabled
- [ ] Maintenance UI loads from localhost in maintenance mode
- [ ] Backup creates SQLite file in `./lc_base/backups/`
- [ ] Restore loads data from backup file
- [ ] Switch updates environment variables correctly
- [ ] Wrangler install/uninstall execute npm commands
- [ ] Restart button triggers graceful shutdown
- [ ] Server restarts in maintenance mode when `.maintenance` file exists

### Performance

- [ ] Middleware adds <1ms latency per request
- [ ] Backup operation completes in <30s for 100MB database
- [ ] Restore operation completes in <30s for 100MB database
- [ ] Graceful shutdown completes within default timeout

### Security

- [ ] Path traversal attacks rejected (403 Forbidden)
- [ ] External IPs blocked in maintenance mode (503 Service Unavailable)
- [ ] Wrangler commands use safe `exec.Command()` (no shell injection)
- [ ] Environment variables not exposed in API responses

### User Experience

- [ ] Clear error messages for all failure scenarios
- [ ] Toast notifications for all operations
- [ ] Loading states on buttons during operations
- [ ] Help text explains what each operation does
- [ ] CLI provides helpful output and instructions

### Testing

- [ ] Unit tests cover all error paths
- [ ] Integration tests verify backup/restore cycle
- [ ] E2E tests verify full maintenance mode workflow
- [ ] Manual testing checklist completed

## References

### Existing Code Patterns

- **Middleware pattern**: `internal/middleware/` (CORS, logging, etc.)
- **Database abstraction**: `pkg/database/factory.go` (SQLite/D1 switching)
- **Storage abstraction**: `pkg/storage/factory.go` (Filesystem/R2 switching)
- **Graceful shutdown**: `internal/app.go` (handleShutdown, app.Shutdown)
- **Cobra CLI**: `cmd/main.go` (existing commands: serve, install, migrate)
- **Svelte admin UI**: `web/admin/src/routes/_/` (existing admin pages)

### External Documentation

- [Fiber v3 Graceful Shutdown](https://docs.gofiber.io/guide/shutdown/)
- [Cloudflare Wrangler D1 Commands](https://developers.cloudflare.com/d1/wrangler-commands/)
- [Cloudflare R2 API](https://developers.cloudflare.com/r2/api/)
- [Go exec.Command](https://pkg.go.dev/os/exec#Command)
- [Svelte French Toast](https://github.com/kbrgl/svelte-french-toast) (existing toast library)

### Design Decisions Log

1. **Why file-based toggle instead of database flag?**
   - Simpler implementation (no database dependency)
   - Works even if database is corrupted
   - Easy to check via SSH without running SQL queries

2. **Why IP allowlist instead of authentication?**
   - Simpler for localhost-only access
   - No need to manage separate maintenance admin credentials
   - Standard admin authentication still required for UI access

3. **Why manual backup/restore instead of automatic migration?**
   - Safer: admin explicitly controls data movement
   - Easier to verify backup before switching
   - Clearer separation of concerns
   - Can test restore on new database before committing

4. **Why wrangler CLI instead of direct D1 API?**
   - Wrangler handles authentication and API versioning
   - Official Cloudflare tool with better error messages
   - Simpler implementation (no need to implement D1 REST API client)

5. **Why binary backup format instead of SQL dumps?**
   - Faster backup/restore for large databases
   - Smaller file sizes
   - SQLite-compatible format works for both SQLite and D1
   - Exact data type preservation

---

**End of Design Document**
