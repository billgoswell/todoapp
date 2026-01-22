# CommandLineTodo - Improvement and Testing Plan

## Executive Summary

The CommandLineTodo monorepo consists of three applications (CLI, Server, Mobile) with **minimal test coverage (3.95%)** and several areas for improvement. This plan outlines a phased approach to increase code quality, test coverage, and fix known issues across all components.

**Current State:**
- **Total test coverage:** 3.95% (3 test files out of 76 source files)
- **todo-server:** Only conversion helpers tested; zero endpoint tests
- **todo-cmdline:** Only date parsing tested; sync logic untested
- **todo-mobile:** Zero tests (Jest configured but unused)
- **Known issues:** 1 critical timestamp bug in sync operations

## Phase 1: Critical Fixes & Foundation (Weeks 1-2)

### 1.1 Fix Known Issues

#### Issue: Incorrect Timestamp in Sync Payload
**File:** `todo-cmdline/cmd/app/syncclient.go:221`
**Severity:** Critical
**Description:** `UpdatedAt` field uses `dateAdded` instead of actual `updated_at` timestamp, breaking conflict resolution

**Action Items:**
- [ ] Add `updated_at` field to todoItem struct in `model.go`
- [ ] Track `updated_at` when creating/modifying items in `handlers.go`
- [ ] Update sync payload in `syncclient.go` to use actual `updated_at`
- [ ] Add test to verify `updated_at` is sent correctly in sync requests
- [ ] Test conflict resolution with multiple device updates

**Test:** Create `cmd/app/syncclient_test.go` with test case `TestSyncPayloadTimestamp`

### 1.2 Set Up Testing Infrastructure

#### Server Testing
**File:** `todo-server/internal/handlers/handlers_test.go` (new)
**Goal:** Establish testing patterns for HTTP handlers

```go
// Test fixtures
- CreateTestDB() - in-memory PostgreSQL or test container
- CreateTestRouter() - Gin router with middleware
- CreateTestAuthToken() - Valid API key for tests

// Table-driven test pattern
func TestGetTasks(t *testing.T) {
    tests := []struct{
        name string
        apiKey string
        expectedStatus int
        expectedCount int
    }
}
```

**Test Coverage Target:** All 8 HTTP handlers

#### CLI Testing
**File:** `todo-cmdline/cmd/app/handlers_test.go` (new)
**Goal:** Test TUI event handlers and data operations

**Test Coverage Target:** Core handler logic (create, update, delete, list operations)

#### Mobile Testing
**File:** `todo-mobile/src/__tests__/` (new directory structure)
**Goal:** Establish Jest testing for React Native components

```
src/__tests__/
├── api/
├── database/
├── sync/
├── state/
└── components/
```

## Phase 2: Server Testing (Weeks 2-3)

### 2.1 Unit Tests for Handlers

**Target:** 100% handler coverage

**tests to implement:**

```bash
# Authentication & Health
- TestHealth_ReturnsOK
- TestAuthMiddleware_ValidToken
- TestAuthMiddleware_InvalidToken
- TestAuthMiddleware_MissingToken

# Tasks
- TestGetTasks_ReturnsAllTasks
- TestGetTasks_FiltersByStatus
- TestCreateTask_ValidPayload
- TestCreateTask_InvalidPayload
- TestUpdateTask_ExistingTask
- TestUpdateTask_NonExistentTask
- TestDeleteTask_SoftDeletesTask
- TestDeleteTask_IsMarkedAsDeleted

# Lists
- TestGetLists_ReturnsAllLists
- TestCreateList_ValidPayload
- TestCreateList_DuplicateName
- TestUpdateList_ArchiveList
- TestDeleteList_SoftDelete

# Sync
- TestPull_SinceTimestamp
- TestPull_ReturnsChangedItems
- TestPush_CreatesNewItems
- TestPush_UpdatesExistingItems
- TestPush_ConflictResolution_LastWriteWins
- TestPush_HandlesSoftDeletes
```

### 2.2 Database Layer Tests

**Target:** `todo-server/internal/db/postgres_test.go`

