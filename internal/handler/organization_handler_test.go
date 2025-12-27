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
