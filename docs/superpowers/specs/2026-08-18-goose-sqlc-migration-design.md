# Goose+Sqlc Migration Design

**Date:** 2026-08-18  
**Author:** Claude Sonnet 4.5  
**Status:** Design Approved

---

## Overview

This spec defines the migration from the current goose-only database layer to a goose+sqlc hybrid system with zero-overhead function pointers for database access. The migration consolidates all data access under `internal/store/` and moves migration infrastructure to `internal/goosemigration/`.

---

## Goals

1. **Zero Runtime Overhead:** Use function pointers for database abstraction (0ns per query)
2. **Single Binary:** Support both PostgreSQL and SQLite in one binary with runtime switching
3. **Clean Architecture:** Separate business logic helpers from simple database queries
4. **Migration Safety:** Move old code to `internal/goosemigration/` for migration infrastructure
5. **Maintainability:** Reduce code duplication and improve testability

---

## Design Decisions

### Decision 1: Database Abstraction Strategy
**Chosen:** Function pointers initialized once at startup  
**Alternatives Considered:**
- Interface-based adapter (~5-10ns overhead per query)
- Direct type checking in business logic (code duplication)
- Build tags (separate binaries required)

**Rationale:** Function pointers provide true zero runtime overhead while maintaining single binary + runtime database switching capability.

### Decision 2: Directory Structure
**Chosen:** `internal/store/` for all data access, `internal/store/db/` for sqlc-generated code  
**Alternatives Considered:**
- Keep `internal/db/` separate from business logic
- Use `internal/repository/` or `internal/dao/`

**Rationale:** Grouping all database operations under `internal/store/` creates a clear boundary. Handlers only import `internal/store`, never the underlying `db/` package.

### Decision 3: Migration Strategy
**Chosen:** Big Bang (single PR)  
**Alternatives Considered:**
- Incremental migration (2-3 weeks, module by module)
- Feature flag migration (3-4 weeks, gradual rollout)

**Rationale:** User preference for speed. Risk is acceptable with comprehensive testing strategy.

### Decision 4: Business Logic Separation
**Chosen:** Extract Type 2 methods to `internal/store/*.go`, Type 1 becomes function pointers  
**Alternatives Considered:**
- Keep all methods in wrapper layer
- Convert all to standalone functions

**Rationale:** Current codebase has significant business logic in query layer (validation, type conversion, transactions). This logic belongs in `store/`, not in function pointers.

---

## Architecture

### Final Directory Structure

```
internal/
├── store/                           # NEW: All data access layer
│   ├── db/                         # NEW: sqlc + function pointers (private)
│   │   ├── init.go                # Initialize function pointers at startup
│   │   ├── queries.go             # Function pointer variables (~50 pointers)
│   │   ├── types.go               # Unified parameter/result types
│   │   ├── postgres/              # MOVED: sqlc generated
│   │   │   ├── db.go
│   │   │   ├── models.go
│   │   │   ├── auth.sql.go
│   │   │   └── ...
│   │   └── sqlite/                # MOVED: sqlc generated
│   │       ├── db.go
│   │       ├── models.go
│   │       ├── auth.sql.go
│   │       └── ...
│   ├── settings.go                # NEW: Business logic helpers
│   ├── auth.go                    # NEW: Business logic helpers
│   ├── carts.go                   # NEW: Business logic helpers
│   ├── products.go                # NEW: Business logic helpers
│   ├── pages.go                   # NEW: Business logic helpers
│   └── converters.go              # NEW: Type conversion helpers
│
├── goosemigration/                 # NEW: Migration infrastructure only
│   └── database/                  # MOVED from internal/database
│       ├── config.go
│       ├── connect.go
│       ├── migrate.go
│       ├── migrate_data.go
│       └── ...
│
└── handlers/                       # UPDATED: Use internal/store
    ├── private/
    │   ├── auth.go                # Uses store.* and db.*Func()
    │   └── ...
    └── public/
        └── ...

db/                                 # UNCHANGED
├── migrations/                    # UNCHANGED (goose reads from here)
│   ├── embed.go
│   ├── postgres/*.sql
│   └── sqlite/*.sql
└── queries/                       # UNCHANGED (sqlc reads from here)
    ├── postgres/*.sql
    └── sqlite/*.sql
```

### Component Responsibilities

