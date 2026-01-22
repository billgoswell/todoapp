# Query Optimization Testing Guide

This guide explains how to verify that the database query optimizations are working effectively.

## Setup

### Prerequisites
- PostgreSQL server running
- `commandlinetodo` database created and initialized
- `psql` command-line client available

### Initialize Test Data

```bash
# Connect to the database
psql -U postgres -d commandlinetodo

-- Create a test user
INSERT INTO users (api_key) VALUES ('test-key-123');

-- Get the user ID (usually 1 if first user)
SELECT id FROM users WHERE api_key = 'test-key-123';

-- Create test lists (assuming user_id = 1)
INSERT INTO todo_lists (user_id, client_id, name, updated_at, version)
VALUES
  (1, 'list-1', 'Work', NOW(), 1),
  (1, 'list-2', 'Personal', NOW(), 1),
  (1, 'list-3', 'Shopping', NOW(), 1);

-- Populate with test tasks (1000 tasks across lists)
INSERT INTO tasks (user_id, client_id, todo_list_id, todo, priority, done, date_added, updated_at, version, deleted)
SELECT
  1,
  'task-' || seq,
  ((seq - 1) % 3) + (SELECT MIN(id) FROM todo_lists WHERE user_id = 1),
  'Task ' || seq,
  ((seq - 1) % 4) + 1,
  (seq % 10) = 0,  -- 10% done
  NOW() - INTERVAL '1 day' * (seq / 10),
  NOW() - INTERVAL '1 day' * (seq / 100),
  1,
  FALSE
FROM generate_series(1, 1000) seq;
```

## Testing Index Usage

### 1. Verify Composite Index Usage (GetTasksSince Query)

Before optimization, the query might use separate indexes inefficiently:

```sql
-- Show query plan BEFORE optimization (single column indexes)
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, user_id, client_id, todo_list_id, todo, priority, done,
       date_added, date_completed, due_date, deleted, deleted_at,
       created_at, updated_at, version
FROM tasks
WHERE user_id = 1 AND updated_at >= to_timestamp(1609459200)
ORDER BY updated_at DESC
LIMIT 100;
```

**Expected Output (Before Optimization):**
```
Index Scan using idx_tasks_user_id on tasks  (cost=0.42..2847.81 rows=100 width=115)
  Index Cond: (user_id = 1)
  Filter: (updated_at >= '2021-01-01 00:00:00')  -- Sequential filter!
  Planning Time: 0.532 ms
  Execution Time: 245.231 ms
  Buffers: shared hit=1230 read=45
```

After adding the composite index `idx_tasks_user_updated`:

**Expected Output (After Optimization):**
```
Index Scan using idx_tasks_user_updated on tasks  (cost=0.42..847.81 rows=100 width=115)
  Index Cond: (user_id = 1 AND updated_at >= '2021-01-01 00:00:00')
  Planning Time: 0.231 ms
  Execution Time: 23.456 ms
  Buffers: shared hit=120 read=3
```

**Key Improvements:**
- Uses composite index (not sequential filter)
- Execution time: ~10x faster (245ms → 23ms)
- Buffers hit: Much lower (fewer pages read)

### 2. Verify Partial Index Usage (Active Items)

Test that partial indexes are used for active item queries:

```sql
-- Query for non-deleted tasks (uses partial index)
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, user_id, client_id, todo_list_id, todo, priority, done,
       date_added, date_completed, due_date, deleted, deleted_at,
       created_at, updated_at, version
FROM tasks
WHERE user_id = 1 AND updated_at >= to_timestamp(1609459200) AND deleted = FALSE
ORDER BY updated_at DESC
LIMIT 100;
```

**Expected Output (With Partial Index):**
```
Index Scan using idx_tasks_active on tasks  (cost=0.42..500.81 rows=95 width=115)
  Index Cond: (user_id = 1 AND updated_at >= '2021-01-01 00:00:00')
  Planning Time: 0.231 ms
  Execution Time: 15.123 ms
  Buffers: shared hit=105 read=2
```

