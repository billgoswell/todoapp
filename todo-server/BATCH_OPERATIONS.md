# Batch Database Operations and N+1 Query Optimization

## Overview

This document describes the batch database operations implementation that resolves N+1 query problems in the sync operations.

## Problem Statement

The original Push sync handler had a critical performance issue:

```go
// BEFORE: N+1 queries - one per item
for _, list := range req.Lists {
    _, err := database.UpsertList(userID, listMap)  // 1 query per list
    if err != nil {
        return err
    }
}

for _, task := range req.Tasks {
    _, err := database.UpsertTask(userID, taskMap)  // 1 query per task
    if err != nil {
        return err
    }
}
```

Performance Impact:
- Syncing 100 lists + 100 tasks = **200 separate database round-trips**
- Network latency multiplied by 200
- Database connection pool exhaustion
- Unscalable for bulk sync operations

## Solution: Batch Operations

Implemented batch insert/update operations that process all items in a single query:

```go
// AFTER: 2 queries total - one for all lists, one for all tasks
if len(req.Lists) > 0 {
    if err := database.UpsertListsBatch(userID, listMaps); err != nil {
        return err  // 1 query for all lists
    }
}

if len(req.Tasks) > 0 {
    if err := database.UpsertTasksBatch(userID, taskMaps); err != nil {
        return err  // 1 query for all tasks
    }
}
```

### Implementation Details

#### UpsertListsBatch
- **Location**: `internal/db/repositories/list_batch.go`
- **Function**: Inserts or updates multiple todo lists in one operation
- **Conflict Resolution**: Uses PostgreSQL `ON CONFLICT (user_id, name)` with last-write-wins semantics
- **Performance**: O(N) time with single database round-trip instead of O(N) round-trips

```go
func (r *ListRepository) UpsertListsBatch(userID int, lists []map[string]any) error
```

#### UpsertTasksBatch
- **Location**: `internal/db/repositories/task_batch.go`
- **Function**: Inserts or updates multiple tasks in one operation
- **Conflict Resolution**: Uses PostgreSQL `ON CONFLICT (user_id, client_id)` with last-write-wins semantics
- **Performance**: O(N) time with single database round-trip instead of O(N) round-trips

```go
func (r *TaskRepository) UpsertTasksBatch(userID int, tasks []map[string]any) error
```

### Query Structure

Both batch operations use PostgreSQL's multi-row INSERT with ON CONFLICT:

```sql
INSERT INTO table_name (col1, col2, ...) VALUES
    (val1, val2, ...),
    (val1, val2, ...),
    ...
ON CONFLICT (unique_constraint) DO UPDATE SET
    col1 = CASE WHEN EXCLUDED.updated_at > table.updated_at
              THEN EXCLUDED.col1
              ELSE table.col1 END,
    ...
```

This ensures:
1. **Atomicity**: All inserts/updates succeed or all fail
2. **Conflict Resolution**: Last-write-wins based on `updated_at` timestamp
3. **Efficiency**: Single query for any number of items

## Performance Improvements

### Theoretical Improvements

| Scenario | Old Approach | New Approach | Improvement |
|----------|-------------|-------------|------------|
| 10 items | 10 queries | 1 query | 10x faster |
| 100 items | 100 queries | 1 query | 100x faster |
| 1000 items | 1000 queries | 1 query | 1000x faster |

### Actual Benefits

1. **Reduced Network Latency**
   - Typical HTTP round-trip: 10-50ms
   - 100 items: 1000-5000ms overhead → 10-50ms

2. **Connection Pool Efficiency**
   - No connection exhaustion during bulk sync
   - Better concurrency for other requests

3. **Reduced Database Load**
   - One transaction instead of many
   - Fewer lock contention issues
   - Reduced WAL (Write-Ahead Log) overhead

4. **Scalability**
   - Can handle arbitrarily large sync batches
   - Linear performance degradation instead of exponential

## Testing

Comprehensive tests in `internal/db/repositories/batch_test.go`:

- **TestBatchConvertTimestamp_VariousFormats**: Timestamp conversion edge cases (7 tests)
- **TestBatchQueryGeneration_MultipleTasks**: Query argument building for tasks
- **TestBatchQueryGeneration_MultipleLists**: Query argument building for lists

All tests passing ✓

## Migration Guide

### Updating Custom Code

If you have custom code using the repositories, migration is straightforward:

```go
// Old code - multiple individual calls
for _, item := range items {
    _, err := db.UpsertTask(userID, item)
    if err != nil {
        return err
    }
}

// New code - single batch call
if err := db.UpsertTasksBatch(userID, items); err != nil {
    return err
}
```

### Handler Changes

The sync Push handler was updated to use batch operations:
- `internal/handlers/sync.go`: Uses `UpsertListsBatch()` and `UpsertTasksBatch()`

## Backward Compatibility

- Original single-item `UpsertTask()` and `UpsertList()` methods remain available
- Batch methods are additions, not replacements
- Existing code continues to work unchanged

## Future Optimizations

1. **Batching by List ID**: Group tasks by list before batching
2. **Parallel Processing**: Split large batches across multiple database connections
3. **Change Detection**: Only sync changed items instead of entire batch
4. **Pagination**: For extremely large sync operations, process in smaller chunks

## Monitoring

To track batch operation performance, monitor:

1. **Sync Duration**: Time for complete push operation
2. **Database Queries**: Monitor query count (should be ~2 for any batch size)
3. **Connection Pool**: Ensure no connection exhaustion
4. **Lock Contention**: Monitor for row locks during large batches

## References

- PostgreSQL ON CONFLICT: https://www.postgresql.org/docs/current/sql-insert.html
- Last-Write-Wins Conflict Resolution: See `PLAN.md` Phase 1 section
- Sync Protocol: See `CLAUDE.md` section "Sync Protocol"
