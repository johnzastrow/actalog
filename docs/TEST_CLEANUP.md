# Test Cleanup Plan

> Generated: 2026-01-09
> Status: In Progress

This document tracks the cleanup of test anti-patterns identified in the codebase. The goal is to ensure all tests actually verify business logic rather than Go language features.

---

## Summary

| Category | Tests to Fix | Priority |
|----------|-------------|----------|
| Struct field assignment tests | ~40 | HIGH |
| Panic-expectation tests | ~70 | HIGH |
| No-assertion tests | ~10 | MEDIUM |
| Language feature tests | ~10 | LOW |
| Missing error path tests | ~20 | MEDIUM |

**Estimated effort:** 4-6 hours of focused work

---

## Anti-Pattern 1: Struct Field Assignment Tests (DELETE)

These tests verify that Go's struct assignment works. They provide zero value.

### Files to Clean

- [ ] `internal/handler/admin_handler_test.go`
  - Lines 63-116: `TestWODMismatch_Struct`
  - Lines 118-145: `TestMovementMismatch_Struct`
  - Lines 147-218: Other struct tests

- [ ] `internal/handler/pr_handler_test.go`
  - Lines 26-82: `TestPersonalRecord_Struct`
  - Lines 84-108: `TestPRResponse_Struct`
  - Lines 110-177: Other struct tests

- [ ] `internal/handler/subscription_handler_test.go`
  - Lines 26-46: `TestCreateUserSubscriptionRequest_Struct`
  - Lines 48-59: `TestCreateUserSubscriptionRequest_PermanentFree`
  - Lines 61-81: `TestCreateOrganizationSubscriptionRequest_Struct`
  - Lines 83-97: `TestCancelSubscriptionRequest_Struct`
  - Lines 99-114: `TestMarkUserSubscriptionAsPaidRequest_Struct`
  - Lines 116-125: `TestMarkUserSubscriptionAsPaidRequest_NilFields`
  - Lines 127-143: `TestSetUserSubscriptionPermanentRequest_Struct`

- [ ] `internal/handler/user_handler_test.go`
  - Lines 26-48: `TestUpdateProfileRequest_Struct`
  - Lines 50-68: `TestChangePasswordRequest_Struct`

- [ ] `pkg/email/email_test.go`
  - Lines 66-102: `TestNewService` (field assignment checks)
  - Lines 104-124: `TestConfig_Struct`
  - Lines 126-149: `TestMessage`
  - Lines 151-166: `TestMessage_MinimalFields`
  - Lines 168-195: `TestConfig_Defaults`
  - Lines 197-211: `TestEmailService_Interface`
  - Lines 602-626: `TestConfig_Ports`
  - Lines 629-647: `TestSMTPDebugInfo_EmptyFields`
  - Lines 649-663: `TestSMTPDebugInfo_WithValues`
  - Lines 666-685: `TestSMTPDebugInfo_PartialValues`

### Action
```bash
# These tests should be deleted entirely
# They test Go's assignment operator, not business logic
```

---

## Anti-Pattern 2: Panic-Expectation Tests (REWRITE)

These tests create handlers with nil dependencies and expect panics. They should be rewritten to use mocks and test actual behavior.

### Files to Clean

- [ ] `internal/handler/admin_handler_test.go`
  - Lines 333-348: `TestAdminHandler_ListUserCreatedWODs_NilService`
  - Lines 350-365: `TestAdminHandler_ListUserCreatedMovements_NilService`
  - Lines 367-382: `TestAdminHandler_GetUserStats_NilService`
  - (Continue through line ~620 - approximately 15 tests)

- [ ] `internal/handler/auth_handler_test.go`
  - Lines 400-414: `TestAuthHandler_Register_ValidInput` (misleading name)
  - Lines 416-430: `TestAuthHandler_Login_ValidInput`
  - (Continue through line ~739 - approximately 20 tests)

- [ ] `internal/handler/movement_handler_test.go`
  - Lines 148-162: `TestMovementHandler_ListAll_NilRepo`
  - Lines 164-178: `TestMovementHandler_GetByID_NilRepo`
  - (Continue through line ~451 - approximately 10 tests)

- [ ] `internal/handler/pr_handler_test.go`
  - Lines 191-205: `TestPRHandler_GetPersonalRecords_NilService`
  - (Continue through line ~500 - approximately 10 tests)

- [ ] `internal/handler/wod_handler_test.go`
  - Lines 84-99: `TestWODHandler_ListAll_NilRepo`
  - (Continue through line ~721 - approximately 15 tests)

- [ ] `internal/handler/user_workout_handler_test.go`
  - Multiple `_NilService` tests

- [ ] `internal/handler/workout_template_handler_test.go`
  - Multiple `_NilService` tests

### Action

Replace panic-expectation pattern:
```go
// BEFORE (bad)
func TestHandler_Method_NilService(t *testing.T) {
    handler := &Handler{logger: createTestLogger()}
    defer func() {
        if r := recover(); r == nil {
            t.Error("Expected panic")
        }
    }()
    handler.Method(rr, req)
}

// AFTER (good)
func TestHandler_Method_Success(t *testing.T) {
    svc := createTestService()
    handler := &Handler{service: svc, logger: createTestLogger()}

    req := createAuthenticatedRequest(...)
    rr := httptest.NewRecorder()

    handler.Method(rr, req)

    assertStatusCode(t, rr, http.StatusOK)
    assertBodyContains(t, rr, "expected_content")
}

func TestHandler_Method_ServiceError(t *testing.T) {
    svc, mocks := createTestServiceWithMocks()
    mocks.repo.SetError(errors.New("db error"))

    handler := &Handler{service: svc, logger: createTestLogger()}
    handler.Method(rr, req)

    assertStatusCode(t, rr, http.StatusInternalServerError)
}
```

