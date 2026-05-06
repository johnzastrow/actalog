package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
)

func TestAdminUserHandler_UnlockUser_Unauthorized(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/users/1/unlock", "")
	rr := httptest.NewRecorder()

	handler.UnlockUser(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestAdminUserHandler_UnlockUser_InvalidID(t *testing.T) {
	handler := &AdminUserHandler{}

	// chi.URLParam returns empty string without router context
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/abc/unlock", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UnlockUser(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestAdminUserHandler_DisableUser_Unauthorized(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/users/1/disable", "")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestAdminUserHandler_DisableUser_InvalidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/abc/disable", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestAdminUserHandler_EnableUser_Unauthorized(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/users/1/enable", "")
	rr := httptest.NewRecorder()

	handler.EnableUser(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestAdminUserHandler_EnableUser_InvalidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/abc/enable", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.EnableUser(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestAdminUserHandler_ChangeUserRole_Unauthorized(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodPut, "/api/admin/users/1/role", `{"role": "admin"}`)
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestAdminUserHandler_ChangeUserRole_InvalidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/abc/role", `{"role": "admin"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestAdminUserHandler_ChangeUserRole_InvalidJSON(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/1/role", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	// Without chi router context, URL param is empty -> Invalid user ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestAdminUserHandler_ChangeUserRole_InvalidRole(t *testing.T) {
	handler := &AdminUserHandler{}

	// Note: This test would need chi router context to get past the ID parsing
	// Testing the role validation directly would require a mock chi context
	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/1/role", `{"role": "superuser"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	// Without chi router context, URL param is empty -> Invalid user ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestAdminUserHandler_ToggleEmailVerification_Unauthorized(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/users/1/toggle-email-verification", `{"verified": true}`)
	rr := httptest.NewRecorder()

	handler.ToggleEmailVerification(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestAdminUserHandler_ToggleEmailVerification_InvalidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/abc/toggle-email-verification", `{"verified": true}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ToggleEmailVerification(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestAdminUserHandler_GetUserDetails_InvalidID(t *testing.T) {
	handler := &AdminUserHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/admin/users/abc", "")
	rr := httptest.NewRecorder()

	handler.GetUserDetails(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestAdminUserHandler_DeleteUser_Unauthorized(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodDelete, "/api/admin/users/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestAdminUserHandler_DeleteUser_InvalidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestNewAdminUserHandler(t *testing.T) {
	handler := NewAdminUserHandler(nil, nil, nil)
	if handler == nil {
		t.Error("NewAdminUserHandler should return a non-nil handler")
	}
}

func TestAdminUserHandler_ChangeUserRole_ValidIDInvalidJSON(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/1/role", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestAdminUserHandler_ChangeUserRole_InvalidRoleValue(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/1/role", `{"role": "superuser"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Role must be 'athlete', 'coach', or 'admin'")
}

func TestAdminUserHandler_ToggleEmailVerification_ValidIDInvalidJSON(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/toggle-email-verification", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.ToggleEmailVerification(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

// Removed 9 panic-expectation tests:
// - TestAdminUserHandler_ListUsers_NilService
// - TestAdminUserHandler_ChangeUserRole_ValidRole (2 subtests)
// - TestAdminUserHandler_ToggleEmailVerification_ValidInput
// - TestAdminUserHandler_UnlockUser_ValidID
// - TestAdminUserHandler_DisableUser_ValidID
// - TestAdminUserHandler_EnableUser_ValidID
// - TestAdminUserHandler_GetUserDetails_ValidID
// - TestAdminUserHandler_DeleteUser_ValidID
// - TestAdminUserHandler_ListUsers_WithQueryParams (7 subtests)
// These tests verified nil pointer panics, not business logic.

// Tests with mock service

func TestAdminUserHandler_ListUsers_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createTestRequest(http.MethodGet, "/api/admin/users", "")
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "users")
	assertBodyContains(t, rr, "total")
}

func TestAdminUserHandler_ListUsers_WithPagination(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	tests := []struct {
		name     string
		query    string
		wantCode int
	}{
		{"with valid limit", "?limit=10", http.StatusOK},
		{"with valid offset", "?offset=5", http.StatusOK},
		{"with both params", "?limit=10&offset=5", http.StatusOK},
		{"with invalid limit", "?limit=abc", http.StatusOK},   // falls back to default
		{"with negative limit", "?limit=-1", http.StatusOK},   // falls back to default
		{"with invalid offset", "?offset=abc", http.StatusOK}, // falls back to default
		{"with negative offset", "?offset=-1", http.StatusOK}, // falls back to default
		{"with zero limit", "?limit=0", http.StatusOK},        // falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/admin/users"+tt.query, "")
			rr := httptest.NewRecorder()

			handler.ListUsers(rr, req)

			assertStatusCode(t, rr, tt.wantCode)
		})
	}
}

func TestAdminUserHandler_UnlockUser_UserNotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/999/unlock", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.UnlockUser(rr, req)

	// Service returns 500 for user not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestAdminUserHandler_UnlockUser_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/unlock", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UnlockUser(rr, req)

	// User exists, should succeed
	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_DisableUser_InvalidJSON(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	// Handler treats invalid JSON as no reason (reason is optional)
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/2/disable", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	// Handler succeeds with empty reason
	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_DisableUser_EmptyReason(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	// Reason is optional, so empty body succeeds
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/2/disable", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_DisableUser_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/2/disable", `{"reason": "Policy violation"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_DisableUser_CannotDisableSelf(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/disable", `{"reason": "Self disable"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "cannot disable your own account")
}

func TestAdminUserHandler_EnableUser_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/enable", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.EnableUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_ChangeUserRole_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/2/role", `{"role": "admin"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_ChangeUserRole_CannotChangeOwnRole(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/1/role", `{"role": "athlete"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "cannot change your own role")
}

func TestAdminUserHandler_ToggleEmailVerification_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/toggle-email-verification", `{"verified": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.ToggleEmailVerification(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_GetUserDetails_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createTestRequest(http.MethodGet, "/api/admin/users/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetUserDetails(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_GetUserDetails_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	req := createTestRequest(http.MethodGet, "/api/admin/users/999", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetUserDetails(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
}

func TestAdminUserHandler_DeleteUser_CannotDeleteSelf(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	// User 1 trying to delete themselves
	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "cannot delete your own account")
}

func TestAdminUserHandler_DeleteUser_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, nil, createTestLogger())

	// Admin (user 1) deleting user 2
	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/2", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

// knownUpdatedAt is a fixed non-zero timestamp used by the AdminUserService test helpers.
var knownUpdatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// mockUserRepoNilOnMiss wraps MockUserRepository and returns (nil, nil) from
// GetByID when the user is not found — matching the real repository contract.
// The standard MockUserRepository incorrectly returns (nil, ErrMockNotFound).
type mockUserRepoNilOnMiss struct {
	*MockUserRepository
}

func (m *mockUserRepoNilOnMiss) GetByID(id int64) (*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil // real-repo contract: missing row → (nil, nil)
}

// createTestAdminUserService creates an AdminUserService backed by mocks.
// User 2 ("test@example.com") has UpdatedAt = knownUpdatedAt so that
// optimistic-concurrency tests can supply a matching timestamp.
func createTestAdminUserService() *service.AdminUserService {
	mockUserRepo := &mockUserRepoNilOnMiss{NewMockUserRepository()}
	// Set a known non-zero UpdatedAt on user 2 so the optimistic-concurrency
	// check in UpdateProfile can be exercised.
	for _, u := range mockUserRepo.users {
		if u.ID == 2 {
			u.UpdatedAt = knownUpdatedAt
		}
	}
	mockRefreshTokenRepo := NewMockRefreshTokenRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	mockEmailService := NewMockEmailService()
	auditLogService := service.NewAuditLogService(mockAuditLogRepo)
	return service.NewAdminUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockEmailService,
		auditLogService,
		createTestLogger(),
		"http://localhost:3000",
	)
}

// createTestAdminUserServiceWithProtectedUser returns an AdminUserService whose
// user store includes a protected user (br8kwall@gmail.com) with ID 99, so that
// handler tests can exercise the ErrProtectedUser → 403 path.
func createTestAdminUserServiceWithProtectedUser() *service.AdminUserService {
	inner := NewMockUserRepository()
	inner.users = append(inner.users, &domain.User{
		ID:    99,
		Email: "br8kwall@gmail.com",
		Name:  "Protected User",
		Role:  "admin",
	})
	mockUserRepo := &mockUserRepoNilOnMiss{inner}
	mockRefreshTokenRepo := NewMockRefreshTokenRepository()
	mockAuditLogRepo := NewMockAuditLogRepository()
	mockEmailService := NewMockEmailService()
	auditLogService := service.NewAuditLogService(mockAuditLogRepo)
	return service.NewAdminUserService(
		mockUserRepo,
		mockRefreshTokenRepo,
		mockEmailService,
		auditLogService,
		createTestLogger(),
		"http://localhost:3000",
	)
}

// TestAdminUserHandler_UpdateProfile_HappyPath — PATCH with correct updated_at → 200.
func TestAdminUserHandler_UpdateProfile_HappyPath(t *testing.T) {
	adminUserService := createTestAdminUserService()
	handler := NewAdminUserHandler(nil, adminUserService, createTestLogger())

	body := fmt.Sprintf(`{"name":"New Name","updated_at":"%s"}`, knownUpdatedAt.Format(time.RFC3339Nano))
	req := createAuthenticatedRequest(http.MethodPatch, "/api/admin/users/2", body, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "New Name")
}

// TestAdminUserHandler_UpdateProfile_StaleUpdatedAtReturns409 — PATCH with old updated_at → 409.
func TestAdminUserHandler_UpdateProfile_StaleUpdatedAtReturns409(t *testing.T) {
	adminUserService := createTestAdminUserService()
	handler := NewAdminUserHandler(nil, adminUserService, createTestLogger())

	// Provide a timestamp that does NOT match the stored UpdatedAt.
	stale := knownUpdatedAt.Add(-time.Hour)
	body := fmt.Sprintf(`{"name":"Stale Name","updated_at":"%s"}`, stale.Format(time.RFC3339Nano))
	req := createAuthenticatedRequest(http.MethodPatch, "/api/admin/users/2", body, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusConflict)
}

// TestAdminUserHandler_UpdateProfile_MissingUpdatedAtReturns400 — PATCH without updated_at → 400.
func TestAdminUserHandler_UpdateProfile_MissingUpdatedAtReturns400(t *testing.T) {
	adminUserService := createTestAdminUserService()
	handler := NewAdminUserHandler(nil, adminUserService, createTestLogger())

	body := `{"name":"No Timestamp"}`
	req := createAuthenticatedRequest(http.MethodPatch, "/api/admin/users/2", body, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "updated_at must be a valid non-zero timestamp")
}

// TestAdminUserHandler_ForcePasswordReset_Returns204 — POST force-password-reset → 204.
func TestAdminUserHandler_ForcePasswordReset_Returns204(t *testing.T) {
	adminUserService := createTestAdminUserService()
	handler := NewAdminUserHandler(nil, adminUserService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/2/force-password-reset", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.ForcePasswordReset(rr, req)

	assertStatusCode(t, rr, http.StatusNoContent)
}

// TestAdminUserHandler_UpdateProfile_UnknownUserReturns404 verifies that
// UpdateProfile returns HTTP 404 when the service returns ErrUserNotFound
// (user ID does not exist in the repository).
func TestAdminUserHandler_UpdateProfile_UnknownUserReturns404(t *testing.T) {
	adminUserService := createTestAdminUserService()
	handler := NewAdminUserHandler(nil, adminUserService, createTestLogger())

	// User 999 does not exist — mock GetByID returns (nil, nil), triggering the
	// ErrUserNotFound sentinel path in AdminUserService.ensureNotProtected.
	body := fmt.Sprintf(`{"name":"Ghost","updated_at":"%s"}`, knownUpdatedAt.Format(time.RFC3339Nano))
	req := createAuthenticatedRequest(http.MethodPatch, "/api/admin/users/999", body, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.UpdateProfile(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
}

// TestAdminUserHandler_ForcePasswordReset_OnProtectedUserReturns403 verifies
// that ForcePasswordReset returns HTTP 403 with error code "protected_user" when
// the target user is in the protected-user registry.
func TestAdminUserHandler_ForcePasswordReset_OnProtectedUserReturns403(t *testing.T) {
	adminUserService := createTestAdminUserServiceWithProtectedUser()
	handler := NewAdminUserHandler(nil, adminUserService, createTestLogger())

	// User 99 has email br8kwall@gmail.com — a protected account.
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/99/force-password-reset", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "99")
	rr := httptest.NewRecorder()

	handler.ForcePasswordReset(rr, req)

	assertStatusCode(t, rr, http.StatusForbidden)

	var resp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp.Error != "protected_user" {
		t.Errorf("Error = %q, want %q", resp.Error, "protected_user")
	}
}

// stubAdminUserService is a hand-rolled test double that satisfies the handler's
// adminUserServiceIface. It records the most recent call's arguments and returns
// caller-controlled results, so each test can wire in the exact behaviour it
// needs without standing up the full service + repo + audit-log graph.
type stubAdminUserService struct {
	// CreateUser tracking
	createUserCalled  bool
	createUserActorID int64
	createUserFields  service.CreateUserFields
	createUserResult  *domain.User
	createUserErr     error

	// UpdateProfile tracking (unused in current tests but required by the
	// adminUserServiceIface contract).
	updateProfileCalled bool
	updateProfileResult *domain.User
	updateProfileErr    error

	// ForcePasswordReset tracking (unused in current tests but required by
	// the adminUserServiceIface contract).
	forcePasswordResetCalled bool
	forcePasswordResetErr    error
}

func (s *stubAdminUserService) CreateUser(actorID int64, fields service.CreateUserFields) (*domain.User, error) {
	s.createUserCalled = true
	s.createUserActorID = actorID
	s.createUserFields = fields
	return s.createUserResult, s.createUserErr
}

func (s *stubAdminUserService) UpdateProfile(actorID, targetID int64, fields service.ProfileUpdateFields, ifMatchUpdatedAt time.Time) (*domain.User, error) {
	s.updateProfileCalled = true
	return s.updateProfileResult, s.updateProfileErr
}

func (s *stubAdminUserService) ForcePasswordReset(actorID, targetID int64) error {
	s.forcePasswordResetCalled = true
	return s.forcePasswordResetErr
}

// TestAdminUserHandler_CreateUser_201Success exercises the happy path: valid
// JSON in, 201 + user JSON out, service.CreateUser called with the decoded fields.
func TestAdminUserHandler_CreateUser_201Success(t *testing.T) {
	stub := &stubAdminUserService{
		createUserResult: &domain.User{
			ID:    42,
			Email: "newathlete@example.com",
			Name:  "New",
			Role:  "athlete",
		},
	}
	h := NewAdminUserHandler(nil, stub, nil)

	body := `{"email":"newathlete@example.com","password":"ValidPass123Long","name":"New","role":"athlete","email_verified":true}`
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users", body, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Code = %d, want 201; body=%s", rr.Code, rr.Body.String())
	}
	if !stub.createUserCalled {
		t.Error("service.CreateUser not invoked")
	}
	if stub.createUserActorID != 1 {
		t.Errorf("actorID = %d, want 1", stub.createUserActorID)
	}
	if stub.createUserFields.Email != "newathlete@example.com" {
		t.Errorf("Email = %q", stub.createUserFields.Email)
	}
	if !stub.createUserFields.EmailVerified {
		t.Error("EmailVerified should be true (decoded from request)")
	}
}

// TestAdminUserHandler_CreateUser_400OnInvalidJSON returns a clear bad-request
// status without invoking the service.
func TestAdminUserHandler_CreateUser_400OnInvalidJSON(t *testing.T) {
	stub := &stubAdminUserService{}
	h := NewAdminUserHandler(nil, stub, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users", `{not json`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want 400", rr.Code)
	}
	if stub.createUserCalled {
		t.Error("service should not be called on JSON decode failure")
	}
}

// TestAdminUserHandler_CreateUser_400OnValidationError surfaces InvalidInputError as 400.
func TestAdminUserHandler_CreateUser_400OnValidationError(t *testing.T) {
	stub := &stubAdminUserService{
		createUserErr: &domain.InvalidInputError{Field: "password", Message: "must be at least 12 characters"},
	}
	h := NewAdminUserHandler(nil, stub, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users",
		`{"email":"a@b.com","password":"x","name":"X","role":"athlete"}`,
		1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Code = %d, want 400", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "password") {
		t.Errorf("body should include the field name; got %s", rr.Body.String())
	}
}

// TestAdminUserHandler_CreateUser_409OnDuplicateEmail maps the service sentinel
// to a 409 with the structured error code "duplicate_email".
func TestAdminUserHandler_CreateUser_409OnDuplicateEmail(t *testing.T) {
	stub := &stubAdminUserService{createUserErr: service.ErrEmailAlreadyExists}
	h := NewAdminUserHandler(nil, stub, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users",
		`{"email":"a@b.com","password":"ValidPass123Long","name":"X","role":"athlete"}`,
		1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	h.CreateUser(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("Code = %d, want 409; body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "duplicate_email") {
		t.Errorf("body should include error code 'duplicate_email'; got %s", rr.Body.String())
	}
}