**Key Improvements:**
- Partial index is smaller (excludes deleted items)
- Even faster execution than composite index
- Fewer buffer reads

### 3. Compare List Query Optimization

```sql
-- GetListsSince query
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, user_id, client_id, name, display_order, archived,
       created_at, updated_at, version
FROM todo_lists
WHERE user_id = 1 AND updated_at >= to_timestamp(1609459200)
ORDER BY updated_at DESC;
```

**Expected Improvement:**
- Uses `idx_lists_user_updated` composite index
- Execution time: 5-10ms (much faster than 100+ms before)

### 4. Test Task List Query (Dashboard)

```sql
-- Query for dashboard: tasks in a list with completion status
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, client_id, todo, priority, done, updated_at
FROM tasks
WHERE todo_list_id = 1 AND done = FALSE AND deleted = FALSE
ORDER BY priority ASC, updated_at DESC
LIMIT 50;
```

**Expected Output (With idx_tasks_by_list_status):**
```
Index Scan using idx_tasks_by_list_status on tasks  (cost=0.42..250.81 rows=50 width=72)
  Index Cond: (todo_list_id = 1 AND done = FALSE)
  Planning Time: 0.231 ms
  Execution Time: 8.234 ms
  Buffers: shared hit=45 read=1
```

## Performance Benchmarking

### Benchmark Script

Create `benchmark.sql`:

```sql
-- Warm up cache
SELECT COUNT(*) FROM tasks WHERE user_id = 1;
SELECT COUNT(*) FROM todo_lists WHERE user_id = 1;

-- Test 1: GetTasksSince (common sync query)
\timing on
SELECT COUNT(*) FROM tasks
WHERE user_id = 1 AND updated_at >= to_timestamp(1609459200);
\timing off

-- Test 2: GetListsSince (common sync query)
\timing on
SELECT COUNT(*) FROM todo_lists
WHERE user_id = 1 AND updated_at >= to_timestamp(1609459200);
\timing off

-- Test 3: Active tasks (filtered by deleted)
\timing on
SELECT COUNT(*) FROM tasks
WHERE user_id = 1 AND updated_at >= to_timestamp(1609459200) AND deleted = FALSE;
\timing off

-- Test 4: Dashboard query (tasks by list)
\timing on
SELECT COUNT(*) FROM tasks
WHERE todo_list_id = 1 AND done = FALSE AND deleted = FALSE;
\timing off
```

Run benchmark:

```bash
psql -U postgres -d commandlinetodo -f benchmark.sql
```

**Expected Results:**
- Test 1 & 2: 5-25ms execution time
- Test 3: 3-15ms execution time
- Test 4: 2-10ms execution time

### Compare With/Without Indexes

To test without indexes (for comparison):

```sql
-- Temporarily disable indexes (for testing only)
ALTER INDEX idx_tasks_user_updated UNUSABLE;
ALTER INDEX idx_lists_user_updated UNUSABLE;
ALTER INDEX idx_tasks_active UNUSABLE;
ALTER INDEX idx_lists_active UNUSABLE;

-- Run benchmark again
\i benchmark.sql

-- Re-enable indexes
ALTER INDEX idx_tasks_user_updated USABLE;
ALTER INDEX idx_lists_user_updated USABLE;
ALTER INDEX idx_tasks_active USABLE;
ALTER INDEX idx_lists_active USABLE;
```

## Monitoring Query Performance in Production

### PostgreSQL pg_stat_statements Extension

Enable query statistics:

