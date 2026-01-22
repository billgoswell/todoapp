# CLI Reorganization Plan (Phase 2.2)

**Current Status:** In Progress
**Complexity:** High - Involves 19 files with many inter-dependencies
**Estimated Effort:** 10-12 hours

---

## Current Structure

```
cmd/app/
├── main.go                        # Entry point
├── config.go                      # Configuration loading
├── constants.go                   # Constants and key bindings
├── db.go                          # Database initialization (14.7 KB)
├── datastore.go                   # Data operations interface
├── logger.go                      # Logging utilities
├── model.go                       # Data models (Task, TodoList)
├── handlers.go                    # TUI event handlers (13.6 KB)
├── render.go                      # Terminal UI rendering (9.9 KB)
├── syncclient.go                  # HTTP client for server (8.3 KB)
├── syncstore.go                   # Sync wrapper (11 KB)
├── utils.go                       # Utility functions
├── state_management_test.go       # State tests
├── handlers_test.go               # Handler tests
├── handlers_comprehensive_test.go # Comprehensive handler tests
├── syncclient_test.go             # Sync client tests
├── syncclient_comprehensive_test.go # Comprehensive sync tests
├── syncstore_test.go              # Sync store tests
└── utils_test.go                  # Utils tests
```

**Total:** 19 files, 5600+ lines in single directory

---

## Target Structure

```
internal/
├── app/
│   ├── manager.go               # App manager (extracted from main, handlers)
│   ├── model.go                 # Move model.go here
│   └── types.go                 # Exported types for inter-package use
├── handlers/
│   ├── input.go                 # Keyboard/input handlers (extracted from handlers.go)
│   ├── state.go                 # State management (extracted from handlers.go)
│   └── sync_handler.go          # Sync-related handlers (extracted from handlers.go)
├── ui/
│   ├── renderer.go              # Move render.go here
│   ├── colors.go                # Color constants (extracted from constants.go)
│   └── messages.go              # UI message constants (extracted from constants.go)
├── sync/
│   ├── client.go                # Move syncclient.go here
│   ├── store.go                 # Move syncstore.go here
│   └── types.go                 # Sync-related types
├── storage/
│   ├── sqlite.go                # Move db.go here
│   ├── datastore.go             # Move datastore.go here
│   └── schema.go                # Database schema (extracted from db.go)
├── config/
│   ├── config.go                # Move config.go here
│   └── loader.go                # Configuration loading logic
├── state/
│   ├── model.go                 # App state model
│   ├── state.go                 # State management
│   └── flow.go                  # Task creation flow (extracted from model.go)
├── logger/
│   └── logger.go                # Move logger.go here
└── main.go                      # Entry point
```

---

## Migration Steps

### Phase 1: Create Base Packages
- [x] Create directory structure
- [x] Move and refactor config.go to internal/config/
- [ ] Move logger.go to internal/logger/
- [ ] Move model.go to internal/state/

### Phase 2: Storage/Database Layer
- [ ] Move db.go to internal/storage/sqlite.go
- [ ] Move datastore.go to internal/storage/datastore.go
- [ ] Extract schema to internal/storage/schema.go
- [ ] Update imports

### Phase 3: Sync Layer
- [ ] Move syncclient.go to internal/sync/client.go
- [ ] Move syncstore.go to internal/sync/store.go
- [ ] Create internal/sync/types.go
- [ ] Update imports

### Phase 4: UI Layer
- [ ] Move render.go to internal/ui/renderer.go
- [ ] Extract UI constants to internal/ui/colors.go
- [ ] Extract UI messages to internal/ui/messages.go
- [ ] Update imports

### Phase 5: Handlers and Input
- [ ] Create internal/handlers/input.go (keyboard handlers)
- [ ] Create internal/handlers/state.go (state handlers)
- [ ] Create internal/handlers/sync_handler.go (sync handlers)
- [ ] Update imports

### Phase 6: App Management
- [ ] Create internal/app/manager.go
- [ ] Create internal/app/types.go
- [ ] Update main.go to use new structure
- [ ] Update cmd/app/main.go

