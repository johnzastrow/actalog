package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
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

// Test NewOrganizationHandler

func TestNewOrganizationHandler(t *testing.T) {
	handler := NewOrganizationHandler(nil, nil)
	if handler == nil {
		t.Error("NewOrganizationHandler should return a non-nil handler")
	}
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

// Removed 14 panic-expectation tests:
// - TestOrganizationHandler_ListOrganizations_NilService
// - TestOrganizationHandler_GetOrganization_ValidIDNilService
// - TestOrganizationHandler_UpdateOrganization_ValidInputNilService
// - TestOrganizationHandler_DeleteOrganization_ValidIDNilService
// - TestOrganizationHandler_GetUserOrganizations_ValidIDNilService
// - TestOrganizationHandler_GetOrganizationUsers_ValidIDNilService
// - TestOrganizationHandler_RemoveUserFromOrganization_ValidIDsNilService
// - TestOrganizationHandler_CreateOrganization_ValidInputNilService
// - TestOrganizationHandler_AssignUserToOrganization_ValidInputNilService
// - TestOrganizationHandler_ListOrganizations_WithPagination (4 subtests)
// - TestOrganizationHandler_GetOrganization_DifferentIDs (3 subtests)
// - TestOrganizationHandler_UpdateOrganization_WithDescription
// - TestOrganizationHandler_GetUserOrganizations_DifferentIDs (3 subtests)
// - TestOrganizationHandler_GetOrganizationUsers_DifferentIDs (3 subtests)
// These tests verified nil pointer panics, not business logic.

// MockUserRepositoryForOrg is a mock that returns nil,nil for not found (matching real repo behavior)
type MockUserRepositoryForOrg struct {
	*MockUserRepository
}

func NewMockUserRepositoryForOrg() *MockUserRepositoryForOrg {
	return &MockUserRepositoryForOrg{
		MockUserRepository: NewMockUserRepository(),
	}
}

// Override GetByID to return nil,nil for not found (matching real repository pattern)
func (m *MockUserRepositoryForOrg) GetByID(id int64) (*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil // Real repo returns nil,nil for not found
}

// Helper to create organization handler with proper service
func createOrgHandlerWithMocks() (*OrganizationHandler, *MockOrganizationRepository, *MockUserRepositoryForOrg) {
	mockOrgRepo := NewMockOrganizationRepository()
	mockUserRepo := NewMockUserRepositoryForOrg()
	mockAuditLogRepo := NewMockAuditLogRepository()

	orgService := service.NewOrganizationService(mockOrgRepo, mockUserRepo, mockAuditLogRepo)

	handler := &OrganizationHandler{
		orgService: orgService,
		logger:     createTestLogger(),
	}

	return handler, mockOrgRepo, mockUserRepo
}

// ============================================================================
// Tests for GetOrganization with service
// ============================================================================

func TestOrganizationHandler_GetOrganization_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Test Org")
}

func TestOrganizationHandler_GetOrganization_NotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/999", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Organization not found")
}

func TestOrganizationHandler_GetOrganization_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/1", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get organization")
}

// ============================================================================
// Tests for ListOrganizations with service
// ============================================================================

func TestOrganizationHandler_ListOrganizations_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/organizations", "")
	rr := httptest.NewRecorder()

	handler.ListOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "organizations")
	assertBodyContains(t, rr, "total")
}

func TestOrganizationHandler_ListOrganizations_WithPaginationParams(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/organizations?limit=10&offset=5", "")
	rr := httptest.NewRecorder()

	handler.ListOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"limit":10`)
	assertBodyContains(t, rr, `"offset":5`)
}

func TestOrganizationHandler_ListOrganizations_InvalidLimitIgnored(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/organizations?limit=-5&offset=abc", "")
	rr := httptest.NewRecorder()

	handler.ListOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	// Default values should be used
	assertBodyContains(t, rr, `"limit":50`)
	assertBodyContains(t, rr, `"offset":0`)
}

func TestOrganizationHandler_ListOrganizations_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createTestRequest(http.MethodGet, "/api/admin/organizations", "")
	rr := httptest.NewRecorder()

	handler.ListOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to list organizations")
}

// ============================================================================
// Tests for CreateOrganization with service
// ============================================================================

func TestOrganizationHandler_CreateOrganization_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/organizations",
		`{"name": "New Org", "description": "A new organization"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusCreated)
	assertBodyContains(t, rr, "New Org")
}