```bash
# Setup/Teardown
- TestNewDB_ConnectsSuccessfully
- TestMigrate_CreatesSchema

# Tasks
- TestCreateTask_ReturnsID
- TestGetTask_ByID
- TestGetTasks_ByListID
- TestUpdateTask_VersionConflict
- TestDeleteTask_MarkDeleted

# Lists
- TestCreateList_ReturnsID
- TestGetList_ByID
- TestArchiveList_UpdatesStatus

# Sync Metadata
- TestUpdateSyncMetadata
- TestGetLastSyncTime
```

### 2.3 Integration Tests

**File:** `todo-server/tests/integration_test.go`

**Workflow tests:**
- [ ] Create list → Create task → Update task → Mark complete
- [ ] Simultaneous updates from multiple clients (conflict resolution)
- [ ] Pull sync returns correct changes since timestamp
- [ ] Push sync merges client changes with server state
- [ ] Deleted items propagate correctly

## Phase 3: CLI Testing (Weeks 3-4)

### 3.1 Handler Tests

**File:** `todo-cmdline/cmd/app/handlers_test.go`

```bash
# Add Item
- TestAddItem_CreatesInDB
- TestAddItem_SetsDefaults
- TestAddItem_TriggersSync
- TestAddItem_InvalidInput

# Update Item
- TestMarkDone_UpdatesStatus
- TestEditItem_UpdatesFields
- TestSetPriority_ValidRange

# Delete Item
- TestDeleteItem_RemovesFromUI
- TestDeleteItem_MarksDeleted

# List Management
- TestCreateList_AddsToModel
- TestSwitchList_UpdatesSelection
- TestArchiveList_HidesFromView

# Search/Filter
- TestFilterByStatus
- TestFilterByPriority
- TestSearchByText

# Config
- TestLoadConfig_CreatesDefault
- TestLoadConfig_ParsesJSON
- TestLoadConfig_ValidatesFields
```

### 3.2 Sync Store Tests (Expand Existing)

**File:** `todo-cmdline/cmd/app/syncstore_test.go` (enhance)

Add tests for:
- [ ] `FullSync()` - complete pull/push sequence
- [ ] `PushChanges()` - batches and sends pending changes
- [ ] `PullChanges()` - fetches and merges server changes
- [ ] Conflict resolution with realistic scenarios
- [ ] Offline queueing and retry logic
- [ ] Background sync scheduling

### 3.3 Database Tests

**File:** `todo-cmdline/cmd/app/db_test.go` (new)

```bash
# Schema
- TestInitDB_CreatesSchema

# CRUD Operations
- TestCreateItem_InDB
- TestGetItems_FromDB
- TestUpdateItem_Persistence
- TestDeleteItem_SoftDelete

# Query Operations
- TestGetItemsByList
- TestGetItemsByStatus
- TestGetItemsByPriority
- TestSearchItems
```

## Phase 4: Mobile Testing (Weeks 4-5)

### 4.1 API Client Tests

**File:** `todo-mobile/src/__tests__/api/client.test.ts`

```bash
# Authentication
- TestSetApiKey_StoresLocally
- TestSetServerUrl_Validates

# Requests
- TestMakeRequest_IncludesAuthHeader
- TestMakeRequest_HandlesErrors
- TestMakeRequest_RetryOnFailure
- TestMakeRequest_TimeoutHandling

# Endpoints
- TestHealthCheck
- TestGetLists
- TestCreateList
- TestUpdateList
- TestDeleteList
- TestGetTasks
- TestCreateTask
- TestUpdateTask
- TestDeleteTask
- TestPull
- TestPush
```

### 4.2 Sync Engine Tests

**File:** `todo-mobile/src/__tests__/sync/sync.test.ts`

```bash
# Sync Coordinator
- TestStartSync_OnlineMode
- TestStartSync_OfflineMode
- TestPull_MergesChanges
- TestPush_SendsPendingChanges
- TestConflictResolution_LastWriteWins

# Change Log
- TestLogChange_OnCreate
- TestLogChange_OnUpdate
- TestLogChange_OnDelete
- TestMarkSynced_ClearsLog

# Background Sync
- TestScheduleSync_StartsTimer
- TestSyncInterval_RespectsDuration
- TestAutoSyncOnChange_OnItemCreate
- TestAutoSyncOnChange_OnItemUpdate
```

