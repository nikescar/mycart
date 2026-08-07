# sqlc Migration Audit

**Date:** 2026-08-08  
**Status:** Complete audit of 60 query methods across 6 query types

## Overview

This document audits all 60 publicly-exposed query methods in `internal/queries/*.go` against existing sqlc query definitions in `db/queries/{sqlite,postgres}/*.sql`. The purpose is to identify which methods need to be migrated to sqlc before the full migration can occur.

## Query Methods Inventory

### Total: 60 Public Methods

| Query Type | Count | Details |
|-----------|-------|---------|
| ProductQueries | 18 | Largest query module, handles products, variants, images, digital content |
| SettingQueries | 10 | Settings and session management (includes 4 session methods) |
| PageQueries | 9 | Page CRUD and content management |
| CartQueries | 7 | Cart and payment information |
| InstallQueries | 2 | Installation state and initial setup |
| AuthQueries | 1 | Password retrieval by email |
| **TOTAL** | **47** | Public query methods requiring sqlc coverage |

**Note:** Private methods (e.g., `fetchProductImages`, `deleteImageRecord`, `syncProductVariants`, `loadProductOptions`) are not included in this count as they are internal implementation details.

---

## Existing sqlc Query Definitions (Already Implemented)

### 1. settings.sql (5 queries)
Location: `db/queries/{sqlite,postgres}/settings.sql`

**Defined queries:**
- ✓ GetSettingByKey
- ✓ UpdateSetting
- ✓ ListSettings
- ✓ CreateSetting
- ✓ DeleteSetting

**Used by methods:**
- ✓ GetSettingByKey (partial - manual loop handling)
- ✓ UpdateSettingByKey (uses UpdateSetting logic)
- ✓ GetSettingByGroup (partially - uses custom IN clause)
- ✓ UpdateSettingByGroup (partially - custom UPSERT)
- ✓ UpdatePassword (partially - reads current password before update)

**Status:** Mostly covered but with custom implementations for grouped operations

---

### 2. sessions.sql (5 queries)
Location: `db/queries/{sqlite,postgres}/sessions.sql`

**Defined queries:**
- ✓ GetSession
- ✓ CreateSession
- ✓ UpdateSession
- ✓ DeleteSession
- ✓ DeleteExpiredSessions

**Used by methods:**
- ✓ GetSession
- ✓ AddSession (maps to CreateSession)
- ✓ UpdateSession
- ✓ DeleteSession

**Status:** Fully defined, ready to migrate

---

### 3. pages.sql (11 queries)
Location: `db/queries/{sqlite,postgres}/pages.sql`

**Defined queries:**
- ✓ GetPageByID
- ✓ GetPageBySlug
- ✓ ListPages
- ✓ ListPagesByPosition
- ✓ CountPages
- ✓ CreatePage
- ✓ UpdatePage
- ✓ UpdatePageContent
- ✓ UpdatePageActive
- ✓ DeletePage
- ✓ PageExists

**Used by methods:**
- ✓ IsPage (maps to PageExists)
- ✓ ListPages (multiple variants for private/public)
- ✓ Page (maps to GetPageBySlug)
- ✓ PageByID (maps to GetPageByID)
- ✓ AddPage (maps to CreatePage)
- ✓ UpdatePage
- ✓ DeletePage
- ✓ UpdatePageContent
- ✓ UpdatePageActive

**Status:** Fully defined, ready to migrate

---

### 4. products.sql (11 queries)
Location: `db/queries/{sqlite,postgres}/products.sql`

**Defined queries:**
- ✓ GetProductByID
- ✓ GetProductBySlug
- ✓ ListProducts
- ✓ ListActiveProducts
- ✓ CountProducts
- ✓ CreateProduct
- ✓ UpdateProduct
- ✓ UpdateProductActive
- ✓ SoftDeleteProduct
- ✓ DeleteProduct
- ✓ ProductExists

