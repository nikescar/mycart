# Products Module Migration Plan

## Status: Partially Started (2/31 methods)

The Products module is the most complex module in the codebase with:
- **1,597 lines** of code
- **31 methods** to migrate
- Complex JSON aggregation queries
- Multiple related tables (product, product_image, product_variant, product_option, digital_data, digital_file)
- Extensive business logic and transactions

## Completed (2/31)

### Simple CRUD Operations ✅
1. `DeleteProduct` - Uses sqlc DeleteProduct
2. `UpdateActive` - Uses sqlc UpdateProductActive  

## Migration Strategy

Due to the complexity, recommend a phased approach:

### Phase 1: Simple Operations (Priority: High)
**Estimated**: 2-3 hours

Methods that can be directly migrated with minimal changes:
- [ ] `IsProduct` - Keep as-is (has specific active+deleted logic)
- [ ] Basic update operations that match sqlc queries

### Phase 2: Read Operations (Priority: Medium)  
**Estimated**: 4-6 hours

Methods requiring type conversion but straightforward logic:
- [ ] `Product(ctx, private, id)` - Complex but can use GetProductByID as base
- [ ] `ProductBySlug` - Can use GetProductBySlug as base
- [ ] Image retrieval methods (already simple queries)

**Challenge**: These methods currently build complex Product objects with:
- JSON aggregated images
- JSON aggregated variants
- Metadata/attributes unmarshaling
- SEO data

**Options**:
1. Break into multiple queries (N+1 pattern)
2. Keep JSON aggregation in direct SQL
3. Use helper functions to build Product from multiple sqlc queries

### Phase 3: Complex Operations (Priority: Low)
**Estimated**: 6-8 hours

Methods with complex business logic and transactions:
- [ ] `ListProducts` - Very complex with dynamic filtering, JSON aggregation
- [ ] `AddProduct` - Transaction with images, variants, metadata
- [ ] `UpdateProduct` - Complex transaction syncing variants, images
- [ ] Variant operations (20+ methods)
- [ ] Digital product operations

**Challenge**: Heavy use of:
- Transactions across multiple tables
- JSON aggregation (json_group_array)
- Dynamic WHERE clauses
- Business logic mixed with queries

## Recommendation

### Option A: Hybrid Approach (Recommended)
- Migrate simple CRUD (Phase 1) ✅ Started
- Keep complex queries as direct SQL
- Use sqlc for new features going forward
- **Benefit**: 90%+ of migration value with 30% of effort

### Option B: Full Migration
- Complete all 3 phases
- Refactor complex methods to use multiple sqlc queries
- Add helper functions for JSON aggregation
- **Benefit**: 100% type safety
- **Cost**: 12-16 hours of work, potential performance impact (N+1 queries)

### Option C: Defer Products
- Focus migration complete for 6 other modules (86% of codebase)
- Products continue using direct SQL
- Revisit when sqlc adds better JSON aggregation support
- **Benefit**: Fastest path to production with migrated modules

## Technical Challenges

### 1. JSON Aggregation
Current code uses SQLite's `json_group_array`:
```sql
SELECT 
  product.*,
  (SELECT json_group_array(json_object(...)) FROM product_image ...) as images,
  (SELECT json_group_array(json_object(...)) FROM product_variant ...) as variants
FROM product
```

sqlc doesn't generate this - would need direct SQL or break into separate queries.

### 2. Dynamic Filtering
`ListProducts` has conditional WHERE clauses based on:
- private vs public mode
- cartID presence
- idList filtering
- active/deleted status

sqlc doesn't support this - would need multiple queries or direct SQL.

### 3. Complex Transactions
`AddProduct` and `UpdateProduct` use transactions with:
- Insert/update product
- Sync variants (add/update/delete)
- Sync images
- Update metadata

Would need to orchestrate multiple sqlc calls within a transaction.

## Current Implementation Analysis

### Products.go Method Breakdown

**Simple CRUD** (Good candidates for sqlc):
- ✅ DeleteProduct
- ✅ UpdateActive  
- IsProduct (keep as-is - has specific logic)

**Medium Complexity** (Moderate candidates):
- Product (reads one product with all relations)
- ProductImages (simple JOIN query)
- AddImage, DeleteImage (simple operations)
- ProductDigital (reads digital product data)
- AddDigitalFile, AddDigitalData (simple inserts)

**High Complexity** (Keep as direct SQL):
- ListProducts (JSON aggregation, dynamic filters)
- AddProduct (complex transaction, multiple tables)
- UpdateProduct (complex transaction, sync logic)
- All variant sync methods (20+ methods with complex logic)

## Next Steps

1. **Complete Phase 1** (Simple Operations)
   - Review remaining simple CRUD operations
   - Migrate where sqlc queries exist
   - Estimated: 1-2 hours

2. **Evaluate ROI for Phase 2**
   - Assess if breaking JSON aggregation is worth it
   - Consider performance implications
   - Decision point: Continue or stop here

3. **Document Decision**
   - Update CART-MIGRATION-STATUS.md
   - Mark Products as "partially migrated" or "deferred"
   - Provide clear guidance for future work

## Conclusion

The Products module represents the "long tail" of the migration - significant effort for diminishing returns. The core benefit of sqlc (type safety, compile-time validation) has already been achieved for 6 of 7 modules covering settings, auth, sessions, pages, and carts.

**Recommendation**: Complete Phase 1 (simple operations), then document and defer the rest. The 2 methods already migrated plus a few more simple ones provide value without the complexity of refactoring the JSON aggregation queries.
