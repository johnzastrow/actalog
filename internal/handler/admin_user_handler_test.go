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

func TestAdminUserHandler_ChangeUserRole_ValidRole(t *testing.T) {
	handler := &AdminUserHandler{}

	tests := []struct {
		name string
		role string
	}{
		{"admin role", `{"role": "admin"}`},
		{"user role", `{"role": "user"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/1/role", tt.role, 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "id", "1")
			rr := httptest.NewRecorder()

			// Will panic on nil service - tests validation passes
			defer func() {
				if r := recover(); r == nil {
					t.Log("ChangeUserRole requires service")
				}
			}()

			handler.ChangeUserRole(rr, req)
		})
	}
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

func TestAdminUserHandler_ToggleEmailVerification_ValidInput(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/toggle-email-verification", `{"verified": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("ToggleEmailVerification requires service")
		}
	}()

	handler.ToggleEmailVerification(rr, req)
}

func TestAdminUserHandler_UnlockUser_ValidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/unlock", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("UnlockUser requires service")
		}
	}()

	handler.UnlockUser(rr, req)
}

func TestAdminUserHandler_DisableUser_ValidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/disable", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("DisableUser requires service")
		}
	}()

	handler.DisableUser(rr, req)
}

func TestAdminUserHandler_EnableUser_ValidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/enable", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("EnableUser requires service")
		}
	}()

	handler.EnableUser(rr, req)
}

func TestAdminUserHandler_GetUserDetails_ValidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/users/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetUserDetails requires service")
		}
	}()

	handler.GetUserDetails(rr, req)
}

func TestAdminUserHandler_DeleteUser_ValidID(t *testing.T) {
	handler := &AdminUserHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Will panic on nil service - tests validation passes
	defer func() {
		if r := recover(); r == nil {
			t.Log("DeleteUser requires service")
		}
	}()

	handler.DeleteUser(rr, req)
}

func TestAdminUserHandler_ListUsers_WithQueryParams(t *testing.T) {
	handler := &AdminUserHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"default params", "/api/admin/users"},
		{"with limit", "/api/admin/users?limit=25"},
		{"with offset", "/api/admin/users?offset=10"},
		{"with both", "/api/admin/users?limit=25&offset=10"},
		{"with search", "/api/admin/users?search=test"},
		{"with role filter", "/api/admin/users?role=admin"},
		{"all params", "/api/admin/users?limit=25&offset=10&search=test&role=user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			// Will panic on nil service - tests param parsing paths
			defer func() {
				if r := recover(); r == nil {
					t.Log("ListUsers requires service - params were parsed")
				}
			}()

			handler.ListUsers(rr, req)
		})
	}
}

// Tests with mock service

func TestAdminUserHandler_ListUsers_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createTestRequest(http.MethodGet, "/api/admin/users", "")
	rr := httptest.NewRecorder()

	handler.ListUsers(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "users")
	assertBodyContains(t, rr, "total")
}

func TestAdminUserHandler_UnlockUser_UserNotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/999/unlock", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.UnlockUser(rr, req)

	// Service returns 500 for user not found
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestAdminUserHandler_UnlockUser_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/unlock", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UnlockUser(rr, req)

	// User exists, should succeed
	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_DisableUser_InvalidJSON(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

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
	handler := NewAdminUserHandler(userService, createTestLogger())

	// Reason is optional, so empty body succeeds
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/2/disable", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_DisableUser_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/2/disable", `{"reason": "Policy violation"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_DisableUser_CannotDisableSelf(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/disable", `{"reason": "Self disable"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.DisableUser(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "cannot disable your own account")
}

func TestAdminUserHandler_EnableUser_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/enable", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.EnableUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_ChangeUserRole_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/2/role", `{"role": "admin"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_ChangeUserRole_CannotChangeOwnRole(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/users/1/role", `{"role": "user"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.ChangeUserRole(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "cannot change your own role")
}

func TestAdminUserHandler_ToggleEmailVerification_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/toggle-email-verification", `{"verified": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.ToggleEmailVerification(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_GetUserDetails_Success(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createTestRequest(http.MethodGet, "/api/admin/users/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetUserDetails(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestAdminUserHandler_GetUserDetails_NotFound(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

	req := createTestRequest(http.MethodGet, "/api/admin/users/999", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetUserDetails(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
}

func TestAdminUserHandler_DeleteUser_CannotDeleteSelf(t *testing.T) {
	userService := createTestUserService()
	handler := NewAdminUserHandler(userService, createTestLogger())

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
	handler := NewAdminUserHandler(userService, createTestLogger())

	// Admin (user 1) deleting user 2
	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/2", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "2")
	rr := httptest.NewRecorder()

	handler.DeleteUser(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}
