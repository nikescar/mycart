# sqlc Migration and Business Logic Testing Design

**Date:** 2026-08-08  
**Status:** Approved  
**Approach:** All Queries First, Then Migrate (TDD)

## Overview

### What We're Building
A complete migration from direct SQL queries to sqlc-generated type-safe queries across all 60 methods in `internal/queries/`, with comprehensive business logic tests running against both SQLite and PostgreSQL.

### Success Criteria
1. **All 60 methods use sqlc** - no direct SQL (QueryContext/ExecContext) except in sqlc-generated code
2. **Zero breaking changes** - existing API surface unchanged
3. **Both DBs work** - all tests pass on SQLite AND PostgreSQL
4. **High test coverage** - 90%+ of business logic tested, including edge cases
5. **No sqlc internals tested** - tests focus on business rules, workflows, validation

### Key Benefits
- **Type safety**: Compile-time SQL validation, no typos
- **Multi-DB support**: Automatic placeholder conversion (`?` vs `$1`)
- **Less boilerplate**: sqlc generates repetitive CRUD code
- **Easier debugging**: Clear separation between query definition and business logic
- **Fewer bugs**: Eliminates entire class of placeholder-related bugs (already found 4 PostgreSQL bugs from this issue)

---

## Architecture

### Current State
```
internal/queries/*.go (60 methods)
  ├─ Direct SQL with QueryContext/ExecContext
  └─ Manual placeholder handling (? vs $1)

internal/database/
  ├─ sqlite_adapter.go (adapter exists but unused)
  └─ postgres_adapter.go (adapter exists but unused)

internal/db/sqlite/ (sqlc generated, unused)
internal/db/postgres/ (sqlc generated, unused)
```

### Target State
```
internal/queries/*.go (business logic layer)
  ├─ Wraps sqlc-generated queries
  ├─ Adds business validation
  └─ Maps between models.* and sqlc types

internal/database/ (adapter layer)
  ├─ sqlite_adapter.go → wraps internal/db/sqlite.Queries
  └─ postgres_adapter.go → wraps internal/db/postgres.Queries

internal/db/sqlite/ (sqlc generated)
  └─ Type-safe SQLite queries

internal/db/postgres/ (sqlc generated)
  └─ Type-safe PostgreSQL queries
```

### Query Flow
```
Handler → queries.DB().GetProduct(ctx, id)
  → ProductQueries.GetProduct() [business validation]
    → dbAdapter.Queries().GetProduct(ctx, id) [sqlc generated]
      → *sql.DB.QueryContext() [stdlib]
```

### Key Design Principles
1. **Adapters select DB-specific sqlc code** - SQLiteAdapter uses `internal/db/sqlite.Queries`, PostgresAdapter uses `internal/db/postgres.Queries`
2. **Queries layer stays database-agnostic** - business logic doesn't know which DB
3. **Models stay in `internal/models`** - sqlc generates its own types, we map between them
4. **One source of truth** - query definitions in `db/queries/*.sql`, business logic in `internal/queries/*.go`

---

## Query Definitions - Filling the Gaps

### Current Coverage Analysis
**Existing sqlc queries** (256 lines across 9 files):
- ✅ settings.sql - 5 queries (GetSettingByKey, UpdateSetting, ListSettings, CreateSetting, DeleteSetting)
- ✅ sessions.sql - 4 queries  
- ✅ pages.sql - 11 queries
- ✅ products.sql - 11 queries
- ✅ carts.sql - 13 queries
- ✅ product_images.sql - 5 queries
- ✅ digital_files.sql - 4 queries
- ✅ digital_data.sql - 6 queries
- ✅ subdomains.sql - 3 queries

### Missing Queries (need to add ~18)

**1. Auth queries** (auth.go has 1 method):
- `GetPasswordByEmail` - currently uses direct SQL

**2. Install queries** (install.go has 2 methods):
- `IsInstalled` - check if cart is set up
- Already handled by settings queries