### 4.3 Database Tests

**File:** `todo-mobile/src/__tests__/database/db.test.ts`

```bash
# Schema
- TestInitializeDatabase_CreatesSchema

# Lists
- TestCreateList_ReturnsID
- TestGetLists_ReturnsAll
- TestUpdateList_Persists
- TestDeleteList_SoftDelete

# Tasks
- TestCreateTask_AssociatesWithList
- TestGetTasks_ByListID
- TestUpdateTask_UpdatesFields
- TestDeleteTask_MarksDeleted

# Sync Metadata
- TestStoreSyncMetadata
- TestRetrieveSyncMetadata
```

### 4.4 Component Tests

**File:** `todo-mobile/src/__tests__/components/`

```bash
# Core Components
- TestHomeScreen_DisplaysLists
- TestHomeScreen_ListOperations
- TestTaskListScreen_DisplaysTasks
- TestTaskListScreen_CRUD
- TestTaskDetailScreen_EditForm
- TestTaskDetailScreen_Validation

# UI Components
- TestTaskItem_DisplaysCorrectly
- TestTaskItem_InteractionHandling
- TestListItem_DisplaysCorrectly
- TestDueDatePicker_SelectsDate
- TestSearchFilter_FiltersResults
- TestPriorityBadge_DisplaysLevel

# State
- TestUseTasksProvider_StateUpdates
- TestUseTasksProvider_PersistenceSync
```

## Phase 5: Integration & E2E Testing (Weeks 5-6)

### 5.1 Cross-Application Sync Tests

**File:** `tests/e2e/sync.test.ts` (new)

**Scenarios to test:**
- [ ] Create task on CLI → Sync to server → Appear on mobile
- [ ] Update task on mobile → Sync to server → Update on CLI
- [ ] Delete on CLI → Sync to server → Marked deleted on mobile
- [ ] Offline edits on mobile → Online sync → Conflict resolution
- [ ] Simultaneous edits on CLI and mobile → Server resolves correctly
- [ ] Network failure → Queue changes → Resume on reconnect

### 5.2 Load & Performance Tests

**Server:** `todo-server/tests/performance_test.go`
- [ ] Handle 1000 concurrent sync requests
- [ ] 10,000 tasks per user - list retrieval performance
- [ ] Bulk upload 500 tasks - push sync performance
- [ ] Database query optimization - measure slow queries

**Mobile:** `todo-mobile/src/__tests__/performance/`
- [ ] Render 1000 tasks in list - frame rate maintenance
- [ ] Sync 500 items - memory usage
- [ ] Database query performance with large datasets

**CLI:** `todo-cmdline/cmd/app/performance_test.go`
- [ ] Render 1000 items in TUI - responsiveness
- [ ] Database queries under load
- [ ] Sync performance with large datasets

## Phase 6: Code Quality & Documentation (Weeks 6-7)

### 6.1 Code Coverage Analysis

**Tools:**
- `go test -cover` for server and CLI
- `jest --coverage` for mobile

**Targets:**
- [ ] Server: Achieve 70%+ coverage
- [ ] CLI: Achieve 65%+ coverage
- [ ] Mobile: Achieve 60%+ coverage
- [ ] All: Maintain 100% coverage for sync logic

### 6.2 Error Handling & Edge Cases

**Audit and test:**
- [ ] Invalid JSON payloads
- [ ] Missing required fields
- [ ] Corrupted database
- [ ] Network timeouts
- [ ] Concurrent modifications
- [ ] Database constraint violations
- [ ] Large file uploads/downloads
- [ ] Unicode and special characters

### 6.3 Documentation Updates

**Create/Update:**
- [ ] API documentation (OpenAPI/Swagger spec)
- [ ] Sync protocol specification
- [ ] Database schema documentation
- [ ] Conflict resolution algorithm documentation
- [ ] Testing guide for contributors
- [ ] Troubleshooting guide