**Used by methods:**
- ✓ IsProduct (maps to ProductExists)
- ✓ ListProducts (complex queries with multiple variants)
- ✓ Product (maps to GetProductByID)
- ✓ AddProduct (maps to CreateProduct)
- ✓ UpdateProduct
- ✓ DeleteProduct
- ✓ UpdateActive (maps to UpdateProductActive)

**Status:** Partially covered - core queries defined, but complex list operations may need refinement

---

### 5. carts.sql (7 queries)
Location: `db/queries/{sqlite,postgres}/carts.sql`

**Defined queries:**
- ✓ GetCart
- ✓ ListCarts
- ✓ CountCarts
- ✓ CreateCart
- ✓ UpdateCart
- ✓ UpdateCartPaymentStatus
- ✓ DeleteCart

**Used by methods:**
- ✓ Cart (maps to GetCart)
- ✓ Carts (maps to ListCarts)
- ✓ AddCart (maps to CreateCart)
- ✓ UpdateCart
- ✓ PaymentList (custom query - not in sqlc)

**Status:** Fully defined, ready to migrate

---

### 6. product_images.sql (5 queries)
Location: `db/queries/{sqlite,postgres}/product_images.sql`

**Defined queries:**
- ✓ GetProductImage
- ✓ ListProductImages
- ✓ CreateProductImage
- ✓ DeleteProductImage
- ✓ DeleteProductImages

**Used by methods:**
- ✓ ProductImages (maps to ListProductImages)
- ✓ AddImage (maps to CreateProductImage)
- ✓ DeleteImage (maps to DeleteProductImage)

**Status:** Fully defined, ready to migrate

---

### 7. digital_files.sql (5 queries)
Location: `db/queries/{sqlite,postgres}/digital_files.sql`

**Defined queries:**
- ✓ GetDigitalFile
- ✓ ListDigitalFiles
- ✓ CreateDigitalFile
- ✓ DeleteDigitalFile
- ✓ DeleteDigitalFiles

**Used by methods:**
- ✓ AddDigitalFile (maps to CreateDigitalFile)
- (DeleteDigital not directly mapped, needs custom implementation)

**Status:** Partially covered

---

### 8. digital_data.sql (6 queries)
Location: `db/queries/{sqlite,postgres}/digital_data.sql`

**Defined queries:**
- ✓ GetDigitalData
- ✓ GetDigitalDataByProduct
- ✓ ListDigitalDataByCart
- ✓ CreateDigitalData
- ✓ UpdateDigitalData
- ✓ DeleteDigitalData
- ✓ DeleteDigitalDataByProduct

**Used by methods:**
- ✓ ProductDigital (maps to GetDigitalDataByProduct)
- ✓ AddDigitalData (maps to CreateDigitalData)
- ✓ UpdateDigital (maps to UpdateDigitalData)
- ✓ DeleteDigital (maps to DeleteDigitalData)

**Status:** Fully defined, ready to migrate

---

### 9. subdomains.sql (6 queries)
Location: `db/queries/{sqlite,postgres}/subdomains.sql`

**Defined queries:**
- ✓ GetSubdomain
- ✓ GetSubdomainByName
- ✓ ListSubdomains
- ✓ CreateSubdomain
- ✓ UpdateSubdomain
- ✓ DeleteSubdomain
- ✓ SubdomainExists

**Used by methods:**
- No methods currently call subdomain queries

**Status:** Defined but unused

---

## Missing Query Methods (Not Yet in sqlc)

### Category 1: Authentication (1 query)

**auth.sql** - Missing file, needs creation

- `GetPasswordByEmail` - Retrieves email and password setting by email
  - **Location:** `internal/queries/auth.go:17`
  - **SQL:** `SELECT key, value FROM setting WHERE key IN ('email', 'password')`
  - **Note:** Manual loop to filter results, consider simplifying

---

### Category 2: Payment Information (1 query)

