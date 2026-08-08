# JSON Scanning Error Fix

## Issue Fixed
**Error:** `sql: Scan error on column index 6, name "cart": unsupported Scan, storing driver.Value type string into type *json.RawMessage`

## Root Cause
Both PostgreSQL (`github.com/lib/pq`) and SQLite (`modernc.org/sqlite`) drivers return JSON/JSONB columns as `string` type, but `json.RawMessage` (which is `[]byte`) cannot scan from strings directly - it only accepts byte slices.

## Solution
Modified SQL queries to explicitly cast JSON columns to TEXT, then handle string-to-bytes conversion in application code.

### Files Modified

#### 1. `/db/queries/postgres/carts.sql`
```sql
-- Before:
SELECT ... cart, ... FROM cart

-- After:
SELECT ... cart::text as cart, ... FROM cart
```

#### 2. `/db/queries/sqlite/carts.sql`
```sql
-- Before:
SELECT ... cart, ... FROM cart

-- After:
SELECT ... CAST(cart AS TEXT) as cart, ... FROM cart
```

#### 3. Regenerated sqlc code
```bash
go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
```

Result: `GetCartRow.Cart` changed from `json.RawMessage` to `string` in both:
- `internal/db/postgres/carts.sql.go`
- `internal/db/sqlite/carts.sql.go`

#### 4. `/internal/queries/cart.go`
Updated Cart() method to convert string to []byte before unmarshalling:

```go
// SQLite case:
if len(v.Cart) > 0 {
    if err := json.Unmarshal([]byte(v.Cart), &cart.Cart); err != nil {
        return nil, err
    }
}

// PostgreSQL case:
if len(v.Cart) > 0 {
    if err := json.Unmarshal([]byte(v.Cart), &cart.Cart); err != nil {
        return nil, err
    }
}
```

## Tests Fixed
- ✅ `TestGetCart` - All subtests passing
- ✅ `TestCarts` - All subtests passing
- ✅ All cart-related handlers now work with both SQLite and PostgreSQL

## Impact
- **Database Compatibility:** Works uniformly across SQLite and PostgreSQL
- **Migration Safety:** No schema changes required
- **Backward Compatible:** Existing data continues to work
- **sqlc Integration:** Clean solution using sqlc's type mapping

## Technical Notes
- `json.RawMessage` is type alias for `[]byte`, not a distinct type with custom Scan()
- Both database drivers return JSON as string to preserve exact representation
- Explicit cast ensures consistent behavior across database engines
- String-to-bytes conversion is zero-copy in Go (just reinterprets the backing array)
