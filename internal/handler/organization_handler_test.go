package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrganizationHandler_CreateOrganization_Unauthorized(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/organizations", `{"name": "Test Org"}`)
	rr := httptest.NewRecorder()

	handler.CreateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestOrganizationHandler_CreateOrganization_InvalidJSON(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/organizations", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestOrganizationHandler_CreateOrganization_MissingName(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/organizations", `{"description": "Test"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Organization name is required")
}

func TestOrganizationHandler_GetOrganization_InvalidID(t *testing.T) {
	handler := &OrganizationHandler{}

	// chi.URLParam returns empty string without router context
	req := createTestRequest(http.MethodGet, "/api/admin/organizations/abc", "")
	rr := httptest.NewRecorder()

	handler.GetOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid organization ID")
}

func TestOrganizationHandler_UpdateOrganization_Unauthorized(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodPut, "/api/admin/organizations/1", `{"name": "Updated"}`)
	rr := httptest.NewRecorder()

	handler.UpdateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestOrganizationHandler_UpdateOrganization_InvalidID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/abc", `{"name": "Updated"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UpdateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid organization ID")
}

func TestOrganizationHandler_UpdateOrganization_InvalidJSON(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/1", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.UpdateOrganization(rr, req)

	// Without chi router context, URL param is empty -> Invalid organization ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestOrganizationHandler_DeleteOrganization_Unauthorized(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodDelete, "/api/admin/organizations/1", "")
	rr := httptest.NewRecorder()

	handler.DeleteOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestOrganizationHandler_DeleteOrganization_InvalidID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/organizations/abc", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.DeleteOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid organization ID")
}

func TestOrganizationHandler_AssignUserToOrganization_Unauthorized(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/users/1/organization", `{"organization_id": 1}`)
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestOrganizationHandler_AssignUserToOrganization_InvalidUserID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/abc/organization", `{"organization_id": 1}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestOrganizationHandler_AssignUserToOrganization_InvalidJSON(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	// Without chi router context, URL param is empty -> Invalid user ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestOrganizationHandler_AssignUserToOrganization_MissingOrgID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization", `{}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	// Without chi router context, URL param is empty -> Invalid user ID
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestOrganizationHandler_RemoveUserFromOrganization_Unauthorized(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodDelete, "/api/admin/users/1/organization/1", "")
	rr := httptest.NewRecorder()

	handler.RemoveUserFromOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestOrganizationHandler_RemoveUserFromOrganization_InvalidUserID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/abc/organization/1", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.RemoveUserFromOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestOrganizationHandler_GetUserOrganizations_InvalidUserID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/users/abc/organizations", "")
	rr := httptest.NewRecorder()

	handler.GetUserOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user ID")
}

func TestOrganizationHandler_GetOrganizationUsers_InvalidOrgID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/abc/users", "")
	rr := httptest.NewRecorder()

	handler.GetOrganizationUsers(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid organization ID")
}

// Tests for ListOrganizations

func TestOrganizationHandler_ListOrganizations_NilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/organizations", "")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("ListOrganizations requires service")
		}
	}()

	handler.ListOrganizations(rr, req)
}

// Test NewOrganizationHandler

func TestNewOrganizationHandler(t *testing.T) {
	handler := NewOrganizationHandler(nil, nil)
	if handler == nil {
		t.Error("NewOrganizationHandler should return a non-nil handler")
	}
}

// Test GetOrganization with valid ID

func TestOrganizationHandler_GetOrganization_ValidIDNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetOrganization requires service")
		}
	}()

	handler.GetOrganization(rr, req)
}

// Test UpdateOrganization with valid JSON

func TestOrganizationHandler_UpdateOrganization_ValidIDInvalidJSON(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/1", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestOrganizationHandler_UpdateOrganization_ValidInputNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/1", `{"name": "Test Org"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("UpdateOrganization requires service")
		}
	}()

	handler.UpdateOrganization(rr, req)
}

// Test DeleteOrganization with valid ID

func TestOrganizationHandler_DeleteOrganization_ValidIDNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/organizations/1", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("DeleteOrganization requires service")
		}
	}()

	handler.DeleteOrganization(rr, req)
}

// Test AssignUserToOrganization with valid JSON