**3. Product variant queries** (products.go has complex variant handling):
- `GetProductWithVariants` - fetch product + all variants + options
- `CreateProductVariant`
- `UpdateProductVariant`  
- `DeleteProductVariant`
- `ListProductOptions`
- `CreateProductOption`
- `DeleteProductOption`

**4. Advanced product queries**:
- `UpdateProductActive` - toggle active status
- `GetProductsByCategory` - if categories exist
- Bulk operations for images/digital files

**5. Cart email template queries**:
- `GetCartEmailTemplate` (mail_letter_purchase setting)
- Complex JSON handling for templates

### Adding Missing Queries Strategy
For each missing query:
1. **Analyze current SQL** in the direct implementation
2. **Define in appropriate .sql file** with `-- name:` annotation
3. **Handle SQLite vs PostgreSQL differences**:
   - Date functions: `datetime('now')` → `NOW()`
   - JSON: `json_extract` vs `jsonb` operators
   - Auto-increment handling
4. **Run `sqlc generate`** to create Go code
5. **Write test first** (TDD) before using in business logic

---

## Testing Strategy - Dual-Database TDD

### Test Structure
```
internal/queries/*_test.go
  ├─ Unit tests (business logic validation)
  └─ Integration tests (both SQLite + PostgreSQL)

Test execution:
  ├─ SQLite: Fast in-memory (existing pattern)
  └─ PostgreSQL: Real connection (new, uses TEST_DATABASE_URL)
```

### Testing Layers

**1. Business Logic Tests (what we write)**
```go
func TestAddProduct_ValidatesRequiredFields(t *testing.T) {
    runOnBothDBs(t, func(t *testing.T, db *Base) {
        product := &models.Product{Name: ""} // invalid
        _, err := db.AddProduct(ctx, product)
        if err == nil {
            t.Error("expected validation error for empty name")
        }
    })
}

func TestGetProduct_WithVariants(t *testing.T) {
    runOnBothDBs(t, func(t *testing.T, db *Base) {
        // Create product with 2 variants
        // Verify all variants loaded correctly
        // Test business rule: variant prices override base price
    })
}
```

**2. sqlc-Generated Code (NOT tested by us)**
- sqlc team tests query generation
- We trust sqlc handles SQL correctly
- We only test: "does our business logic work?"

**3. Multi-DB Test Helper**
```go
func runOnBothDBs(t *testing.T, testFunc func(*testing.T, *Base)) {
    t.Run("SQLite", func(t *testing.T) {
        db := bootstrapSQLite(t)
        testFunc(t, db)
    })
    
    t.Run("PostgreSQL", func(t *testing.T) {
        if testing.Short() {
            t.Skip("skipping PostgreSQL test in short mode")
        }
        db := bootstrapPostgreSQL(t)
        testFunc(t, db)
    })
}
```

### Test Categories

**A. Validation Tests** (business rules)
- Required fields, length limits, format validation
- Example: product name not empty, email valid format

**B. Workflow Tests** (multi-step operations)  
- Create product → add variants → fetch with variants → verify structure
- Install → check installed → try install again → verify error

**C. Edge Cases** (boundary conditions)
- Empty results, missing data, null handling
- Concurrent updates, race conditions (if applicable)

**D. Multi-DB Compatibility** (platform differences)
- Date handling across DBs
- JSON operations (SQLite json_extract vs PostgreSQL jsonb)
- Transaction isolation

### Missing Tests to Add
- `DeletePage` - no test exists
- Product variant complex scenarios  
- Cart payment validation edge cases
- Multi-step workflows across modules

### Test Execution
```bash
# Fast: SQLite only (local development)
go test -short ./internal/queries/...

# Thorough: Both databases (CI and pre-commit)
export TEST_DATABASE_URL="postgresql://..."
go test ./internal/queries/...
```

---

## Migration Steps (Approach 2 - All Queries First)

### Phase 1: Prepare Query Definitions (TDD Foundation)
**Goal:** Complete all sqlc query definitions before touching implementation

**1. Audit existing queries** (1-2 hours)
- Map all 60 methods to their SQL statements
- Identify the ~18 missing query definitions
- Document SQLite vs PostgreSQL differences

