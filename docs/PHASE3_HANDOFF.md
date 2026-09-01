# Phase 3 Handoff: Storage Abstraction Integration

## Completed Work (Phase 1-2)

### Phase 1: Infrastructure ✅
- Created `pkg/database` abstraction layer (Database, Tx, Stmt interfaces)
- Implemented SQLite adapter with full interface support
- Created D1 adapter stubs for Cloudflare deployment
- Created `pkg/storage` abstraction layer (Storage interface)
- Implemented Filesystem storage adapter
- Created R2 storage adapter stub
- Added environment detection in `internal/config/detect.go`
- Factory pattern for auto-routing based on environment

**Commits:**
- `d3e7536` - Phase 1: Database and storage abstraction layers

### Phase 2: Database Integration ✅
- Migrated all 7 query modules to use `database.Database` interface
- Updated 200+ method calls (QueryContext → Query, etc.)
- Added Prepare/Stmt support for prepared statements
- Fixed helper functions to accept `database.Tx` parameters
- Updated app.go, install.go, tests, and seed script
- **Full compilation success**

**Commits:**
- `172e2d4` - Phase 2: Integrate database abstraction layer

**Files Modified (31 total):**
```
pkg/database/interface.go       - Added Prepare/Stmt interfaces
pkg/database/sqlite.go          - Implemented Prepare, added sqliteStmt
pkg/database/d1.go              - Added Prepare stubs
internal/queries/*.go           - All modules use database.Database
internal/app.go                 - Uses config.LoadDatabaseConfig()
internal/install.go             - Uses database factory
internal/handlers/private/product.go - Fixed CSV importer DB access
cmd/seed_products/main.go       - Updated initialization
```

## Phase 3 Plan: Storage Integration

### Current State Analysis

**File Metadata Model:**
```go
// Product images and digital files use this
type File struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    Ext      string `json:"ext"`
    OrigName string `json:"orig_name,omitempty"`
}

// Digital products support files
type Digital struct {
    Type   string `json:"type"`  // "file", "data", "api"
    Filled bool   `json:"filled,omitempty"`
    Files  []File `json:"files,omitempty"`
    Data   []Data `json:"data,omitempty"`
}
```

**Storage Interface (already created):**
```go
type Storage interface {
    Put(ctx context.Context, path string, data io.Reader) error
    Get(ctx context.Context, path string) (io.ReadCloser, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    List(ctx context.Context, prefix string) ([]string, error)
}
```

### Integration Points to Find

1. **File Upload Handlers** (need to locate):
   - Product image uploads
   - Digital product file uploads
   - Search for: `FormFile`, `MultipartForm`, `SaveFile`

2. **File Serving** (need to locate):
   - Image serving endpoints
   - Digital file downloads
   - Search for: `SendFile`, `ServeFile`, static file routes

3. **Storage Initialization**:
   - Add storage factory to `internal/app.go` (similar to database)
   - Inject storage into handlers that need it

4. **CSV Import/Export**:
   - `cmd/seed_products/main.go` - already uses filesystem
   - Handler at `internal/handlers/private/product.go` line ~645

### Next Steps (Start Fresh Session)

**Step 1: Locate Integration Points**
```bash
# Find file upload handling
grep -rn "FormFile\|MultipartForm" internal/handlers
grep -rn "SaveFile" internal/

# Find file serving
grep -rn "SendFile\|ServeFile" internal/handlers
grep -rn "StaticFS\|Static" internal/

# Find direct filesystem operations
grep -rn "os\\.Create\|os\\.Open\|ioutil" internal/handlers
```

**Step 2: Initialize Storage in App**
```go
// internal/app.go (add after database init)
storageConfig := config.LoadStorageConfig()
store, err := storage.New(storageConfig)
if err != nil {
    log.Err(err).Send()
    return err
}
```

**Step 3: Update Handlers**
- Pass `storage.Storage` to handlers that need it
- Replace direct filesystem calls with storage interface
- Test local filesystem operations still work

**Step 4: Test & Validate**
- Verify uploads work with filesystem adapter
- Verify downloads work
- Verify file deletion works
- Run full test suite

## Environment Detection

Already implemented in `internal/config/detect.go`:

```go
func LoadDatabaseConfig() *database.Config {
    if IsCloudflare() {
        return &database.Config{Type: "d1", ...}
    }
    return &database.Config{Type: "sqlite", Path: "./lc_base/data.db"}
}

func LoadStorageConfig() *storage.Config {
    if IsCloudflare() {
        return &storage.Config{Type: "r2", ...}
    }
    return &storage.Config{Type: "filesystem", BasePath: "./"}
}
```

## Remaining Phases (After Phase 3)

**Phase 4: Cloudflare Infrastructure**
- Create `Dockerfile` with distroless base
- Create `wrangler.jsonc` for Container Workers
- Create TypeScript worker entry point (`src/index.ts`)
- Configure environment variables

**Phase 5: Testing & Documentation**
- Test local CLI server (existing functionality)
- Test Cloudflare deployment simulation
- Update README with deployment instructions
- Document environment variable requirements

## Key Files Reference

**Abstraction Layers:**
- `pkg/database/interface.go` - Database abstraction
- `pkg/database/factory.go` - Database routing
- `pkg/storage/interface.go` - Storage abstraction  
- `pkg/storage/factory.go` - Storage routing
- `internal/config/detect.go` - Environment detection

**Current Branch:**
```
feature/cloudflare-container-workers
```

**Test Command:**
```bash
go build ./...              # Full compilation
go test ./internal/queries  # Query tests
go run ./cmd serve          # Local server
```

## Notes

- File handling appears minimal in current codebase
- Files might be handled through SvelteKit frontends
- Integration may be simpler than initially expected
- Storage factory pattern already complete (Phase 1)
- Just need to wire it into the application initialization

**Ready for fresh session to complete Phase 3-5.**