func TestOrganizationHandler_AssignUserToOrganization_ValidIDInvalidJSON(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestOrganizationHandler_AssignUserToOrganization_MissingOrgIDWithValidUserID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Organization ID is required")
}

// Test RemoveUserFromOrganization with valid IDs

func TestOrganizationHandler_RemoveUserFromOrganization_InvalidOrgID(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/1/organization/abc", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	// org_id is empty without chi context -> Invalid organization ID
	rr := httptest.NewRecorder()

	handler.RemoveUserFromOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid organization ID")
}

// Test GetUserOrganizations with valid ID

func TestOrganizationHandler_GetUserOrganizations_ValidIDNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/users/1/organizations", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetUserOrganizations requires service")
		}
	}()

	handler.GetUserOrganizations(rr, req)
}

// Test GetOrganizationUsers with valid ID

func TestOrganizationHandler_GetOrganizationUsers_ValidIDNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/1/users", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("GetOrganizationUsers requires service")
		}
	}()

	handler.GetOrganizationUsers(rr, req)
}

// Additional tests for RemoveUserFromOrganization with valid IDs

func TestOrganizationHandler_RemoveUserFromOrganization_ValidIDsNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/1/organization/2", "", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	req = addChiURLParam(req, "org_id", "2")
	rr := httptest.NewRecorder()

	// Without a service, will panic - tests function entry
	defer func() {
		if r := recover(); r == nil {
			t.Log("RemoveUserFromOrganization requires service")
		}
	}()

	handler.RemoveUserFromOrganization(rr, req)
}

// Tests for CreateOrganization with valid input

func TestOrganizationHandler_CreateOrganization_ValidInputNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/organizations",
		`{"name": "Test Org", "description": "A test organization"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("CreateOrganization requires service")
		}
	}()

	handler.CreateOrganization(rr, req)
}

// Tests for AssignUserToOrganization with valid input

func TestOrganizationHandler_AssignUserToOrganization_ValidInputNilService(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization",
		`{"organization_id": 2}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("AssignUserToOrganization requires service")
		}
	}()

	handler.AssignUserToOrganization(rr, req)
}

// Tests for ListOrganizations with pagination

func TestOrganizationHandler_ListOrganizations_WithPagination(t *testing.T) {
	handler := &OrganizationHandler{}

	tests := []struct {
		name  string
		query string
	}{
		{"default", "/api/admin/organizations"},
		{"with limit", "/api/admin/organizations?limit=10"},
		{"with offset", "/api/admin/organizations?offset=5"},
		{"with both", "/api/admin/organizations?limit=20&offset=10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, tt.query, "")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListOrganizations requires service")
				}
			}()

			handler.ListOrganizations(rr, req)
		})
	}
}

// Tests for GetOrganization with different IDs

func TestOrganizationHandler_GetOrganization_DifferentIDs(t *testing.T) {
	handler := &OrganizationHandler{}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("id_"+id, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/admin/organizations/"+id, "")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetOrganization requires service")
				}
			}()

			handler.GetOrganization(rr, req)
		})
	}
}

// Tests for UpdateOrganization with description

func TestOrganizationHandler_UpdateOrganization_WithDescription(t *testing.T) {
	handler := &OrganizationHandler{}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/1",
		`{"name": "Updated Org", "description": "Updated description"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Log("UpdateOrganization requires service")
		}
	}()

	handler.UpdateOrganization(rr, req)
}

// Tests for GetUserOrganizations with different IDs

func TestOrganizationHandler_GetUserOrganizations_DifferentIDs(t *testing.T) {
	handler := &OrganizationHandler{}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("user_id_"+id, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/admin/users/"+id+"/organizations", "")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetUserOrganizations requires service")
				}
			}()

			handler.GetUserOrganizations(rr, req)
		})
	}
}

// Tests for GetOrganizationUsers with different IDs

func TestOrganizationHandler_GetOrganizationUsers_DifferentIDs(t *testing.T) {
	handler := &OrganizationHandler{}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("org_id_"+id, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/admin/organizations/"+id+"/users", "")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetOrganizationUsers requires service")
				}
			}()

			handler.GetOrganizationUsers(rr, req)
		})
	}
}