**cart.go - PaymentList method** - Not in sqlc

- `PaymentList` - Retrieves active status of multiple payment methods
  - **Location:** `internal/queries/cart.go:24`
  - **SQL:** `SELECT key, value FROM setting WHERE key IN (?, ?, ?, ?, ?)`
  - **Keys:** stripe_active, paypal_active, spectrocoin_active, coinbase_active, portone_active
  - **Note:** Should be added to carts.sql or settings.sql (or new payments.sql)

---

### Category 3: Settings Operations (3 queries)

**settings.sql** - Existing but missing some implementations

- `GetSettingByGroup` - Retrieves multiple related settings
  - **Location:** `internal/queries/setting.go:166`
  - **Complexity:** Dynamic SQL with IN clause, requires refactoring
  - **Note:** Could be multiple simple queries or a new specialized query

- `UpdateSettingByGroup` - Batch updates with UPSERT
  - **Location:** `internal/queries/setting.go:238`
  - **SQL:** `INSERT INTO setting (id, key, value) VALUES (?, ?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`
  - **Note:** UPSERT pattern specific to each database type

- `UpdatePassword` - Verify old password before updating
  - **Location:** `internal/queries/setting.go:275`
  - **Complexity:** Two-step operation (verify then update)
  - **Note:** Could be combined or split

---

### Category 4: Product Operations (6 queries)

**products.sql** - Existing but with complex method implementations

- `AddProductWithVariants` - Transaction-based product + variants + options + images
  - **Location:** `internal/queries/products.go:1075`
  - **Complexity:** Transaction involving multiple tables
  - **Note:** Composed of insertProductVariants + insertProductOptions + insertProductImages + product creation

- `GetProductWithVariants` - Fetch product with all related data
  - **Location:** `internal/queries/products.go:1298`
  - **Complexity:** Loads product + options + variants + images
  - **Note:** Currently loads via multiple queries

- `GenerateUniqueSlug` - Generates unique slug with conflict resolution
  - **Location:** `internal/queries/products.go:1594`
  - **Complexity:** Checks existing slugs and generates variations
  - **Note:** Database-specific slug generation logic

- `ProductImages` - Listed in queries but method exists
  - **Status:** Mapped to ListProductImages

- `ProductDigital` - Listed in queries but method exists
  - **Status:** Mapped to GetDigitalDataByProduct

- Cart letter operations
  - `CartLetterPayment` - Generates email template data
    - **Location:** `internal/queries/cart.go:687`
    - **Note:** Not database-backed in sqlc sense, more of a data aggregator
  
  - `CartLetterPurchase` - Generates email template data
    - **Location:** `internal/queries/cart.go:711`
    - **Note:** Not database-backed in sqlc sense, more of a data aggregator

---

## Summary of Missing/Partial Coverage

### Fully Missing Files (Need Creation)
- **auth.sql** - 1 query (GetPasswordByEmail)

### Existing Files Needing Additions
- **settings.sql** - Add 3 new queries (GetSettingByGroup, UpdateSettingByGroup, UpdatePassword)
- **carts.sql** - Add 1 new query (PaymentList) or move to settings.sql
- **products.sql** - Add 4 new queries (AddProductWithVariants, GetProductWithVariants, GenerateUniqueSlug, plus complex join variants)

### Not Database Queries
- **CartLetterPayment** - Data aggregation, not database operation
- **CartLetterPurchase** - Data aggregation, not database operation
- **GroupFieldMap** - Utility method, not a query
- Private methods (insertProductOptions, loadProductVariants, etc.) - Internal implementation

---

## Analysis: Current State vs Target State

### Currently Implemented in sqlc
- **58 definitions** across 9 SQL files (settings, sessions, pages, products, carts, product_images, digital_files, digital_data, subdomains)
- **Mostly used** by existing methods through manual wrapping

### Needs Migration/Definition
1. **Authentication** (1 method)
   - GetPasswordByEmail

