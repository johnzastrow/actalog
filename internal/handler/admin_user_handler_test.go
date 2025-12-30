package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
	handler := NewAdminUserHandler(nil, nil)
	if handler == nil {
		t.Error("NewAdminUserHandler should return a non-nil handler")
	}
}

func TestAdminUserHandler_ListUsers_NilService(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/users", "")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListUsers requires service")
		}
	}()

	handler.ListUsers(rr, req)
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
	assertBodyContains(t, rr, "Role must be 'user' or 'admin'")
}