**2. Add missing queries to `db/queries/`** (2-3 hours)
- auth.sql: `GetPasswordByEmail`
- products.sql: Add variant queries (6-8 new queries)
- carts.sql: Add any missing cart operations
- settings.sql: Ensure all setting operations covered

**3. Handle DB-specific syntax** (1 hour)
- Create parallel PostgreSQL versions in `db/queries/postgres/`
- SQLite → PostgreSQL translations:
  - `datetime('now')` → `NOW()`
  - `?` → `$1, $2, $3`
  - `INTEGER AUTOINCREMENT` → `SERIAL`

**4. Generate sqlc code**
```bash
sqlc generate
```
- Produces updated `internal/db/sqlite/*.go`
- Produces updated `internal/db/postgres/*.go`
- Verify no compilation errors

### Phase 2: Write All Tests First (Pure TDD)
**Goal:** Have comprehensive test coverage before changing any implementation

**5. Set up dual-DB test infrastructure** (1-2 hours)
- Create `runOnBothDBs()` helper
- `bootstrapSQLite()` - existing pattern
- `bootstrapPostgreSQL()` - new, uses TEST_DATABASE_URL
- Update existing tests to use dual-DB runner

**6. Write business logic tests for all modules** (6-8 hours)
```
Priority order (write tests for these):
├─ Settings (6 methods) - foundation
├─ Install/Auth (3 methods) - critical path  
├─ Session (4 methods) - independent
├─ Pages (9 methods) - moderate
├─ Cart (7 methods) - complex workflows
└─ Products (31 methods) - most complex
```

**7. Add missing test coverage** (2-3 hours)
- `DeletePage` test
- Product variant edge cases
- Complex multi-step workflows
- Cart payment validation scenarios

**At this point:** All tests written, all FAILING (because implementation still uses direct SQL)

### Phase 3: Migrate Implementation (Make Tests Pass)
**Goal:** Replace direct SQL with sqlc, make all tests green

**8. Update adapter Queries() method** (30 mins)
```go
// internal/database/sqlite_adapter.go
func (a *SQLiteAdapter) Queries() interface{} {
    return sqlite.New(a.db) // returns *sqlite.Queries
}

// internal/database/postgres_adapter.go  
func (a *PostgresAdapter) Queries() interface{} {
    return postgres.New(a.db) // returns *postgres.Queries
}
```

**9. Migrate each module to use sqlc** (4-6 hours)
```go
// OLD (direct SQL):
query := "SELECT id, key, value FROM setting WHERE key = ?"
row := q.DB.QueryRowContext(ctx, query, key)

// NEW (sqlc):
queries := dbAdapter.Queries().(interface {
    GetSettingByKey(ctx, key) (sqlite.Setting, error)
})
result, err := queries.GetSettingByKey(ctx, key)
```

Migration order:
- settings.go → Use GetSettingByKey, UpdateSetting, etc.
- install.go → Use setting queries
- auth.go → Use GetPasswordByEmail
- session.go → Use session queries
- pages.go → Use page queries
- cart.go → Use cart queries
- products.go → Use product + variant queries (most complex)

**10. Type mapping** (ongoing during migration)
- Map `sqlite.Setting` → `models.SettingName`
- Map `postgres.Product` → `models.Product`
- Handle nullability differences

**11. Remove direct SQL helpers** (30 mins)
- Delete `buildPlaceholders()` function
- Remove manual placeholder logic
- Clean up unused imports

### Phase 4: Verification & Cleanup
**Goal:** Ensure everything works, clean up dead code

**12. Run full test suite**
```bash
# SQLite tests
go test -v ./internal/queries/...

# PostgreSQL tests
export TEST_DATABASE_URL="postgresql://..."
go test -v ./internal/queries/...
```

**13. Verify both databases in real app**
- Test install with SQLite
- Test install with PostgreSQL
- Smoke test major workflows

