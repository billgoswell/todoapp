# Server and Sync Testing Guide

## Overview

This document describes how to test the command-line todo server and client synchronization system.

**Phase 2 Update**: Comprehensive testing infrastructure with hybrid approach supporting both local database containers and live server integration testing.

## Test Results Summary

✅ **All Unit Tests Passing**
- Handler conversion tests (4/4 passing)
- Client sync logic tests (8/8 passing)
- Total: 12/12 tests passing

## Server Setup for Manual Testing

### Prerequisites

- PostgreSQL 12+ running on `localhost:5432`
- Server built from `cmd/server/main.go`
- Client built from `cmd/app/`

### Step 1: Set Up PostgreSQL

```bash
# Using Docker
docker run -d \
  --name todo-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=commandlinetodo \
  -p 5432:5432 \
  postgres:15-alpine

# Wait a few seconds for the database to start
sleep 5
```

### Step 2: Create Server User

```bash
# Create a test user (you'll need access to psql)
psql -h localhost -U postgres -d commandlinetodo -c \
  "INSERT INTO users (api_key) VALUES ('test-api-key-123');"
```

Or use the API endpoint (once server is running):

```bash
# First, bootstrap with a temporary API key, then create permanent ones via API
```

### Step 3: Start the Server

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=commandlinetodo
export SERVER_HOST=127.0.0.1
export SERVER_PORT=8080

# Run the server
cd commandlinetodo-server
go build -o server ./cmd/server/main.go
./server
```

## Testing the API Endpoints

### Health Check Endpoint

Test that the server is running:

```bash
curl -X GET http://localhost:8080/api/v1/health
```

Expected response:
```json
{"status":"ok"}
```

### Authentication

All other endpoints require Bearer token authentication:

```bash
# Example API key (you would create real ones)
API_KEY="test-api-key-123"

# All requests should include Authorization header:
curl -H "Authorization: Bearer $API_KEY" ...
```

### Test Scenario 1: Create and Sync a Todo List

**Step 1: Create a list on the client**

```bash
# Client creates a list (simulated with curl)
API_KEY="test-api-key-123"

curl -X POST http://localhost:8080/api/v1/lists \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client-list-001",
    "name": "My Shopping List",
    "display_order": 0,
    "archived": false,
    "created_at": 1699564800,
    "updated_at": 1699564800,
    "version": 1
  }'
```

**Step 2: Verify list was created**

```bash
curl -X GET http://localhost:8080/api/v1/lists \
  -H "Authorization: Bearer $API_KEY"
```

Expected response includes the created list.

### Test Scenario 2: Create and Sync a Task

**Step 1: Create a task**

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client-task-001",
    "todo_list_id": 1,
    "todo": "Buy milk",
    "priority": 3,
    "done": false,
    "date_added": 1699564800,
    "date_completed": null,
    "due_date": 1699651200,
    "deleted": false,
    "deleted_at": null,
    "created_at": 1699564800,
    "updated_at": 1699564800,
    "version": 1
  }'
```

**Step 2: Retrieve all tasks**

```bash
curl -X GET http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $API_KEY"
```

### Test Scenario 3: Pull Changes from Server

Simulates a client pulling the latest changes:

```bash
curl -X POST http://localhost:8080/api/v1/sync/pull \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "since": 0
  }'
```

Expected response: Tasks and lists modified since timestamp 0 (all items)

### Test Scenario 4: Push Changes to Server

Simulates a client pushing local changes:

```bash
curl -X POST http://localhost:8080/api/v1/sync/push \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {
        "client_id": "client-task-002",
        "todo_list_id": 1,
        "todo": "Buy bread",
        "priority": 2,
        "done": false,
        "date_added": 1699568400,
        "date_completed": null,
        "due_date": null,
        "deleted": false,
        "deleted_at": null,
        "created_at": 1699568400,
        "updated_at": 1699568400,
        "version": 1
      }
    ],
    "lists": []
  }'
```

Expected response:
```json
{"status":"ok"}
```

### Test Scenario 5: Conflict Resolution (Last Write Wins)

**Step 1: Create an item with version 1**

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "client-task-003",
    "todo_list_id": 1,
    "todo": "Original task",
    "priority": 3,
    "done": false,
    "date_added": 1699564800,
    "date_completed": null,
    "due_date": null,
    "deleted": false,
    "deleted_at": null,
    "created_at": 1699564800,
    "updated_at": 1699564800,
    "version": 1
  }'
