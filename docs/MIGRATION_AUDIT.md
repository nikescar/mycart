# Database Layer Migration Audit

**Date:** 2026-08-18  
**Branch:** feature/postgresql-support  
**Goal:** Migrate all business logic from goosemigration/queries (goose-generated) to internal/store (sqlc-generated)

## Summary

| Component | Status | Progress |
|-----------|--------|----------|
| Function Pointer Infrastructure | ✅ Complete | 100% |
| Settings Operations | ✅ Complete | 100% |
| Auth Operations | ✅ Complete | 100% |
| Session Operations | ⚠️  Partial | 66% (2/3 methods) |
| Products Operations | ❌ Not Started | 0% (0/15+ methods) |
| Carts Operations | ❌ Not Started | 0% (0/4 methods) |
| Pages Operations | ❌ Not Started | 0% (0/2 methods) |
| **Handlers Migrated** | **⚠️  1/27** | **4%** |

## Detailed Analysis

### ✅ Completed Operations (internal/store/)

**Settings (internal/store/settings.go):**
- `GetSettingByGroup(ctx, settings)` - Complex type conversion and JSON unmarshalling
- `GetSettingByGroupTyped[T](ctx)` - Generic typed version
- `UpdateSettingByGroup(ctx, settings)` - Complex field mapping and validation
- `UpdatePassword(ctx, password)` - Password hashing and security
- `GetSettingByKey(ctx, ...keys)` - Multi-key retrieval with map return
- `UpdateSettingByKey(ctx, setting)` - Single key update

**Auth (internal/store/auth.go):**
- `GetPasswordByEmail(ctx, email)` - Email validation + password retrieval

**Sessions (internal/store/sessions.go):**
- `AddSession(ctx, key, value, expires)` - Database-specific UPSERT syntax
- `DeleteSession(ctx, key)` - Simple delete wrapper

### ⚠️  Partial Migration

**Sessions - Missing:**
- `GetSession(ctx, key)` - Used in handlers/private/setting.go:72 for version caching

### ❌ Not Migrated Operations

**Pages (used in 2+ handlers):**
```
db.Page(ctx, slug string) (*models.Page, error)
db.ListPages(ctx) ([]*models.Page, error) [if exists]
db.AddPage(ctx, page) (*models.Page, error) [if exists]
db.UpdatePage(ctx, page) error [if exists]
db.DeletePage(ctx, pageID) error [if exists]
```

**Products (used in 10+ handlers):**
```
db.ListProducts(ctx, private, limit, offset, cartID, idList...)
db.Product(ctx, private, productID)
db.AddProduct(ctx, product)
db.AddProductWithVariants(ctx, product)
db.UpdateProduct(ctx, product)
db.DeleteProduct(ctx, productID)
db.UpdateActive(ctx, productID)
db.ProductImages(ctx, productID)
db.AddImage(ctx, productID, uuid, ext, origName)
db.DeleteImage(ctx, productID, imageID)
db.ProductDigital(ctx, productID)
db.AddDigitalFile(ctx, productID, uuid, ext, origName)
db.AddDigitalData(ctx, productID, content)
db.UpdateDigital(ctx, digital)
db.DeleteDigital(ctx, productID, digitalID)
db.GenerateUniqueSlug(ctx, name, excludeID)
```

**Carts (used in 3+ handlers):**
```
db.Carts(ctx, limit, offset)
db.Cart(ctx, cartID)
queries.BuildCartItems(cart, products) - utility function
[Plus cart update/creation operations not yet identified]
```

## Handler Migration Status

### ✅ Migrated (1)
- `internal/handlers/private/auth.go` - SignIn, SignOut

### ❌ Not Migrated (26+)

**Private Handlers (needs migration):**
- `internal/handlers/private/cart.go` - Carts, Cart, CartSendMail
- `internal/handlers/private/install.go` - Install operations
- `internal/handlers/private/page.go` - Page CRUD operations
- `internal/handlers/private/product.go` - 15+ product endpoints
- `internal/handlers/private/setting.go` - Version, GetSetting, UpdateSetting, TestLetter

**Public Handlers (needs migration):**
- `internal/handlers/public/cart.go` - Public cart operations
- `internal/handlers/public/page.go` - Page(slug)
- `internal/handlers/public/product.go` - Products(), Product(id)
- `internal/handlers/public/payment_portone.go` - Payment operations
- `internal/handlers/public/setting.go` - Public settings

**Tests (20+ files):**
- All test files still use goosemigration/queries pattern

## Migration Strategy

### Option A: Complete Migration (Recommended for Production)

**Effort:** High (2-4 hours)  
**Benefit:** Full migration, no legacy code

1. Extract Pages business logic → `internal/store/pages.go`
2. Extract Products business logic → `internal/store/products.go`
3. Extract Carts business logic → `internal/store/carts.go`
4. Add GetSession to `internal/store/sessions.go`
5. Update all 26 handlers to use `internal/store`
6. Update all 20 test files
7. Remove `internal/goosemigration/queries` dependency from handlers
8. Verify build and tests

### Option B: Gradual Migration (Current Approach)

**Effort:** Medium (incremental)  
**Benefit:** Lower risk, testable at each step

1. Pick one domain (Pages, Products, or Carts)
2. Extract business logic to internal/store/
3. Update handlers for that domain
4. Test and verify
5. Repeat for next domain

### Option C: Coexistence (Hybrid - Not Recommended)

**Effort:** Low (minimal changes)  
**Benefit:** Quick short-term solution  
**Risk:** Technical debt, two patterns in codebase

- Keep goosemigration/queries for complex operations
- Move only to store layer
- Accept dual patterns

## Recommendations

**For Production Readiness:**
- Complete Option A (full migration)
- Ensures single source of truth
- Cleaner architecture
- Better maintainability

**For Incremental Progress:**
- Use Option B with this order:
  1. **Pages** (simplest, 2-3 methods)
  2. **Carts** (medium complexity, 4-5 methods)
  3. **Products** (most complex, 15+ methods)

## Verification Checklist

After migration complete:
- [ ] No imports of `internal/goosemigration/queries` in handlers
- [ ] All handlers use `internal/store` methods
- [ ] All store methods use `internal/store/db` function pointers (sqlc-generated)
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] No references to queries.DB() in business logic
- [ ] MIGRATION_STATUS.md updated with final counts

## Current Blockers

None. Infrastructure is complete. Work is purely mechanical conversion following established patterns (see auth.go example).
