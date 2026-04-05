# Testing Documentation

**Last Updated:** 2026-04-03
**Version:** 1.2.2

## Overview

This document tracks testing progress and coverage for ActaLog, providing guidelines for writing tests and maintaining the test suite.

## Test Coverage Summary

### Overall Coverage: 81.6%

| Service | Coverage | Status |
|---------|----------|--------|
| audit_log_service.go | 100.0% | ✅ Complete |
| backup_service.go | 100.0% | ✅ Complete |
| data_change_log_service.go | 100.0% | ✅ Complete |
| email_log_service.go | 100.0% | ✅ Complete |
| movement_service.go | 100.0% | ✅ Complete |
| notification_like_service.go | 100.0% | ✅ Complete |
| subscription_service.go | 100.0% | ✅ Complete |
| user_settings_service.go | 95.2% | ✅ Complete |
| user_workout_service.go | 100.0% | ✅ Complete |
| wod_service.go | 100.0% | ✅ Complete |
| wodify_import_service.go | 100.0% | ✅ Complete |
| workout_template_service.go | 100.0% | ✅ Complete |
| workout_wod_service.go | 100.0% | ✅ Complete |
| export_service.go | 83.6% | 🔄 Good |
| notification_service.go | 82.6% | 🔄 Good |
| workout_service.go | 80.0% | 🔄 Good |
| organization_service.go | 77.8% | 🔄 Needs Improvement |
| user_service.go | 61.9% | ⚠️ Needs Improvement |
| import_service.go | 60.8% | ⚠️ Needs Improvement |

### Coverage Goals

| Component | Target | Current | Status |
|-----------|--------|---------|--------|
| Overall | >80% | 81.6% | ✅ Met |
| Service Layer | >90% | 81.6% | 🔄 Close |
| Repository Layer | >85% | - | ⏳ Pending |
| Handler Layer | >75% | - | ⏳ Pending |

## Test Infrastructure

### Directory Structure

```
internal/service/
├── test_helpers.go            # Shared mock repositories (600+ lines)
├── *_service_test.go          # Service unit tests
```

### Mock Repository Infrastructure

Located in `internal/service/test_helpers.go`:

**User-Related Mocks:**
- `mockUserRepo` - User CRUD with error injection support
- `mockUserSettingsRepo` - User settings management
- `mockRefreshTokenRepo` - JWT refresh token handling

**Workout Mocks:**
- `mockUserWorkoutRepo` - User workout logging with stats
- `mockUserWorkoutMovementRepo` - Movement performance tracking
- `mockUserWorkoutWODRepo` - WOD performance tracking
- `mockWorkoutRepo` - Workout template management
- `mockWorkoutMovementRepo` - Template-movement associations
- `mockWorkoutWODRepo` - Template-WOD associations

**Entity Mocks:**
- `mockMovementRepo` - Movement CRUD with category support
- `mockWODRepo` - WOD management with search
- `mockWorkoutTemplateRepo` - Template CRUD

**System Mocks:**
- `mockAuditLogRepo` - Audit trail logging
- `mockNotificationRepo` - Notification system
- `mockSubscriptionAccessRepo` - Subscription access checks
- `mockEmailService` - Email sending simulation

### Mock Pattern

All mocks use constructor functions for proper initialization:

```go
type mockUserWorkoutRepo struct {
    userWorkouts          map[int64]*domain.UserWorkout
    nextID                int64
    getByIDError          error
    createError           error
}

func newMockUserWorkoutRepo() *mockUserWorkoutRepo {
    return &mockUserWorkoutRepo{
        userWorkouts: make(map[int64]*domain.UserWorkout),
        nextID:       1,
    }
}
```

### Error Injection

Mocks support error injection for testing error paths:

```go
mockRepo := newMockUserWorkoutRepo()
mockRepo.createError = errors.New("database connection lost")
// Test will now receive this error from Create()
```

## Running Tests

### Quick Commands

```bash
# Run all tests
make test

# Run service tests only
go test -v ./internal/service/...

# Run with coverage
go test -coverprofile=coverage.out ./internal/service/...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestBackupService ./internal/service/...

# Run with race detection
go test -race ./internal/service/...
```

### Coverage Report

```bash
# Generate per-function coverage
go test -coverprofile=coverage.out ./internal/service/...
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

## Test Patterns

### Table-Driven Tests

All tests follow the table-driven pattern for comprehensive coverage:

```go
func TestService_Method(t *testing.T) {
    tests := []struct {
        name          string
        input         InputType
        setupMock     func(*mockRepo)
        expectedError error
        validate      func(*testing.T, ResultType)
    }{
        {
            name:  "successful operation",
            input: validInput,
            setupMock: func(m *mockRepo) {
                m.data[1] = expectedData
            },
            expectedError: nil,
            validate: func(t *testing.T, result ResultType) {
                if result.ID != expected.ID {
                    t.Errorf("expected ID %d, got %d", expected.ID, result.ID)
                }
            },
        },
        {
            name:          "handles not found",
            input:         missingInput,
            expectedError: service.ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := newMockRepo()
            if tt.setupMock != nil {
                tt.setupMock(mockRepo)
            }
            svc := NewService(mockRepo)

            result, err := svc.Method(tt.input)

            if tt.expectedError != nil {
                if !errors.Is(err, tt.expectedError) {
                    t.Errorf("expected error %v, got %v", tt.expectedError, err)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if tt.validate != nil {
                tt.validate(t, result)
            }
        })
    }
}
```

### Integration Tests with SQLite

For tests requiring a real database (backup/restore, migrations):

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open test database: %v", err)
    }

    // Create schema
    _, err = db.Exec(`
        CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT UNIQUE);
        CREATE TABLE movements (id INTEGER PRIMARY KEY, name TEXT);
        -- ... full schema
    `)
    if err != nil {
        t.Fatalf("failed to create schema: %v", err)
    }

    return db
}
```