```

**Step 2: Simulate conflicting updates from two clients**

Client A updates (newer timestamp - 1699654400):
```bash
curl -X POST http://localhost:8080/api/v1/sync/push \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {
        "client_id": "client-task-003",
        "todo_list_id": 1,
        "todo": "Updated by client A",
        "priority": 3,
        "done": false,
        "date_added": 1699564800,
        "date_completed": null,
        "due_date": null,
        "deleted": false,
        "deleted_at": null,
        "created_at": 1699564800,
        "updated_at": 1699654400,
        "version": 2
      }
    ],
    "lists": []
  }'
```

**Step 3: Verify the newer version wins**

Pull the changes:
```bash
curl -X POST http://localhost:8080/api/v1/sync/pull \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"since": 1699564800}'
```

The response should show "Updated by client A" as the task is at the newer timestamp.

### Test Scenario 6: Update a Task

```bash
curl -X PUT http://localhost:8080/api/v1/tasks/1 \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "client_id": "client-task-001",
    "todo_list_id": 1,
    "todo": "Buy milk (2L)",
    "priority": 2,
    "done": true,
    "date_added": 1699564800,
    "date_completed": 1699568400,
    "due_date": 1699651200,
    "deleted": false,
    "deleted_at": null,
    "created_at": 1699564800,
    "updated_at": 1699568400,
    "version": 2
  }'
```

### Test Scenario 7: Delete a Task (Soft Delete)

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/1 \
  -H "Authorization: Bearer $API_KEY"
```

Verify the task is marked as deleted:
```bash
curl -X GET http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $API_KEY"
```

Task should still appear with `"deleted": true`.

### Test Scenario 8: Archive a List

```bash
curl -X PUT http://localhost:8080/api/v1/lists/1 \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "client_id": "client-list-001",
    "name": "My Shopping List",
    "display_order": 0,
    "archived": true,
    "created_at": 1699564800,
    "updated_at": 1699568400,
    "version": 2
  }'
```

## Client Sync Testing

### Enable Sync in Client

Set environment variables before running the client:

```bash
export TODO_SYNC_ENABLED=true
export TODO_SYNC_SERVER_URL=http://localhost:8080/api/v1
export TODO_SYNC_API_KEY=test-api-key-123
export TODO_SYNC_INTERVAL=10
export TODO_SYNC_AUTO_SYNC_ON_CHANGE=true
```

Or create a config file at `~/.config/commandlinetodo/config.json`:

```json
{
  "database": {
    "path": "~/.local/share/commandlinetodo/todo.db"
  },
  "sync": {
    "enabled": true,
    "server_url": "http://localhost:8080/api/v1",
    "api_key": "test-api-key-123",
    "device_id": "device-001",
    "sync_interval_seconds": 10,
    "auto_sync_on_change": true,
    "retry_attempts": 3,
    "timeout_seconds": 10
  }
}
```

### Run Client with Sync

```bash
export TODO_SYNC_ENABLED=true
export TODO_SYNC_SERVER_URL=http://localhost:8080/api/v1
export TODO_SYNC_API_KEY=test-api-key-123
./todo
```

### Verify Sync Behavior

1. **Initial Sync**: Client should pull any existing data from server on startup
2. **Auto-Sync**: When creating/modifying tasks, client should sync after 10 seconds
3. **Manual Sync**: Press 's' key to manually trigger a sync
4. **Sync Status**: UI should show sync status (online/offline, last sync time, pending changes)

## Expected Sync Flow

### Pull (Client gets server changes)

```
1. Client sends POST /api/v1/sync/pull with {since: lastSyncTime}
2. Server returns all tasks and lists modified since that timestamp
3. Client applies "last write wins" conflict resolution
4. Local database is updated with server changes
```

### Push (Client sends local changes)

```
1. Client reads change_log table for unsynced changes
2. Client sends POST /api/v1/sync/push with {tasks: [...], lists: [...]}
3. Server uses PostgreSQL ON CONFLICT to merge changes
4. Server only updates if client's updated_at > server's updated_at
5. Client marks changes as synced in change_log
```

### Full Sync Sequence

```
1. Check connectivity
2. If online:
   a. Pull changes from server
   b. Apply conflict resolution
   c. Push local changes
   d. Update last_sync_time
3. If offline:
   a. Log changes locally
   b. Wait for connectivity
```

## Troubleshooting

### Connection Refused

- Ensure PostgreSQL is running on localhost:5432
- Ensure server is running on localhost:8080
- Check environment variables are set correctly

### Authentication Failed

- Verify API key is correct
- Ensure Authorization header is in format: `Bearer <api-key>`
- Check that user exists in database

### Sync Not Happening