| Component | Responsibility | Imports |
|-----------|---------------|---------|
| **internal/store/db/** | sqlc-generated code + function pointers | postgres, sqlite packages |
| **internal/store/** | Business logic helpers, type conversions | internal/store/db, internal/models |
| **internal/goosemigration/database/** | Database connections, migrations, data migration | goose, embed |
| **internal/handlers/** | HTTP request handling | internal/store only |

---

## Detailed Design

### 1. Sqlc Configuration Changes

**File:** `sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "db/migrations/postgres"         # ← UNCHANGED
    queries: "db/queries/postgres"           # ← UNCHANGED
    gen:
      go:
        package: "postgres"
        out: "internal/store/db/postgres"    # ← CHANGED (was internal/db/postgres)
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: false
        emit_exact_table_names: false
        emit_empty_slices: true
        
  - engine: "sqlite"
    schema: "db/migrations/sqlite"           # ← UNCHANGED
    queries: "db/queries/sqlite"             # ← UNCHANGED
    gen:
      go:
        package: "sqlite"
        out: "internal/store/db/sqlite"      # ← CHANGED (was internal/db/sqlite)
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: false
        emit_exact_table_names: false
        emit_empty_slices: true
```

**Impact:** Sqlc will generate code in new location. Old `internal/db/` can be deleted.

---

### 2. Function Pointer Infrastructure

#### Overview

Function pointers are initialized once at application startup based on database type (postgres or sqlite). After initialization, all database calls use direct function pointers with zero runtime overhead.

#### Key Files

- **`internal/store/db/init.go`** - Initialization logic (~300 lines)
- **`internal/store/db/queries.go`** - Function pointer declarations (~100 lines)
- **`internal/store/db/types.go`** - Unified types (~200 lines)

#### Example: init.go (abbreviated)

```go
package db

import (
    "database/sql"
    "github.com/shurco/mycart/internal/store/db/postgres"
    "github.com/shurco/mycart/internal/store/db/sqlite"
)

// Init initializes function pointers based on database type
func Init(db *sql.DB, dbType string) error {
    if dbType == "postgres" || dbType == "postgresql" {
        initPostgres(db)
    } else {
        initSQLite(db)
    }
    return nil
}

func initPostgres(db *sql.DB) {
    q := postgres.New(db)
    GetPasswordByEmailFunc = q.GetPasswordByEmail
    GetSettingByKeyFunc = q.GetSettingByKey
    // ... assign all ~50 function pointers
}

func initSQLite(db *sql.DB) {
    q := sqlite.New(db)
    GetPasswordByEmailFunc = q.GetPasswordByEmail
    GetSettingByKeyFunc = q.GetSettingByKey
    // ... assign all ~50 function pointers
}
```

**Performance:**
- **Startup:** ~5-10ns per function pointer (one-time cost)
- **Per query:** **0ns overhead** (direct function call)
- **Memory:** ~400 bytes total (50 function pointers × 8 bytes)

---

### 3. Business Logic Layer

Business logic helpers extracted from `internal/queries/` to `internal/store/`:

**Type 1 Methods (Simple Pass-Through):**
- Convert to function pointers in `internal/store/db/`
- Zero overhead, direct database calls
- Example: `GetSettingByKey`, `GetProductByID`

**Type 2 Methods (Business Logic):**
- Extract to `internal/store/*.go`
- Complex operations: validation, type conversion, transactions
- Example: `GetSettingByGroup[T]()`, `UpdatePassword()`, `GetCarts()`

**Type 3 Helpers (Utilities):**
- Pure utility functions in `internal/store/converters.go`
- No database access
- Example: `unmarshalJSONToPointer[T]()`, `marshalJSONFromPointer[T]()`

#### Key Files

- **`internal/store/settings.go`** - Settings business logic (~400 lines)
- **`internal/store/carts.go`** - Cart business logic (~200 lines)
- **`internal/store/products.go`** - Product business logic (~100 lines)
- **`internal/store/pages.go`** - Page business logic (~50 lines)
- **`internal/store/auth.go`** - Auth business logic (~50 lines)
- **`internal/store/converters.go`** - Type conversion utilities (~100 lines)

---

### 4. Handler Updates

Handlers import `internal/store` and `internal/store/db`:

**Before:**
```go
import "github.com/shurco/mycart/internal/queries"

func SignIn(c fiber.Ctx) error {
    db := queries.DB()
    password, err := db.GetPasswordByEmail(ctx, email)
    jwt, err := queries.GetSettingByGroup[models.JWT](ctx, db)
}
```

**After:**
```go
import (
    "github.com/shurco/mycart/internal/store"
    "github.com/shurco/mycart/internal/store/db"
)

func SignIn(c fiber.Ctx) error {
    // Function pointer (zero overhead)
    passwordSetting, err := db.GetPasswordByEmailFunc(ctx)
    
    // Business logic helper
    jwt, err := store.GetSettingByGroup[models.JWT](ctx)
}
```

---

### 5. Application Initialization

**File:** `internal/app.go`

```go
import (
    "github.com/shurco/mycart/db/migrations"
    "github.com/shurco/mycart/internal/goosemigration/database"
    "github.com/shurco/mycart/internal/store/db"
)

func Initialize() error {
    // Load database config
    cfg, err := database.LoadConfig()
    if err != nil {
        return err
    }
    
    // Connect to database
    dbAdapter, err := database.ConnectWithRetry(cfg)
    if err != nil {
        return err
    }
    
    // Run migrations
    if err := database.RunMigrations(dbAdapter.DB(), cfg.Type, migrations.Embed()); err != nil {
        return err
    }
    
    // Initialize function pointers (NEW!)
    if err := db.Init(dbAdapter.DB(), cfg.Type); err != nil {
        return err
    }
    
    return nil
}
```

---

## Migration Steps (Big Bang)

### Step 1: Update sqlc.yaml
Change output paths from `internal/db/*` to `internal/store/db/*`

### Step 2: Regenerate sqlc code
```bash
sqlc generate
```

### Step 3: Create function pointer infrastructure
- `internal/store/db/init.go`
- `internal/store/db/queries.go`
- `internal/store/db/types.go`

### Step 4: Extract business logic to store/
- `internal/store/settings.go`
- `internal/store/carts.go`
- `internal/store/products.go`
- `internal/store/pages.go`
- `internal/store/auth.go`
- `internal/store/converters.go`

### Step 5: Move migration infrastructure
```bash
mkdir -p internal/goosemigration
mv internal/database internal/goosemigration/database
mv internal/queries internal/goosemigration/queries  # temporary
```

### Step 6: Update migration infrastructure imports
- `internal/migrate_db.go`: Update import path

### Step 7: Initialize function pointers at startup
- Update `internal/app.go` to call `db.Init()`

### Step 8: Update all handlers (~27 files)
- Replace `queries.DB()` calls
- Type 1 → `db.FunctionNameFunc()`
- Type 2 → `store.HelperName()`

### Step 9: Update all tests (~20 files)
- Replace `queries.NewFromDB()` with `db.Init()`

### Step 10: Delete old code
```bash
rm -rf internal/db/
rm -rf internal/goosemigration/queries/
```

### Step 11: Verify
```bash
go mod tidy
go build ./...
go test ./... -count=1 -race
sqlc generate
go run ./cmd serve
```

---

## Testing Strategy

### Test Coverage Requirements

| Component | Minimum Coverage |
|-----------|-----------------|
| **internal/store/db/** | 90% |
| **internal/store/** | 85% |
| **internal/handlers/** | 80% |
| **internal/goosemigration/** | 75% |

### Pre-Merge Checklist

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Database compatibility tests pass (postgres + sqlite)
- [ ] Full test suite: `go test ./... -count=1 -race`
- [ ] Code coverage >= 80%
- [ ] Build succeeds: `go build ./...`
- [ ] App starts with sqlite: `DB_TYPE=sqlite go run ./cmd serve`
- [ ] App starts with postgres: `DB_TYPE=postgres go run ./cmd serve`
- [ ] Migrations work: `go run ./cmd migrate-to-postgres`
- [ ] Sqlc generation: `sqlc generate`
- [ ] No old imports: `grep -r "internal/queries" . --include="*.go"`
- [ ] Old directories removed

---

## Performance Analysis

### Startup Cost (One-Time)
- Function pointer initialization: ~250-500ns total
- Memory footprint: ~400 bytes

### Runtime Cost (Per Query)
- **Before:** ~7-15ns (wrapper layer overhead)
- **After:** **0ns** (direct function call)
- **Improvement:** 100% elimination of per-query overhead

**Real-World Impact:** Database I/O dominates (1-10ms), but zero overhead is still better than any overhead.

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Big Bang breaks everything | Medium | High | Comprehensive test suite |
| Type conversion bugs | Medium | Medium | Compatibility tests |
| Missed business logic | Low | Medium | Side-by-side code comparison |
| Uninitialized function pointer | Low | High | Initialization tests |

---

## Success Criteria

- [ ] All handlers work with postgres and sqlite
- [ ] Zero per-query overhead (verified by benchmarks)
- [ ] Test coverage >= 80%
- [ ] Single binary supports both databases
- [ ] No API breaking changes

---

## Files Affected Summary

- **Create:** 9 new files in `internal/store/`
- **Move:** 2 directories to `internal/goosemigration/`
- **Update:** ~60-70 files (handlers, tests, config)
- **Delete:** 2 directories (old code)

---

**End of Design Document**