### Phase 7: Test Organization
- [ ] Move and update test files to match package structure
- [ ] Update import paths in tests
- [ ] Verify all tests pass

### Phase 8: Documentation
- [ ] Create internal/README.md
- [ ] Document package boundaries
- [ ] Create migration guide

---

## Key Dependencies to Manage

### Circular Dependencies to Avoid
- `handlers` → `storage` (OK)
- `handlers` → `ui` (OK)
- `handlers` → `app` (OK)
- ~~`storage` → `handlers`~~ (AVOID)
- ~~`ui` → `handlers`~~ (AVOID - UI should be passive)

### Import Strategy
1. `internal/storage/` - No dependencies on other internal packages
2. `internal/config/` - No dependencies on other internal packages
3. `internal/logger/` - No dependencies on other internal packages
4. `internal/sync/` - Depends on `storage`, `config`
5. `internal/ui/` - Depends on `app` (for models), `state`
6. `internal/handlers/` - Depends on `storage`, `sync`, `ui`, `app`, `state`
7. `internal/app/` - Depends on `storage`, `sync`, `handlers`, `ui`, `state`

---

## File Mapping Details

### handlers.go → internal/handlers/
Original handlers.go (13.6 KB) should be split:

**input.go** - ~5 KB
- handleKeyInput()
- handleInputChar()
- handleInputBackspace()
- handleInputDelete()
- handleInputTab()
- handleInputEnter()

**state.go** - ~4 KB
- handleState()
- returnToMain()
- setState()

**sync_handler.go** - ~4 KB
- Sync-related handlers

### model.go → internal/state/
Original model.go (5 KB) should be split:

**model.go** - Core model struct
**flow.go** - TaskCreationFlow logic
**state.go** - State-related types

### render.go → internal/ui/
Move as-is with minor adjustments

### db.go → internal/storage/
Split db.go (14.7 KB):

**sqlite.go** - DB connection, pool management (~2 KB)
**schema.go** - Schema definition (~3 KB)
**datastore.go** - Data operations (already separate)

---

## Testing Strategy

### Before Refactoring
1. Note current test coverage (baseline)
2. Ensure all tests pass

### During Refactoring
1. Maintain test file structure alongside package structure
2. Update import paths as files move
3. Run tests after each file move

### After Refactoring
1. Verify all tests still pass
2. Check code coverage hasn't decreased
3. Document any coverage changes

---

## Rollback Plan

If issues arise:
1. Save current state: `cp -r cmd/app cmd/app.backup`
2. Identify issue
3. Fix in new structure (preferred) or restore from backup
4. Retest

---

## Benefits of Reorganization

### Maintainability
- Clear separation of concerns
- Easier to find related code
- Reduced cognitive load per file

### Testability
- Easier to isolate and test components
- Better mock opportunities
- More focused test files

### Extensibility
- Adding new features easier with clear boundaries
- Less need to touch many files for new feature
- Better code reuse across features

### Consistency
- Matches Mobile app structure (repository pattern)
- Aligns with Server refactoring
- Easier for new developers

---

## Timeline

**Week 1:**
- Mon-Tue: Phases 1-2 (Config, Logger, Storage)
- Wed-Thu: Phase 3 (Sync layer)
- Fri: Phase 4 (UI layer)

**Week 2:**
- Mon-Tue: Phase 5 (Handlers)
- Wed: Phase 6 (App manager)
- Thu: Phase 7 (Tests)
- Fri: Phase 8 (Documentation, cleanup)

---

## Success Criteria

✅ All 19 files moved/refactored to appropriate packages
✅ All imports updated and working
✅ All tests pass
✅ Build succeeds with no warnings
✅ Code coverage maintained or improved
✅ Documentation updated
✅ Package boundaries clear and enforced
✅ No circular dependencies

---

## Notes

- This is the largest refactoring in Phase 2
- Keep frequent commits (one per file move)
- Test frequently during migration
- Update imports carefully to avoid breakage

See `CODE_AUDIT_SUMMARY.md` Section 4 for original analysis.