- Verify TODO_SYNC_ENABLED is set to true
- Check that server URL is correct
- Review client logs for sync errors
- Verify TODO_SYNC_INTERVAL is appropriate (10 seconds for testing)

### Database Migration Errors

- Drop and recreate the test database
- Ensure PostgreSQL version is 12+
- Check server logs for specific SQL errors

---

## Phase 2: Automated Testing (New)

### Hybrid Testing Approach

Phase 2 introduces a comprehensive automated testing framework supporting both:

1. **Local Database Tests** - Isolated testcontainers-go PostgreSQL instances
2. **Live Server Integration Tests** - Against running server at 192.168.100.20:8080

### Quick Start

#### Run Local Database Tests (Default)

```bash
# All local tests
go test -v ./internal/tests -run TestLocalDB -timeout 120s

# Specific test
go test -v ./internal/tests -run TestLocalDB_CreateTask
```

#### Run Live Server Tests

```bash
# Set environment variables
export TEST_USE_LOCAL_SERVER=false
export TEST_API_KEY=your-api-key-here

# Run tests
go test -v ./internal/tests -run TestLiveServer
```

### Test Structure

**Local Database Tests** (`internal/tests/local_db_test.go`):
- TestLocalDB_Health - Connection verification
- TestLocalDB_CreateUser - User creation/retrieval
- TestLocalDB_CreateList - List creation/retrieval
- TestLocalDB_CreateTask - Task creation/retrieval
- TestLocalDB_ConflictResolution_LastWriteWins - Last-write-wins logic
- TestLocalDB_MultipleUsers - User data isolation
- TestLocalDB_SyncTimestamp - Timestamp-based filtering
- TestLocalDB_UpdateTask - Task updates

**Live Server Tests** (`internal/tests/integration_test.go`):
- TestLiveServer_Health - Health endpoint check
- TestLiveServer_GetTasks - Retrieve tasks
- TestLiveServer_GetLists - Retrieve lists
- TestLiveServer_CreateTask - Create task on server
- TestLiveServer_SyncPull - Pull sync operation
- TestLiveServer_AuthRequired - Auth validation

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `TEST_USE_LOCAL_SERVER` | `true` | Local DB or live server |
| `TEST_LIVE_SERVER_HOST` | `192.168.100.20` | Server hostname |
| `TEST_LIVE_SERVER_PORT` | `8080` | Server port |
| `TEST_API_KEY` | `` | API key for live tests |
| `TEST_TIMEOUT` | `30` | HTTP request timeout (seconds) |

### Test Configuration

Test configuration is managed in `internal/tests/config.go`:

```go
config := GetTestConfig()
client := NewIntegrationTestClient(config)
ltdb := SetupLocalTestDB(t)
```

### Test Helpers

**Local Database Setup**:
```go
ltdb := SetupLocalTestDB(t)
defer ltdb.Cleanup()

userID := ltdb.CreateTestUser(t, "test-key")
listID := ltdb.CreateTestList(t, userID, "My List")
taskID := ltdb.CreateTestTask(t, userID, listID, "Do it")
```

**Live Server Client**:
```go
client := NewIntegrationTestClient(config)

healthy, _ := client.GetHealth()
tasks, _ := client.GetTasks()
list, _ := client.CreateList(&models.TodoList{...})
resp, _ := client.Pull(since)
client.Push(&models.PushRequest{...})
```

### Files Created/Modified

**New Files**:
- `internal/tests/config.go` - Test configuration
- `internal/tests/helpers.go` - Test helpers and fixtures
- `internal/tests/integration_test.go` - Live server tests
- `internal/tests/local_db_test.go` - Local database tests

**Modified**:
- `internal/db/postgres_test.go` - Enhanced with testcontainers setup
- `internal/handlers/handlers_test.go` - Error handling tests
- `internal/middleware/auth_test.go` - Auth test structure
- `go.mod` - Added testcontainers-go dependency

### Running Tests in CI/CD

```bash
# Local tests (no external dependencies)
go test -v -timeout 120s ./internal/tests -run TestLocalDB

# Live server tests (if available)
export TEST_USE_LOCAL_SERVER=false
export TEST_API_KEY=$CI_API_KEY
go test -v ./internal/tests -run TestLiveServer
```

### Coverage Analysis

```bash
# Generate coverage for database layer
go test -v -coverprofile=coverage.out ./internal/db ./internal/tests
go tool cover -html=coverage.out
```

### Next Steps

- Implement remaining handler tests (17 tests)
- Implement auth middleware tests (6 tests)
- Implement sync operation tests (8 tests)
- Achieve 70%+ test coverage target
