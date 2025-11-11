# Testing Documentation

**Last Updated:** 2025-11-10
**Version:** 0.4.0-dev

## Overview

This document tracks testing progress for the ActaLog v0.4.0 template-based architecture and provides guidelines for writing tests.

## Test Coverage Status

### Unit Tests

#### ✅ Completed

**Test Infrastructure:**
- `internal/service/test_helpers.go` (334 lines)
  - Mock `UserWorkoutRepository` - Full interface implementation
  - Mock `WorkoutRepository` - Full interface implementation
  - Mock `WorkoutMovementRepository` - Full interface implementation
  - Helper functions: `stringPtr()`, `intPtr()`, `int64Ptr()`

**UserWorkoutService Tests:**
- `internal/service/user_workout_service_test.go` (483 lines)
  - **Test Results: 11/16 passing (68%)**

| Test Function | Test Cases | Passing | Status |
|--------------|-----------|---------|---------|
| `TestUserWorkoutService_LogWorkout` | 4 | 4 | ✅ All passing |
| `TestUserWorkoutService_GetLoggedWorkout` | 3 | 1 | ⚠️ Error wrapping issue |
| `TestUserWorkoutService_UpdateLoggedWorkout` | 3 | 2 | ⚠️ Error wrapping issue |
| `TestUserWorkoutService_DeleteLoggedWorkout` | 3 | 2 | ⚠️ Error wrapping issue |
| `TestUserWorkoutService_GetWorkoutStatsForMonth` | 2 | 2 | ✅ All passing |

**Coverage:**
- ✅ Authorization checks (user ownership verification)
- ✅ Business logic validation
- ✅ Repository interaction
- ✅ Success paths
- ⚠️ Error handling paths (needs error wrapping fix)
- ✅ Edge cases (not found, unauthorized, date ranges)

#### 🔄 In Progress

**WODService Tests** (`internal/service/wod_service_test.go` - not yet created):
- [ ] `TestWODService_CreateWOD` - Create custom WOD
- [ ] `TestWODService_GetWOD` - Retrieve WOD by ID
- [ ] `TestWODService_GetWODByName` - Retrieve WOD by name
- [ ] `TestWODService_ListStandardWODs` - List standard CrossFit WODs
- [ ] `TestWODService_ListUserWODs` - List user's custom WODs
- [ ] `TestWODService_ListAllWODs` - List all WODs (standard + custom)
- [ ] `TestWODService_UpdateWOD` - Update custom WOD
- [ ] `TestWODService_DeleteWOD` - Delete custom WOD
- [ ] `TestWODService_SearchWODs` - Search WODs by name

**WorkoutWODService Tests** (`internal/service/workout_wod_service_test.go` - not yet created):
- [ ] `TestWorkoutWODService_AddWODToWorkout` - Link WOD to workout template
- [ ] `TestWorkoutWODService_RemoveWODFromWorkout` - Unlink WOD from workout
- [ ] `TestWorkoutWODService_UpdateWorkoutWOD` - Update WOD score/division
- [ ] `TestWorkoutWODService_ToggleWODPR` - Toggle PR flag
- [ ] `TestWorkoutWODService_ListWODsForWorkout` - List WODs in template