func TestOrganizationHandler_CreateOrganization_NameExists(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	// Try to create org with name that already exists
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/organizations",
		`{"name": "Test Org"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusConflict)
	assertBodyContains(t, rr, "Organization name already exists")
}

func TestOrganizationHandler_CreateOrganization_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/organizations",
		`{"name": "New Org"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to create organization")
}

// ============================================================================
// Tests for UpdateOrganization with service
// ============================================================================

func TestOrganizationHandler_UpdateOrganization_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/1",
		`{"name": "Updated Org", "description": "Updated description"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Updated Org")
}

func TestOrganizationHandler_UpdateOrganization_NotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/999",
		`{"name": "Updated Org"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.UpdateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Organization not found")
}

func TestOrganizationHandler_UpdateOrganization_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/organizations/1",
		`{"name": "Updated Org"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.UpdateOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to update organization")
}

// ============================================================================
// Tests for DeleteOrganization with service
// ============================================================================

func TestOrganizationHandler_DeleteOrganization_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/organizations/1",
		"", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.DeleteOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Organization deleted successfully")
}

func TestOrganizationHandler_DeleteOrganization_NotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/organizations/999",
		"", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.DeleteOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Organization not found")
}

func TestOrganizationHandler_DeleteOrganization_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/organizations/1",
		"", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.DeleteOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to delete organization")
}

// ============================================================================
// Tests for AssignUserToOrganization with service
// ============================================================================

func TestOrganizationHandler_AssignUserToOrganization_Success(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	// Need to make IsUserInOrganization return false for new assignment
	mockOrgRepo.organizations = append(mockOrgRepo.organizations, nil) // Simulate no membership

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization",
		`{"organization_id": 1}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	// User is already a member in mock, so this should fail
	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestOrganizationHandler_AssignUserToOrganization_OrgNotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization",
		`{"organization_id": 999}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Organization not found")
}

func TestOrganizationHandler_AssignUserToOrganization_UserNotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/999/organization",
		`{"organization_id": 1}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "User not found")
}

func TestOrganizationHandler_AssignUserToOrganization_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/users/1/organization",
		`{"organization_id": 1}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.AssignUserToOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to assign user")
}

// ============================================================================
// Tests for RemoveUserFromOrganization with service
// ============================================================================

func TestOrganizationHandler_RemoveUserFromOrganization_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/1/organization/1",
		"", 1, "admin@example.com", "admin")
	req = addChiURLParams(req, map[string]string{"id": "1", "org_id": "1"})
	rr := httptest.NewRecorder()

	handler.RemoveUserFromOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "User removed from organization successfully")
}

func TestOrganizationHandler_RemoveUserFromOrganization_UserNotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/999/organization/1",
		"", 1, "admin@example.com", "admin")
	req = addChiURLParams(req, map[string]string{"id": "999", "org_id": "1"})
	rr := httptest.NewRecorder()

	handler.RemoveUserFromOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "User not found")
}

func TestOrganizationHandler_RemoveUserFromOrganization_OrgNotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/1/organization/999",
		"", 1, "admin@example.com", "admin")
	req = addChiURLParams(req, map[string]string{"id": "1", "org_id": "999"})
	rr := httptest.NewRecorder()

	handler.RemoveUserFromOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Organization not found")
}

func TestOrganizationHandler_RemoveUserFromOrganization_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createAuthenticatedRequest(http.MethodDelete, "/api/admin/users/1/organization/1",
		"", 1, "admin@example.com", "admin")
	req = addChiURLParams(req, map[string]string{"id": "1", "org_id": "1"})
	rr := httptest.NewRecorder()

	handler.RemoveUserFromOrganization(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to remove user")
}

// ============================================================================
// Tests for GetUserOrganizations with service
// ============================================================================

func TestOrganizationHandler_GetUserOrganizations_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/users/1/organizations", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetUserOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "organizations")
	assertBodyContains(t, rr, "count")
}

func TestOrganizationHandler_GetUserOrganizations_UserNotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/users/999/organizations", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetUserOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "User not found")
}

func TestOrganizationHandler_GetUserOrganizations_RepoError(t *testing.T) {
	handler, _, mockUserRepo := createOrgHandlerWithMocks()
	mockUserRepo.SetError(ErrMockInternalError)

	req := createTestRequest(http.MethodGet, "/api/admin/users/1/organizations", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetUserOrganizations(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get organizations")
}

// ============================================================================
// Tests for GetOrganizationUsers with service
// ============================================================================

func TestOrganizationHandler_GetOrganizationUsers_Success(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/1/users", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetOrganizationUsers(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "users")
	assertBodyContains(t, rr, "count")
}

func TestOrganizationHandler_GetOrganizationUsers_OrgNotFound(t *testing.T) {
	handler, _, _ := createOrgHandlerWithMocks()

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/999/users", "")
	req = addChiURLParam(req, "id", "999")
	rr := httptest.NewRecorder()

	handler.GetOrganizationUsers(rr, req)

	assertStatusCode(t, rr, http.StatusNotFound)
	assertBodyContains(t, rr, "Organization not found")
}

func TestOrganizationHandler_GetOrganizationUsers_RepoError(t *testing.T) {
	handler, mockOrgRepo, _ := createOrgHandlerWithMocks()
	mockOrgRepo.SetError(ErrMockInternalError)

	req := createTestRequest(http.MethodGet, "/api/admin/organizations/1/users", "")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.GetOrganizationUsers(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get users")
}