```sql
-- Enable extension (requires superuser)
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- View slowest queries
SELECT query, mean_exec_time, calls, total_exec_time
FROM pg_stat_statements
WHERE query LIKE '%tasks%' OR query LIKE '%todo_lists%'
ORDER BY mean_exec_time DESC
LIMIT 10;

-- View queries by total time spent
SELECT query, total_exec_time, calls, mean_exec_time
FROM pg_stat_statements
WHERE query LIKE '%WHERE%'
ORDER BY total_exec_time DESC
LIMIT 10;

-- Reset statistics
SELECT pg_stat_statements_reset();
```

### Index Usage Statistics

Check if indexes are being used:

```sql
-- View index sizes and usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE tablename IN ('tasks', 'todo_lists')
ORDER BY idx_scan DESC;

-- Identify unused indexes
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE tablename IN ('tasks', 'todo_lists')
    AND idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;
```

### Cache Hit Ratio

```sql
-- Check cache hit ratio (should be > 99%)
SELECT
    sum(heap_blks_read) as heap_read,
    sum(heap_blks_hit) as heap_hit,
    sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as ratio
FROM pg_statio_user_tables
WHERE relname IN ('tasks', 'todo_lists');
```

## Validation Checklist

After deploying query optimizations, verify:

- [ ] Migration runs successfully (new indexes created)
- [ ] Server starts and connects to database
- [ ] GetTasksSince returns correct results
- [ ] GetListsSince returns correct results
- [ ] Sync operations complete successfully
- [ ] EXPLAIN ANALYZE shows index scans (not sequential scans)
- [ ] Execution times are 5-10x faster than before
- [ ] No slow queries appear in pg_stat_statements
- [ ] Cache hit ratio > 99%
- [ ] All indexes are being used (idx_scan > 0)

## Troubleshooting

### Index Not Being Used

If EXPLAIN ANALYZE shows sequential scan instead of index scan:

```sql
-- Check index statistics are up to date
ANALYZE tasks;
ANALYZE todo_lists;

-- Re-run EXPLAIN ANALYZE
EXPLAIN (ANALYZE, BUFFERS)
SELECT ... FROM tasks WHERE user_id = 1 AND updated_at >= ...;

-- Check if query planner is configured correctly
SHOW random_page_cost;  -- Usually 4.0
SHOW seq_page_cost;     -- Usually 1.0

-- If still not using index, check index is not corrupted
REINDEX INDEX idx_tasks_user_updated;
```

### Slow Index Scan

If index scan is slow, verify:

1. **Index exists and is not being disabled:**
   ```sql
   SELECT schemaname, tablename, indexname, indisvalid, indisready
   FROM pg_indexes
   WHERE indexname LIKE 'idx_tasks%' OR indexname LIKE 'idx_lists%';
   ```

2. **Statistics are current:**
   ```sql
   ANALYZE tasks;
   ANALYZE todo_lists;
   ```

3. **Index has not bloated:**
   ```sql
   SELECT
       relname,
       pg_size_pretty(pg_relation_size(oid)) as size,
       (pg_relation_size(oid) - pg_relation_size(oid, 'main')) as bloat
   FROM pg_class
   WHERE relname LIKE 'idx_tasks%' OR relname LIKE 'idx_lists%';
   ```

### Partial Index Not Matching Query

Partial indexes are only used if the WHERE clause in the query matches the partial index condition:

```sql
-- This uses idx_tasks_active (matches partial condition)
SELECT * FROM tasks WHERE user_id = 1 AND deleted = FALSE;

-- This does NOT use idx_tasks_active (doesn't check deleted = FALSE)
SELECT * FROM tasks WHERE user_id = 1;

-- To use partial index, explicitly check the partial condition
SELECT * FROM tasks WHERE user_id = 1 AND deleted = FALSE;
```

## References

- PostgreSQL EXPLAIN: https://www.postgresql.org/docs/current/sql-explain.html
- Index Types: https://www.postgresql.org/docs/current/indexes-types.html
- Partial Indexes: https://www.postgresql.org/docs/current/indexes-partial.html
- pg_stat_statements: https://www.postgresql.org/docs/current/pgstatstatements.html