**WorkoutService Template Tests** (`internal/service/workout_service_test.go` - needs rewrite):
- [ ] `TestWorkoutService_CreateTemplate` - Create workout template
- [ ] `TestWorkoutService_GetTemplate` - Retrieve template by ID
- [ ] `TestWorkoutService_ListTemplates` - List templates (standard + user's)
- [ ] `TestWorkoutService_UpdateTemplate` - Update template
- [ ] `TestWorkoutService_DeleteTemplate` - Delete template
- [ ] `TestWorkoutService_GetTemplateUsageStats` - Usage statistics

#### ⏳ Pending

**Repository Tests:**
- [ ] `UserWorkoutRepository` tests
- [ ] `WODRepository` tests
- [ ] `WorkoutWODRepository` tests
- [ ] `WorkoutRepository` template operation tests

**Integration Tests:**
- [ ] `user_workout_handler` HTTP tests
- [ ] `wod_handler` HTTP tests
- [ ] `workout_wod_handler` HTTP tests
- [ ] End-to-end API workflow tests

**Frontend Tests:**
- [ ] Component tests
- [ ] Store tests (Pinia)
- [ ] E2E tests (Cypress/Playwright)

## Test Architecture

### Directory Structure

```
test/
├── unit/              # Unit tests (fast, isolated)
└── integration/       # Integration tests (database, HTTP)

internal/service/
├── test_helpers.go    # Shared mock repositories
├── *_service_test.go  # Service unit tests
```

### Test Patterns

#### Table-Driven Tests

All tests use table-driven test pattern for multiple scenarios:

```go
func TestUserWorkoutService_LogWorkout(t *testing.T) {
    tests := []struct {
        name          string
        userID        int64
        workoutID     int64
        setupMock     func(*mockWorkoutRepo)
        expectedError error
    }{
        {
            name: "successful workout log",
            userID: 1,
            workoutID: 1,
            setupMock: func(m *mockWorkoutRepo) {
                // Setup mock data
            },
            expectedError: nil,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Mock Repository Pattern

Shared mock repositories implement full domain interfaces:

```go
type mockUserWorkoutRepo struct {
    userWorkouts map[int64]*domain.UserWorkout
    nextID       int64
    getByIDError error
    createError  error
    // ...
}

func newMockUserWorkoutRepo() *mockUserWorkoutRepo {
    return &mockUserWorkoutRepo{
        userWorkouts: make(map[int64]*domain.UserWorkout),
        nextID:       1,
    }
}
```

## Known Issues

### Error Wrapping

**Issue:** Service layer wraps errors with `fmt.Errorf()`, but tests use direct error comparison.

**Current (fails):**
```go
if err != tt.expectedError {
    t.Errorf("expected error %v, got %v", tt.expectedError, err)
}
```

**Should be:**
```go
if tt.expectedError != nil {
    if !errors.Is(err, tt.expectedError) {
        t.Errorf("expected error %v, got %v", tt.expectedError, err)
    }
}
```

**Affected Tests:**
- `TestUserWorkoutService_GetLoggedWorkout` - 2/3 failing
- `TestUserWorkoutService_UpdateLoggedWorkout` - 1/3 failing
- `TestUserWorkoutService_DeleteLoggedWorkout` - 1/3 failing

**Fix:** Update test assertions to use `errors.Is()` for wrapped error comparison.

### Deprecated Test Files

- `internal/service/workout_service_test.go.old` - v0.3.x tests (incompatible with v0.4.0)
- `internal/service/user_service_test.go` - Needs update for new UserService constructor

## Running Tests

### Run All Tests
```bash
make test
```

### Run Specific Service Tests
```bash
go test -v -run TestUserWorkoutService ./internal/service/user_workout_service_test.go ./internal/service/user_workout_service.go ./internal/service/test_helpers.go
```

### Run Unit Tests Only
```bash
make test-unit
```

### Run Integration Tests Only
```bash
make test-integration
```

### Test with Coverage
```bash
go test -coverprofile=coverage.out ./internal/service/...
go tool cover -html=coverage.out
```

## Test Guidelines

### Writing Good Tests

1. **Test Behavior, Not Implementation**
   - Focus on what the service does, not how
   - Test public API, not internal details

2. **Use Descriptive Test Names**
   - Good: `"successful workout log"`
   - Bad: `"test1"`

3. **Cover Edge Cases**
   - Happy path (success scenarios)
   - Error conditions (not found, unauthorized, validation failures)
   - Boundary conditions (empty lists, nil values, date ranges)

4. **Keep Tests Independent**
   - Each test should run in isolation
   - No shared state between tests
   - Use fresh mock instances per test

5. **Mock External Dependencies**
   - Repository interfaces
   - Email services
   - External APIs

### Test Organization

```go
// 1. Arrange - Set up test data
mockRepo := newMockUserWorkoutRepo()
if tt.setupMock != nil {
    tt.setupMock(mockRepo)
}
service := NewUserWorkoutService(mockRepo, ...)

// 2. Act - Execute the operation
result, err := service.LogWorkout(...)

// 3. Assert - Verify the outcome
if tt.expectedError != nil {
    if !errors.Is(err, tt.expectedError) {
        t.Errorf("expected error %v, got %v", tt.expectedError, err)
    }
    return
}

if err != nil {
    t.Errorf("unexpected error: %v", err)
}

// Verify result
if result.UserID != tt.userID {
    t.Errorf("expected user ID %d, got %d", tt.userID, result.UserID)
}
```

## Coverage Goals

### Target Coverage

- **Overall:** >80%
- **Service Layer:** >90%
- **Repository Layer:** >85%
- **Handler Layer:** >75%

### Current Coverage

| Component | Coverage | Status |
|-----------|----------|--------|
| UserWorkoutService | 68% | 🔄 In Progress |
| WODService | 0% | ⏳ Not Started |
| WorkoutWODService | 0% | ⏳ Not Started |
| WorkoutService (templates) | 0% | ⏳ Not Started |
| Repositories | 0% | ⏳ Not Started |
| Handlers | 0% | ⏳ Not Started |

## Next Steps

### Immediate Priorities

1. **Fix Error Wrapping in UserWorkoutService Tests**
   - Update assertions to use `errors.Is()`
   - Target: 16/16 tests passing (100%)

2. **Complete WODService Tests**
   - Create `wod_service_test.go`
   - Add mock WODRepository to `test_helpers.go`
   - Implement 9 test functions

3. **Complete WorkoutWODService Tests**
   - Create `workout_wod_service_test.go`
   - Add mock WorkoutWODRepository to `test_helpers.go`
   - Implement 5 test functions

4. **Rewrite WorkoutService Tests for v0.4.0**
   - Update `workout_service_test.go` for template operations
   - Remove v0.3.x user-specific workout tests
   - Focus on template CRUD operations

### Future Work

1. **Repository Integration Tests**
   - Test with real database (SQLite in-memory)
   - Verify SQL queries and transactions
   - Test migrations

2. **Handler Integration Tests**
   - HTTP request/response testing
   - JWT authentication verification
   - JSON serialization/deserialization

3. **E2E Tests**
   - Complete user workflows
   - Frontend-backend integration
   - Cross-browser compatibility

4. **Performance Tests**
   - Load testing for API endpoints
   - Database query optimization
   - Memory leak detection

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify Package](https://github.com/stretchr/testify) (if needed for assertions)
- [Go Mock](https://github.com/golang/mock) (alternative to manual mocks)

---

**Maintained by:** Development Team
**Review Frequency:** Weekly during v0.4.0 development
