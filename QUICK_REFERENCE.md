# Quick Reference: Testing the ClientID Implementation

## Run All Tests (Local)

```bash
cd todo-cmdline
go test ./cmd/app -v
```

**Expected Output:** All tests pass ✅

## Run Specific New Tests

```bash
# Database schema test
go test ./cmd/app -run TestDBSchemaWithListClientID -v

# Sync payload test  
go test ./cmd/app -run TestTaskPayloadWithListClientID -v
```

## Build Both Binaries

```bash
# Client
cd todo-cmdline && go build -o app cmd/app/main.go

# Server
cd ../todo-server && go build -o server cmd/server/main.go
```

## Verify Schema Changes

### Client Database (SQLite)
```sql
-- Shows the list_client_id column
SELECT sql FROM sqlite_master WHERE name='tasks';
```

### Server Database (PostgreSQL)
```sql
-- When available, run:
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name='tasks' AND column_name='todo_list_client_id';
```

## Key Implementation Points

### What Changed
1. **Task.TodoListClientID** - New stable reference field
2. **tasks.list_client_id** - New database column (both client & server)
3. **TaskPayload.TodoListClientID** - New sync protocol field
4. **Server FK Resolution** - Resolves clientID → server ID before insert

### What Stayed the Same
1. **TodoListID** - Still used for local database constraints
2. **API Protocol** - Backward compatible
3. **Offline Support** - Works exactly the same
4. **Conflict Resolution** - Last-write-wins unchanged

### How It Works

```
Client Creates Task Offline:
  task.todoListID = 1 (local)
  task.listClientID = "list-uuid-001" (stable)
  ↓
  Syncs to Server:
    task.TodoListClientID = "list-uuid-001" (stable)
  ↓
  Server Receives:
    1. Looks up list by clientID "list-uuid-001"
    2. Finds server's list.id = 42
    3. Inserts task with todo_list_id = 42 (correct FK!)
    ✅ SUCCESS - No FK violation
```

## Test Files Structure

```
todo-cmdline/cmd/app/
├── db_test.go                    # Schema + migration tests
├── sync_integration_test.go      # Sync payload tests
├── *_test.go (existing)          # 38+ existing unit tests
└── (all pass with new changes)
```

## Changes at a Glance

| File | Change | Type |
|------|--------|------|
| shared-types/go/models/models.go | Add TodoListClientID | Field |
| todo-server/internal/db/postgres.go | Add column + method | Schema |
| todo-server/internal/db/repositories/list.go | Add GetIDByClientID() | Method |
| todo-server/internal/handlers/sync.go | FK resolution logic | Handler |
| todo-cmdline/cmd/app/db.go | Migration + CRUD | Database |
| todo-cmdline/cmd/app/syncclient.go | Add to payload | Sync |
| todo-cmdline/cmd/app/handlers.go | Set listClientID | Handler |
| todo-cmdline/cmd/app/datastore.go | Add interface | Interface |
| and 5 more... | Various | Various |

## Verification Checklist

- ✅ All 40+ tests pass
- ✅ Client compiles without errors
- ✅ Server compiles without errors
- ✅ Database schema includes list_client_id
- ✅ TaskPayload includes TodoListClientID
- ✅ Sync handler has FK resolution
- ✅ Migrations are idempotent
- ✅ Backward compatible (TodoListID kept)
- ✅ No breaking changes
- ✅ Error handling graceful

## Next Steps for Production

1. **Start PostgreSQL**
2. **Run server with PostgreSQL**
3. **Create test list and tasks**
4. **Verify server database has correct FK values**
5. **Test offline creation → online sync**
6. **Stress test with multiple lists**

## Common Commands

```bash
# Fresh client database
rm todo-cmdline/todo.db

# Run tests with verbose output
go test ./cmd/app -v -timeout 30s

# Build for production
go build -ldflags="-s -w" -o app cmd/app/main.go

# Check what changed
git diff shared-types/go/models/models.go
git diff todo-server/internal/handlers/sync.go
git diff todo-cmdline/cmd/app/db.go
```

## Status

```
✅ Implementation: COMPLETE
✅ Testing: PASSED (40+/40+)
✅ Build: SUCCESSFUL
✅ Ready for: Integration Testing
```

---

**For detailed information, see:**
- TEST_RESULTS.md - Comprehensive test results
- TESTING_SUMMARY.md - Executive summary
- Implementation Plan - Full technical details