**14. Remove dead code** (30 mins)
- Delete old migration placeholder helpers
- Remove any unused query methods
- Clean up imports

**15. Update documentation**
- Add comment: "// Query generated by sqlc, see db/queries/*.sql"
- Document type mappings if non-obvious

### Estimated Timeline
- **Phase 1 (Queries):** 4-6 hours
- **Phase 2 (Tests):** 8-11 hours  
- **Phase 3 (Migration):** 5-7 hours
- **Phase 4 (Verify):** 1-2 hours
- **Total:** ~18-26 hours (2-3 full days)

---

## Error Handling & Edge Cases

### Type Conversion Safety
**Challenge:** sqlc generates DB-specific types, we use `models.*` types

**Solution:** Explicit mapping functions with error handling
```go
func toModelProduct(p sqlite.Product) (*models.Product, error) {
    // Handle nullable fields
    var brief string
    if p.Brief.Valid {
        brief = p.Brief.String
    }
    
    // Validate required fields
    if p.Name == "" {
        return nil, errors.New("product name required")
    }
    
    return &models.Product{
        Core: models.Core{ID: p.ID},
        Name: p.Name,
        Brief: brief,
        // ... map all fields
    }, nil
}
```

### Database Compatibility Issues
**Common pitfalls:**
1. **JSON handling** - SQLite `json_extract` vs PostgreSQL `jsonb ->`
2. **Date formats** - Store as UTC timestamps, handle timezones
3. **Transaction isolation** - Test concurrent operations on both DBs
4. **Null handling** - sqlc uses `sql.NullString`, models use pointers or empty strings

**Mitigation:**
- Dual-DB tests catch compatibility issues early
- Document any DB-specific workarounds in query comments
- Use standard SQL where possible

### Test Failure Handling
**During migration:** Tests will fail as we replace code

**Strategy:**
- Migrate one module at a time within Phase 3
- Run tests after each module migration
- If tests fail: fix immediately before moving to next module
- Don't commit partially migrated code

---

## Rollback Plan

### If Migration Fails Mid-Way

**Before Migration:**
- Create feature branch: `feature/sqlc-migration`
- Tag current main: `git tag pre-sqlc-migration`

**Rollback Options:**

**Option A: Revert commits (safest)**
```bash
git revert <migration-commits>
# Keeps history, safe for shared branches
```

**Option B: Reset branch (if not pushed)**
```bash
git reset --hard pre-sqlc-migration
# Clean slate, only if branch is private
```

**Option C: Keep adapters, revert queries**
- Keep new adapter code (it's good infrastructure)
- Revert only `internal/queries/*.go` changes
- Adapters can stay unused until next attempt

### Verification Checkpoints
✅ **After Phase 1:** sqlc generates without errors  
✅ **After Phase 2:** Tests compile (even if failing)  
✅ **After each module in Phase 3:** Tests pass for that module  
✅ **After Phase 4:** Full test suite green on both DBs

**If any checkpoint fails:** Stop, fix, don't proceed

### Production Safety
- This is a refactoring (no behavior change)
- If tests pass on both DBs → safe to deploy
- Monitor error rates after deploy
- Keep old tag for quick rollback if needed

---

## Success Metrics

### Code Quality
- ✅ Zero direct SQL in `internal/queries/` (except sqlc-generated)
- ✅ No manual placeholder handling
- ✅ All queries defined in `.sql` files
- ✅ Type-safe query execution

### Test Coverage
- ✅ 90%+ of business logic tested
- ✅ All tests pass on both SQLite and PostgreSQL
- ✅ Edge cases covered (DeletePage, variants, workflows)
- ✅ No tests of sqlc internals (only business logic)

### Performance
- ✅ SQLite tests run fast (<10s total)
- ✅ PostgreSQL tests complete in reasonable time (<30s)
- ✅ No performance regression in production

### Maintainability
- ✅ Clear separation: queries in .sql, logic in .go
- ✅ Easy to add new queries (add to .sql, regenerate)
- ✅ Multi-DB support transparent to business logic
- ✅ Less boilerplate, more readable code
