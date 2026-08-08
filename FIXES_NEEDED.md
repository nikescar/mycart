# Specific Code Fixes Needed for PostgreSQL

## 🔧 Critical Fixes (Must Do)

### Fix 1: Empty IN Clause Bug
**File:** `internal/queries/cart.go`  
**Problem:** Building SQL with empty array: `WHERE id IN ()`

**Current code pattern:**
```go
// Somewhere in cart.go - need to find exact location
productIDs := []string{}
query := fmt.Sprintf("SELECT * FROM products WHERE id IN (%s)", 
    strings.Join(productIDs, ","))
```

**Fix:**
```go
// Add this check before building the query
if len(productIDs) == 0 {
    return nil, errors.ErrProductNotFound
}

// Or use a safer query builder
if len(productIDs) == 0 {
    // Return empty result or error
    return []Product{}, nil
}

// Build query only if we have IDs
placeholders := make([]string, len(productIDs))
args := make([]interface{}, len(productIDs))
for i, id := range productIDs {
    placeholders[i] = fmt.Sprintf("$%d", i+1)
    args[i] = id
}
query := fmt.Sprintf("SELECT * FROM products WHERE id IN (%s)", 
    strings.Join(placeholders, ","))
```

---

### Fix 2: Null Pointer Panic in ListProducts
**File:** `internal/queries/products.go`  
**Line:** 35  
**Problem:** Interface conversion panic when column contains NULL

**Current code (around line 35):**
```go
// This line is causing panic - need to see actual code
someField := row["field"].(string) // Panic if NULL
```

**Fix:**
```go
// Safe conversion with null check
var someField string
if val, ok := row["field"].(string); ok {
    someField = val
} else if row["field"] == nil {
    someField = "" // or handle NULL appropriately
} else {
    // Log unexpected type
    someField = fmt.Sprintf("%v", row["field"])
}

// Or use sql.NullString
var someField sql.NullString
err := rows.Scan(&someField)
if someField.Valid {
    // Use someField.String
}
```

---

### Fix 3: Test Database Cleanup
**File:** `internal/queries/testutil.go`

**Add this function:**
```go
// CleanupTestDB removes all test data from the database
func CleanupTestDB(t *testing.T, db *sql.DB) {
    t.Helper()
    
    // Order matters due to foreign keys
    tables := []string{
        "cart_item",       // Must delete first (has FK to cart)
        "cart",
        "product_digital",
        "product_variant",
        "product_image",
        "product",
        "page",
        // Don't delete settings table - needed for tests
    }
    
    for _, table := range tables {
        // PostgreSQL: DELETE is faster than TRUNCATE for small datasets
        query := fmt.Sprintf("DELETE FROM %s", table)
        if _, err := db.Exec(query); err != nil {
            // Only warn, don't fail - table might not exist
            t.Logf("warning: failed to clean %s: %v", table, err)
        }
    }
    
    // Reset sequences for PostgreSQL
    sequences := []string{
        "cart_item_id_seq",
        // Add other sequences if needed
    }
    
    for _, seq := range sequences {
        query := fmt.Sprintf("ALTER SEQUENCE IF EXISTS %s RESTART WITH 1", seq)
        if _, err := db.Exec(query); err != nil {
            t.Logf("warning: failed to reset %s: %v", seq, err)
        }
    }
}
```

**Update existing test setup:**
```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    
    db := getTestDB(t)
    
    // Clean before tests
    CleanupTestDB(t, db)
    
    // Register cleanup after tests
    t.Cleanup(func() {
        CleanupTestDB(t, db)
    })
    
    return db
}
```

---

### Fix 4: Install Handler Connection Issue
**File:** `internal/handlers/private/install.go`

**Problem:** Database connection closing before install completes

**Look for:**
```go
// Likely issue: closing connection too early
defer db.Close() // This might be in wrong place
```

**Fix:**
```go
// Don't close the global DB connection in handlers
// Only close connections you explicitly create

// If creating a new connection:
conn, err := db.Conn(ctx)
if err != nil {
    return err
}
defer conn.Close() // OK to close this

// But global app.DB should NOT be closed in handlers
```

---

### Fix 5: Webhook SQL Syntax
**File:** Check files in `internal/webhook/`

**Problem:** SQL query has syntax error with comma

**Pattern to find:**
```go
// Look for queries with trailing commas or missing fields
query := `
    INSERT INTO table (
        field1,
        field2,  // <-- trailing comma
    ) VALUES ($1, $2)
`
```

**Fix:**
```go
// Remove trailing commas
query := `
    INSERT INTO table (
        field1,
        field2
    ) VALUES ($1, $2)
`

// Or use query builder
```

---

## 🔍 How to Find These Issues

### Search for Empty IN Clauses
```bash
cd internal/queries
grep -n "IN (" *.go
# Look for any place building IN clauses dynamically
```

### Search for Type Assertions
```bash
cd internal/queries  
grep -n "\.(string)" *.go
grep -n "\.(int)" *.go
# Check each one for nil safety
```

### Search for Database Close Calls
```bash
cd internal
grep -rn "\.Close()" --include="*.go" handlers/
# Review each to ensure not closing global connection
```

---

## 🧪 Testing Your Fixes

### Test Individual Package
```bash
# Test queries package
./test.sh package ./internal/queries -v

# Test specific test
go test -v ./internal/queries -run TestValidateCartItems_ProductNotFound
```

### Test Failed Packages Only
```bash
./test.sh failed -v
```

### Full Test Suite
```bash
./test.sh postgres
```

---

## 📝 Verification Checklist

After making fixes, verify:

- [ ] All SQL queries use parameterized placeholders ($1, $2, etc.)
- [ ] No empty IN clauses (check array length first)
- [ ] All type assertions have nil checks
- [ ] No trailing commas in SQL queries
- [ ] Test cleanup runs before AND after each test
- [ ] Global DB connections not closed in handlers
- [ ] All tests pass in isolation: `go test -count=1`
- [ ] Tests pass with race detector: `go test -race`
- [ ] Tests are repeatable: `go test -count=10`

---

## 🎯 Expected Results After Fixes

### Before (Current)
```
FAIL    internal                     5.941s
FAIL    internal/handlers/private   23.279s
FAIL    internal/mailer             19.336s
FAIL    internal/queries            61.636s (PANIC)
FAIL    internal/webhook            11.792s
```

### After (Target)
```
ok      internal                     2.5s
ok      internal/handlers/private   8.2s
ok      internal/mailer             6.1s
ok      internal/queries           12.3s
ok      internal/webhook            3.2s
```

---

## 🚀 Implementation Order

1. **Fix test cleanup** (easiest, helps all other tests)
   - Update `testutil.go`
   - Run: `./test.sh failed`

2. **Fix empty IN clause** (most critical)
   - Update `internal/queries/cart.go`
   - Run: `go test ./internal/queries -run TestValidateCartItems`

3. **Fix nil pointer** (causes panic)
   - Update `internal/queries/products.go:35`
   - Run: `go test ./internal/queries -run TestProduct`

4. **Fix install handlers** (connection lifecycle)
   - Update `internal/handlers/private/install.go`
   - Run: `go test ./internal/handlers/private -run TestInstall`

5. **Fix webhook SQL** (SQL syntax)
   - Update webhook SQL queries
   - Run: `go test ./internal/webhook`

6. **Full test run**
   - Run: `./test.sh postgres`

Total estimated time: **2-4 hours**
