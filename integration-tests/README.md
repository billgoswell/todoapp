# Integration & E2E Testing Framework

Comprehensive integration and end-to-end testing suite for the todo-app monorepo.

## Quick Start

### Run All Tests

```bash
# Go tests (CLI ↔ Server)
cd integration-tests
go test ./... -v

# TypeScript tests (Mobile, Workflows, Performance)
npm test
```

### Run by Category

```bash
# CLI ↔ Server integration
go test ./cli-server -v

# Mobile ↔ Server integration
jest mobile-server/

# Multi-device sync
jest multi-device/

# User workflows
jest workflows/

# Error recovery
jest error-recovery/

# Performance tests
jest performance/
```

## Test Coverage

- **72 Total Tests**
- **6 Categories**
- **3 Platforms** (CLI, Server, Mobile)

### Phase 5a: CLI ↔ Server (15 tests)
- Connection & health checks (3)
- Authentication flow (3)
- CRUD operations (9)

### Phase 5b: Mobile ↔ Server (15 tests)
- App initialization (5)
- Offline operations (5)
- Sync scenarios (5)

### Phase 5c: Multi-Device Sync (12 tests)
- Two-device scenarios (6)
- Three+ device scenarios (4)
- Deletion & archival (2)

### Phase 5d: User Workflows (14 tests)
- Create & complete tasks (3)
- Manage lists (3)
- Offline then online (2)
- Priority & due dates (3)
- Export/import (1)

### Phase 5e: Error Recovery (8 tests)
- Network failures (3)
- Data conflicts (2)
- Edge cases (3)

### Phase 5f: Performance & Load (8 tests)
- Data volume (3)
- Concurrency (2)
- Memory & cleanup (2)
- Long-term usage (1)

## Test Structure

```
integration-tests/
├── README.md                    # This file
├── go.mod                       # Go module definition
│
├── utils/
│   └── test_helpers.go         # Shared test utilities & helpers
│
├── cli-server/                  # Phase 5a: CLI ↔ Server tests
│   ├── connection_test.go
│   ├── authentication_test.go
│   └── crud_operations_test.go
│
├── mobile-server/               # Phase 5b: Mobile ↔ Server tests
│   ├── initialization.test.ts
│   ├── offline-operations.test.ts
│   └── sync-scenarios.test.ts
│
├── multi-device/                # Phase 5c: Multi-device sync tests
│   ├── two-device-sync.test.ts
│   ├── three-plus-device.test.ts
│   └── deletion-scenarios.test.ts
│
├── workflows/                   # Phase 5d: User workflow tests
│   └── user-workflows.test.ts
│
├── error-recovery/              # Phase 5e: Error recovery tests
│   └── error-recovery.test.ts
│
└── performance/                 # Phase 5f: Performance tests
    └── performance.test.ts
```

## Key Features

### Sync Validation
- ✅ Last-write-wins conflict resolution with timestamps
- ✅ Client ID duplicate prevention
- ✅ Change log tracking and metadata
- ✅ Soft delete handling across devices

### Multi-Device Coordination
- ✅ Two-device synchronization scenarios
- ✅ Three+ device coordination
- ✅ Network delay simulation
- ✅ State consistency verification

### Error Handling
- ✅ Graceful network failure handling
- ✅ Exponential backoff with jitter
- ✅ Input validation and rejection
- ✅ Cascade deletion

### Performance Validation
- ✅ Large dataset handling (1000+ items)
- ✅ Concurrent operation safety
- ✅ Memory stability monitoring
- ✅ Throughput under load

## Configuration

### Go Tests

Server URL and API key can be configured via environment variables:

```bash
export SERVER_URL=http://localhost:8080
export API_KEY=your-api-key
export CLI_PATH=../todo-cmdline/todo

go test ./cli-server -v
```

Defaults:
- `SERVER_URL`: http://localhost:8080
- `API_KEY`: test-api-key-12345
- `CLI_PATH`: ../todo-cmdline/todo

### TypeScript Tests

Jest configuration is inherited from the mobile app. Run from integration-tests directory:

```bash
npm test
```

## Dependencies

### Go
- github.com/stretchr/testify (assertions and testing utilities)
- Standard library (net/http, testing, time, etc.)

### TypeScript
- Jest (testing framework)
- Standard assertions

## Extending the Tests

### Add a Go Test

1. Choose appropriate category (cli-server, error-recovery, etc.)
2. Add test function to relevant _test.go file
3. Use `utils.HTTPClient` for server communication
4. Use `testify/assert` and `testify/require` for assertions

Example:
```go
func TestNewScenario(t *testing.T) {
    config := utils.DefaultConfig()
    client := utils.NewHTTPClient(config)

    resp, err := client.Get("/api/v1/tasks")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### Add a TypeScript Test

1. Choose appropriate category
2. Add test function to relevant .test.ts file
3. Use Jest `expect()` and Jest testing utilities
4. Simulate state/devices as needed

Example:
```typescript
test('description', () => {
    const state = { /* initial */ };
    // operations
    expect(state.property).toBe(expectedValue);
});
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run integration tests
  run: |
    cd integration-tests
    go test ./cli-server -v
    npm test
```

### Local Pre-commit

```bash
#!/bin/bash
cd integration-tests
go test ./cli-server -v || exit 1
npm test || exit 1
```

## Troubleshooting

### Go tests fail: "connection refused"
- **Cause**: Server not running at localhost:8080
- **Solution**: Start server or use mock-based tests

### TypeScript tests fail: "Cannot find module"
- **Cause**: Dependencies not installed
- **Solution**: Run `npm install` in parent mobile app directory

### Tests hang
- **Cause**: Timeout values too large
- **Solution**: Reduce timeout in test or config

## Performance Baselines

Expected performance (baseline):
- CLI ↔ Server: < 1s for single operation
- Sync cycle: < 100ms for small datasets (< 100 items)
- Large dataset (1000+ items): < 5s
- Memory: Stable within ±10MB during operation

## Future Enhancements

- [ ] Docker-based isolated test environment
- [ ] Real WebSocket testing for live sync
- [ ] Mobile app E2E tests with Detox
- [ ] Database-level verification after sync
- [ ] Chaos engineering tests
- [ ] Performance benchmarking dashboard

## Related Documentation

- See `PHASE_5_IMPLEMENTATION.md` for detailed implementation notes
- See `PHASE_5_PLAN.md` for original planning document
- See phase-specific docs: PHASE_2_IMPLEMENTATION.md, PHASE_3_IMPLEMENTATION.md, PHASE_4_IMPLEMENTATION.md

## Support

For issues or questions about integration tests:
1. Check test documentation in test files
2. Review README.md (this file)
3. Check PHASE_5_IMPLEMENTATION.md for detailed implementation details
