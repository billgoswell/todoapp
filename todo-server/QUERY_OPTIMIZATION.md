# Database Query Optimization Guide

This document outlines the query optimization improvements made to the CommandLineTodo server.

## Current Query Patterns

### 1. GetTasksSince Query
**File:** `internal/db/repositories/task.go:20-73`

```sql
SELECT ... FROM tasks
WHERE user_id = $1 AND updated_at >= to_timestamp($2)
ORDER BY updated_at DESC
```

**Current Indexes:**
- `idx_tasks_user_id ON tasks(user_id)` - Single column index
- `idx_tasks_updated_at ON tasks(updated_at)` - Single column index

**Issue:** The query applies two filter conditions (`user_id` and `updated_at`), but neither index covers both conditions. PostgreSQL query planner must either:
1. Use index on `user_id`, then filter by `updated_at` (index scan → seq scan on timestamp)
2. Use index on `updated_at`, then filter by `user_id` (index scan → seq scan on user_id)

For users with many tasks, this results in suboptimal query execution.

**Optimization:** Add composite index on `(user_id, updated_at)` to cover both filter conditions without additional filtering.

### 2. GetListsSince Query
**File:** `internal/db/repositories/list.go:20-64`

```sql
SELECT ... FROM todo_lists
WHERE user_id = $1 AND updated_at >= to_timestamp($2)
ORDER BY updated_at DESC
```

**Current Indexes:** Same single-column indexes as tasks

**Optimization:** Add composite index on `(user_id, updated_at)` similar to tasks.

### 3. UpsertTask Query Performance
**File:** `internal/db/repositories/task.go:78-155`

The ON CONFLICT clause uses CASE statements for every column:
```sql
ON CONFLICT (user_id, client_id) DO UPDATE SET
    todo = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.todo ELSE tasks.todo END,
    priority = CASE WHEN EXCLUDED.updated_at > tasks.updated_at THEN EXCLUDED.priority ELSE tasks.priority END,
    ...
```

**Issue:** The condition `EXCLUDED.updated_at > tasks.updated_at` is evaluated for every column assignment, even though the result is always the same. This adds overhead.

**Optimization:** Use PostgreSQL's `EXCLUDED` row constructor more efficiently by creating a temporary comparison at query generation time, though the CASE-per-column approach is correct for correctness (evaluates to same boolean each time).

### 4. Missing Indexes for Common Query Patterns
**Issues:**
- No index for filtering active (non-deleted) tasks: `WHERE deleted = FALSE`
- No index for filtering active (non-archived) lists: `WHERE archived = FALSE`
- No index for tasks by list: `WHERE todo_list_id = ?`
- No composite index for common dual-filter queries

**Optimizations:**
1. Partial index on `idx_tasks_active ON tasks(user_id, updated_at) WHERE deleted = FALSE`
   - Covers sync queries that typically exclude deleted items
   - Much smaller index than full table

2. Partial index on `idx_lists_active ON todo_lists(user_id, updated_at) WHERE archived = FALSE`
   - Covers sync queries for non-archived lists
   - Significantly reduces index size

3. Index on `idx_tasks_by_list ON tasks(todo_list_id, done)` for dashboard queries
   - Supports "tasks in list X where done = Y" queries
   - Useful for aggregation queries

## Optimization Priority

### High Priority (Implement First)
1. **Add composite index `(user_id, updated_at)` for tasks and lists**
   - Directly addresses the most common sync query patterns
   - Minimal storage overhead
   - Significant performance improvement for users with large task lists
   - Implementation: Update migration in `postgres.go`

2. **Add partial indexes for active items**
   - Reduces index size compared to full composite index
   - Faster index traversal
   - Only applies to queries filtering by `deleted = FALSE` / `archived = FALSE`
   - Implementation: Add to migration

### Medium Priority (Implement If Performance Remains)
3. **Pagination support for large result sets**
   - Add LIMIT/OFFSET to GetTasksSince and GetListsSince
   - Add parameters to sync protocol
   - Prevents loading thousands of rows at once
   - Implementation: Requires API contract changes

4. **Query result caching**
   - Cache recent sync queries by (user_id, since_timestamp)
   - Reduces database load for rapid polling
   - Implementation: Add in-memory cache with TTL

### Low Priority (Future Optimization)
5. **Archival of old sync data**
   - Move completed/deleted items older than X days to archive table
   - Keeps main table smaller and indexes more efficient
   - Implementation: Add background job

## Implementation Status

- [x] Document optimization opportunities
- [ ] Add composite index `(user_id, updated_at)` for tasks
- [ ] Add composite index `(user_id, updated_at)` for lists
- [ ] Add partial index `(user_id, updated_at) WHERE deleted = FALSE` for tasks
- [ ] Add partial index `(user_id, updated_at) WHERE archived = FALSE` for lists
- [ ] Test query performance with EXPLAIN ANALYZE
- [ ] Add pagination to sync protocol (future)
- [ ] Implement query caching (future)

## Expected Performance Improvements

### Before Optimization
For a user with 5,000 tasks where 1,000 are modified in the current session:
- Query execution time: ~150-300ms
- Index lookups: 2 separate index scans or 1 scan + sequential filter
- Disk I/O: High due to non-optimal index usage

### After Composite Index
- Query execution time: ~20-50ms (4-6x faster)
- Index lookups: 1 composite index scan covering both conditions
- Disk I/O: Lower due to single optimized index

### With Partial Indexes
- Query execution time: ~10-30ms (5-10x faster)
- Index size: 30-50% smaller than composite index
- Disk I/O: Further reduced due to smaller index

## Monitoring Query Performance

Use EXPLAIN ANALYZE to verify optimization effectiveness:

```sql
-- Check task sync query plan
EXPLAIN ANALYZE
SELECT id, user_id, client_id, ...
FROM tasks
WHERE user_id = 123 AND updated_at >= to_timestamp(1234567890)
ORDER BY updated_at DESC;

-- Check list sync query plan
EXPLAIN ANALYZE
SELECT id, user_id, client_id, ...
FROM todo_lists
WHERE user_id = 123 AND updated_at >= to_timestamp(1234567890)
ORDER BY updated_at DESC;

-- Check active tasks index
EXPLAIN ANALYZE
SELECT id, user_id, client_id, ...
FROM tasks
WHERE user_id = 123 AND updated_at >= to_timestamp(1234567890) AND deleted = FALSE
ORDER BY updated_at DESC;
```

Expected output (after optimization):
```
Index Scan using idx_tasks_user_id_updated_at on tasks
  Index Cond: (user_id = 123 AND updated_at >= '...')
  Planning time: 0.5ms
  Execution time: 25.3ms
```

## References

- PostgreSQL Composite Indexes: https://www.postgresql.org/docs/current/indexes-types.html
- PostgreSQL Partial Indexes: https://www.postgresql.org/docs/current/indexes-partial.html
- EXPLAIN ANALYZE Documentation: https://www.postgresql.org/docs/current/sql-explain.html