---

## Anti-Pattern 3: No-Assertion Tests (FIX OR DELETE)

Tests that don't actually assert anything meaningful.

### Files to Clean

- [ ] `internal/handler/movement_handler_test.go`
  - Lines 148-162: Only logs message, no assertion

- [ ] `internal/service/subscription_service_test.go`
  - Lines 1236-1263: `TestSubscriptionService_ExpireOverdueSubscriptions`
    - Tests unimplemented function returns 0
    - Either implement the function or delete the test

### Action
Add meaningful assertions or delete the test.

---

## Anti-Pattern 4: Language Feature Tests (DELETE)

Tests that verify Go language features work correctly.

### Files to Clean

- [ ] `pkg/logger/logger_test.go`
  - Lines 311-326: `TestLevelNames` - tests map lookups work

- [ ] `pkg/version/version_test.go`
  - Lines 126-140: `TestVersionConstants` - tests constants >= 0

- [ ] `pkg/email/email_test.go`
  - Lines 168-195: `TestConfig_Defaults` - tests zero values are zero
  - Lines 197-211: `TestEmailService_Interface` - tests interface assignment

### Action
Delete these tests. The Go compiler already verifies these things.

---

## Anti-Pattern 5: Missing Error Path Tests (ADD)

Services that lack tests for repository failures.

### Files to Enhance

- [ ] `internal/service/subscription_service_test.go`
  - Add repo error tests for all CRUD operations

- [ ] `internal/service/movement_service_test.go`
  - Add `repo.createError` test cases

- [ ] `internal/service/workout_service_test.go`
  - Add database failure scenarios

### Action
Add tests like:
```go
func TestService_Create_RepoError(t *testing.T) {
    repo := newMockRepo()
    repo.createError = errors.New("database connection failed")

    svc := NewService(repo)
    err := svc.Create(...)

    if err == nil {
        t.Error("expected error when repo fails")
    }
}
```

---

## Anti-Pattern 6: Constructor Tests (SIMPLIFY OR DELETE)

Tests that only verify constructor returns non-nil.

### Files to Clean

- [ ] `internal/service/workout_service_test.go`
  - Lines 9-20: `TestNewWorkoutService` - just checks `!= nil`

- [ ] `internal/handler/email_handler_test.go`
  - Lines 12-19: `TestNewEmailHandler` - just checks `!= nil`

- [ ] `pkg/email/email_test.go`
  - Lines 66-102: `TestNewService` - extensive field checks

### Action
Either delete entirely or reduce to single line:
```go
func TestNewService(t *testing.T) {
    if NewService(deps) == nil {
        t.Fatal("constructor returned nil")
    }
}
```

---

## Anti-Pattern 7: Trivial Helper Tests (DELETE)

Tests for one-line helper functions with no logic.

### Files to Clean

- [ ] `pkg/middleware/rate_limit_test.go`
  - Lines 234-251: `TestFormatInt` - tests `strconv.Itoa` wrapper

- [ ] `pkg/middleware/auth_test.go`
  - Lines 412-420: `TestContains` - tests string contains helper

### Action
Delete these tests. The helpers are too trivial to warrant testing.

---

## Cleanup Checklist

### Phase 1: Delete Low-Value Tests (1-2 hours)
- [ ] Delete all struct field assignment tests
- [ ] Delete all language feature tests
- [ ] Delete trivial helper tests
- [ ] Simplify constructor tests

### Phase 2: Rewrite Panic Tests (2-3 hours)
- [ ] Create mock service helpers for each handler
- [ ] Convert `_NilService` tests to success path tests
- [ ] Add error path tests using mock errors

### Phase 3: Add Missing Tests (1 hour)
- [ ] Add repository error handling tests
- [ ] Strengthen overly broad assertions

### Phase 4: Verify Coverage
- [ ] Run `go test -cover ./...`
- [ ] Ensure coverage remains >= 70%
- [ ] Verify all tests pass

---

## Progress Tracking

| File | Status | Notes |
|------|--------|-------|
| `internal/handler/admin_handler_test.go` | TODO | ~20 tests to fix |
| `internal/handler/auth_handler_test.go` | TODO | ~25 tests to fix |
| `internal/handler/subscription_handler_test.go` | PARTIAL | Recent improvements made |
| `internal/handler/movement_handler_test.go` | TODO | ~15 tests to fix |
| `internal/handler/pr_handler_test.go` | TODO | ~15 tests to fix |
| `internal/handler/wod_handler_test.go` | TODO | ~20 tests to fix |
| `internal/handler/user_handler_test.go` | TODO | ~5 tests to fix |
| `internal/service/*_test.go` | TODO | Minor fixes needed |
| `internal/repository/*_test.go` | OK | Generally sound |
| `pkg/email/email_test.go` | TODO | ~15 tests to delete |
| `pkg/logger/logger_test.go` | TODO | ~2 tests to delete |
| `pkg/version/version_test.go` | TODO | ~1 test to delete |
| `pkg/middleware/*_test.go` | TODO | ~3 tests to delete |

---

## References

- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testing Techniques](https://go.dev/blog/subtests)