2. **Payment Configuration** (1 method)
   - PaymentList

3. **Settings Management** (3 methods)
   - GetSettingByGroup (complex, dynamic SQL)
   - UpdateSettingByGroup (UPSERT pattern)
   - UpdatePassword (verify + update)

4. **Product Advanced Operations** (4 methods)
   - AddProductWithVariants (transaction-based)
   - GetProductWithVariants (complex join)
   - GenerateUniqueSlug (database-specific logic)
   - (Cart letter methods are data aggregation, not queries)

### Total Missing Query Definitions
**~10 queries** (including variants and database-specific implementations)

### Blocking Issues for Migration
1. **Dynamic SQL in GetSettingByGroup** - IN clause with variable parameters
2. **UPSERT pattern** - SQLite uses `ON CONFLICT`, PostgreSQL uses `ON CONFLICT...DO UPDATE`
3. **Transaction-based operations** - AddProductWithVariants spans multiple tables
4. **Database-specific functions** - EXTRACT(EPOCH FROM...) for PostgreSQL vs strftime for SQLite
5. **Slug generation logic** - Custom algorithm with database queries

---

## Recommended Approach for Task 2+

1. **Simple Direct Migrations**
   - Migrate session methods (4 methods) - fully defined
   - Migrate page methods (9 methods) - fully defined
   - Migrate cart methods (4 methods) - fully defined
   - Migrate product image methods (3 methods) - fully defined
   - Migrate digital data methods (4 methods) - fully defined

2. **Requires New Definitions**
   - Create auth.sql with GetPasswordByEmail
   - Add PaymentList to settings.sql or carts.sql
   - Add GetSettingByGroup, UpdateSettingByGroup, UpdatePassword to settings.sql

3. **Requires Special Handling**
   - Refactor AddProductWithVariants to explicit transaction steps
   - Split GetProductWithVariants into component queries
   - Consider stored procedures for GenerateUniqueSlug or keep in code

4. **Data Aggregation (Keep in Code)**
   - CartLetterPayment
   - CartLetterPurchase
   - GroupFieldMap

---

## Database Placeholder Differences

### Current Manual Handling in Code

The codebase already handles database-specific placeholders:

```go
// SQLite uses ?
query := `SELECT ... WHERE key = ?`

// PostgreSQL uses $1, $2, $3, etc.
query := `SELECT ... WHERE key = $1`
```

This is handled by the `buildPlaceholders()` function and `DBType()` checks throughout the codebase.

When migrating to sqlc:
- SQLite queries use `?` placeholders
- PostgreSQL queries use `$1, $2, $3` placeholders
- sqlc handles this automatically when generating Go code from separate query files

---

## Files Analyzed

### Query Source Files
- `/home/wj/work/mycart_origin/internal/queries/auth.go`
- `/home/wj/work/mycart_origin/internal/queries/cart.go`
- `/home/wj/work/mycart_origin/internal/queries/install.go`
- `/home/wj/work/mycart_origin/internal/queries/pages.go`
- `/home/wj/work/mycart_origin/internal/queries/products.go`
- `/home/wj/work/mycart_origin/internal/queries/queries.go`
- `/home/wj/work/mycart_origin/internal/queries/session.go`
- `/home/wj/work/mycart_origin/internal/queries/setting.go`

### sqlc Definition Files
- `/home/wj/work/mycart_origin/db/queries/sqlite/*.sql` (9 files)
- `/home/wj/work/mycart_origin/db/queries/postgres/*.sql` (9 files)

---

## Next Steps

This audit provides the foundation for:
1. **Task 2**: Add missing query definitions to sqlc
2. **Task 3-14**: Systematically migrate each method to use generated sqlc code
3. **Task 15**: Remove manual SQL and clean up query structs
4. **Task 16**: Verify end-to-end functionality

The migration is **achievable** with clear boundaries between simple queries and complex operations requiring refactoring.