### Arrange-Act-Assert Pattern

```go
func TestExample(t *testing.T) {
    // Arrange - Set up test data and mocks
    mockRepo := newMockRepo()
    mockRepo.data[1] = &domain.Entity{ID: 1, Name: "test"}
    service := NewService(mockRepo)

    // Act - Execute the operation
    result, err := service.GetByID(1)

    // Assert - Verify the outcome
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Name != "test" {
        t.Errorf("expected name 'test', got %q", result.Name)
    }
}
```

## Test Coverage by Feature

### Backup Service (100% coverage)

Comprehensive tests in `backup_service_test.go`:

- `TestBackupService_CreateBackup` - Full backup creation with ZIP archive
- `TestBackupService_CreateBackup_UserNotFound` - Error handling for missing user
- `TestBackupService_CreateBackup_WithUploads` - Backup with file attachments
- `TestBackupService_RestoreBackup` - Replace mode restoration
- `TestBackupService_RestoreBackup_MergeMode` - Merge mode with updates
- `TestBackupService_RestoreBackup_SkipMode` - Skip mode preserving existing
- `TestBackupService_RestoreBackup_FileNotFound` - Missing backup file handling
- `TestBackupService_CreateSQLiteDump` - SQLite dump generation
- `TestBackupService_ExportAllTables` - JSON export of all tables
- `TestBackupService_ExportAllTables_EmptyDatabase` - Empty database handling

### User Workout Service (100% coverage)

Tests in `user_workout_service_test.go`:

- `TestUserWorkoutService_LogWorkout` - Workout logging
- `TestUserWorkoutService_LogWorkoutWithPerformance` - Logging with PR detection
- `TestUserWorkoutService_GetLoggedWorkout` - Retrieve with authorization
- `TestUserWorkoutService_UpdateLoggedWorkout` - Update with authorization
- `TestUserWorkoutService_DeleteLoggedWorkout` - Delete with authorization
- `TestUserWorkoutService_GetWorkoutStatsForMonth` - Monthly statistics
- `TestUserWorkoutService_ListByUser` - User workout listing
- `TestUserWorkoutService_GetWorkoutSummary` - Summary calculations

### Subscription Service (100% coverage)

Tests in `subscription_service_test.go`:

- Create/cancel user subscriptions
- Create/cancel organization subscriptions
- Mark subscriptions as paid
- Access checking (user-level and organization-level)
- Permanent free subscription handling

### Import/Export Services

Tests in `import_service_test.go`, `wodify_import_service_test.go`:

- CSV parsing and validation
- Duplicate detection and handling
- Skip vs update modes
- Error recovery
- Multi-format support (Wodify CSV, JSON)

## Writing New Tests

### Guidelines

1. **Test Behavior, Not Implementation**
   - Focus on what the service does, not how
   - Test public API only

2. **Use Descriptive Names**
   - Good: `"successful workout log with PR detection"`
   - Bad: `"test1"`

3. **Cover Edge Cases**
   - Happy path (success scenarios)
   - Error conditions (not found, unauthorized, validation)
   - Boundary conditions (empty lists, nil values, limits)

4. **Keep Tests Independent**
   - Each test runs in isolation
   - No shared state between tests
   - Fresh mock instances per test

5. **Use `errors.Is()` for Error Comparison**
   ```go
   if !errors.Is(err, service.ErrNotFound) {
       t.Errorf("expected ErrNotFound, got %v", err)
   }
   ```

### Adding a New Mock

1. Add interface implementation to `test_helpers.go`
2. Include constructor function with map initialization
3. Add error injection fields as needed
4. Implement all interface methods

## Continuous Integration

Tests run automatically on:
- Pull request creation
- Push to main branch

CI configuration: `.github/workflows/ci.yml`

```yaml
- name: Run tests
  run: go test -v -race -coverprofile=coverage.out ./...
```

## Next Steps

### High Priority
1. Improve `user_service.go` coverage (61.9% → 80%+)
2. Improve `import_service.go` coverage (60.8% → 80%+)
3. Add handler unit tests

### Medium Priority
1. Add repository integration tests
2. Improve `organization_service.go` coverage (77.8% → 90%+)
3. Add E2E API tests

### Future Work
1. Frontend component tests (Vue Test Utils)
2. E2E tests (Playwright/Cypress)
3. Performance/load testing

---

*This file is maintained alongside test improvements. Update coverage stats after significant test additions.*