## Implementation Strategy

### Repository Structure for Tests

```
todo-server/
├── internal/handlers/handlers_test.go
├── internal/db/postgres_test.go
├── tests/
│   ├── integration_test.go
│   └── performance_test.go

todo-cmdline/cmd/app/
├── handlers_test.go
├── db_test.go
├── syncstore_test.go (enhance)
├── syncclient_test.go
└── utils_test.go (existing)

todo-mobile/src/
├── __tests__/
│   ├── api/client.test.ts
│   ├── sync/sync.test.ts
│   ├── database/db.test.ts
│   ├── state/useTasksProvider.test.ts
│   ├── components/
│   │   ├── HomeScreen.test.tsx
│   │   ├── TaskListScreen.test.tsx
│   │   ├── TaskDetailScreen.test.tsx
│   │   └── ...
│   └── integration/sync.test.ts
├── setupTests.ts (Jest config)
└── testUtils.tsx (helpers)
```

### Testing Best Practices

1. **Use table-driven tests** (especially in Go)
2. **Test both happy path and error cases**
3. **Use descriptive test names** following `Test<Function>_<Scenario>` pattern
4. **Mock external dependencies** (API calls, database, network)
5. **Use fixtures** for consistent test data
6. **Keep tests focused** - one behavior per test
7. **Run tests in isolation** - no shared state between tests
8. **Measure and monitor** code coverage trends

### CI/CD Integration

Add to GitHub Actions (or your CI tool):
```yaml
- Run: go test -v -cover ./...  # Server & CLI
- Run: npm test -- --coverage   # Mobile
- Report: Upload coverage to Codecov
- Fail: If coverage drops below target
- Fail: If any test fails
```

## Success Criteria

| Phase | Target | Success Metric |
|-------|--------|-----------------|
| Phase 1 | Critical fixes | All known issues resolved |
| Phase 2 | Server tests | 70%+ coverage, all handlers tested |
| Phase 3 | CLI tests | 65%+ coverage, all handlers tested |
| Phase 4 | Mobile tests | 60%+ coverage, critical paths tested |
| Phase 5 | Integration | E2E sync workflows validated |
| Phase 6 | Quality | Maintainable, documented codebase |

## Risk & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Test data setup complexity | Delays in Phase 2-4 | Create reusable fixtures early (Phase 1) |
| Mobile Jest setup issues | Blocks Phase 4 | Research React Native testing patterns immediately |
| Database test isolation | Flaky tests | Use transactions/rollback for each test |
| Merge conflicts with active development | Integration nightmare | Create feature branch, merge incrementally |
| Performance test infrastructure | Hard to reproduce | Use consistent test environment |

## Estimated Effort

- **Phase 1:** 2-3 days (critical fixes + infrastructure setup)
- **Phase 2:** 5-7 days (server testing + integration)
- **Phase 3:** 5-7 days (CLI testing + database tests)
- **Phase 4:** 7-10 days (mobile testing + component tests)
- **Phase 5:** 5-7 days (E2E + performance testing)
- **Phase 6:** 3-5 days (coverage review + documentation)

**Total:** 4-6 weeks of focused development

## Next Steps

1. **Week 1 (Immediate):**
   - Fix timestamp bug in `syncclient.go`
   - Set up test infrastructure for all three apps
   - Create test fixtures and utilities
   - Establish testing guidelines

2. **Week 2-3:**
   - Implement server handler tests
   - Expand CLI sync tests
   - Begin mobile API client tests

3. **Week 4-5:**
   - Complete mobile testing
   - Add integration tests
   - Performance testing

4. **Week 6:**
   - Coverage analysis and gap filling
   - Documentation updates
   - CI/CD integration

## References

- `TESTING.md` - API testing guide (existing)
- `DEPLOYMENT_PLAN.md` - Production deployment info
- `DEVELOPER_GUIDE.md` - Mobile development patterns
- `CONFIG.md` - CLI configuration

---

**Created:** 2026-01-20
**Status:** Ready for implementation
**Owner:** Development team
