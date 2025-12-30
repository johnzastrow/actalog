package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

type OrganizationHandler struct {
	orgService *service.OrganizationService
	logger     *logger.Logger
}

func NewOrganizationHandler(orgService *service.OrganizationService, logger *logger.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
		logger:     logger,
	}
}

// CreateOrganization handles POST /api/admin/organizations
// @Summary      Create organization (Admin)
// @Description  Create a new organization (admin only)
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body object true "Organization name and description"
// @Success      201 {object} domain.Organization "Created organization"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      409 {object} ErrorResponse "Organization name exists"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/organizations [post]
func (h *OrganizationHandler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Organization name is required")
		return
	}

	org, err := h.orgService.Create(adminUserID, req.Name, req.Description)
	if err != nil {
		if err == service.ErrOrganizationNameExists {
			respondError(w, http.StatusConflict, "Organization name already exists")
			return
		}
		h.logger.Error("Failed to create organization: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	respondJSON(w, http.StatusCreated, org)
}

// ListOrganizations handles GET /api/admin/organizations
// @Summary      List organizations (Admin)
// @Description  Get a paginated list of all organizations
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Max results (default 50)"
// @Param        offset query int false "Skip N results (default 0)"
// @Success      200 {object} map[string]interface{} "Organizations list with total count"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/organizations [get]
func (h *OrganizationHandler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	orgs, total, err := h.orgService.List(limit, offset)
	if err != nil {
		h.logger.Error("Failed to list organizations: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list organizations")
		return
	}

	response := map[string]interface{}{
		"organizations": orgs,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetOrganization handles GET /api/admin/organizations/:id
// @Summary      Get organization (Admin)
// @Description  Get details of a specific organization
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Organization ID"
// @Success      200 {object} domain.Organization "Organization details"
// @Failure      400 {object} ErrorResponse "Invalid organization ID"
// @Failure      404 {object} ErrorResponse "Organization not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	org, err := h.orgService.GetByID(id)
	if err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		h.logger.Error("Failed to get organization: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get organization")
		return
	}

	respondJSON(w, http.StatusOK, org)
}

// UpdateOrganization handles PUT /api/admin/organizations/:id
// @Summary      Update organization (Admin)
// @Description  Update an existing organization's details
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Organization ID"
// @Param        request body object true "Updated organization details"
// @Success      200 {object} domain.Organization "Updated organization"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      404 {object} ErrorResponse "Organization not found"
// @Failure      409 {object} ErrorResponse "Organization name exists"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	org, err := h.orgService.Update(adminUserID, id, req.Name, req.Description)
	if err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		if err == service.ErrOrganizationNameExists {
			respondError(w, http.StatusConflict, "Organization name already exists")
			return
		}
		h.logger.Error("Failed to update organization: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update organization")
		return
	}

	respondJSON(w, http.StatusOK, org)
}

// DeleteOrganization handles DELETE /api/admin/organizations/:id
// @Summary      Delete organization (Admin)
// @Description  Delete an organization (must have no assigned users)
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Organization ID"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid organization ID"
// @Failure      404 {object} ErrorResponse "Organization not found"
// @Failure      409 {object} ErrorResponse "Organization has assigned users"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	err = h.orgService.Delete(adminUserID, id)
	if err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		if err == service.ErrOrganizationHasUsers {
			respondError(w, http.StatusConflict, "Cannot delete organization with assigned users")
			return
		}
		h.logger.Error("Failed to delete organization: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete organization")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Organization deleted successfully",
	})
}

// AssignUserToOrganization handles POST /api/admin/users/:id/organization
// @Summary      Assign user to organization (Admin)
// @Description  Assign a user to an organization
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Param        request body object true "Organization ID to assign"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      404 {object} ErrorResponse "User or organization not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/organization [post]
func (h *OrganizationHandler) AssignUserToOrganization(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		OrganizationID int64 `json:"organization_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.OrganizationID == 0 {
		respondError(w, http.StatusBadRequest, "Organization ID is required")
		return
	}

	err = h.orgService.AssignUserToOrganization(adminUserID, userID, req.OrganizationID)
	if err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		if err == service.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		h.logger.Error("Failed to assign user to organization: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to assign user")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "User assigned to organization successfully",
	})
}

// RemoveUserFromOrganization handles DELETE /api/admin/users/:id/organization/:org_id
// @Summary      Remove user from organization (Admin)
// @Description  Remove a user from an organization
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Param        org_id path int true "Organization ID"
// @Success      200 {object} MessageResponse "Success message"
// @Failure      400 {object} ErrorResponse "Invalid ID"
// @Failure      404 {object} ErrorResponse "User or organization not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/organization/{org_id} [delete]
func (h *OrganizationHandler) RemoveUserFromOrganization(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	orgIDStr := chi.URLParam(r, "org_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	err = h.orgService.RemoveUserFromOrganization(adminUserID, userID, orgID)
	if err != nil {
		if err == service.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		h.logger.Error("Failed to remove user from organization: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to remove user")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "User removed from organization successfully",
	})
}

// GetUserOrganizations handles GET /api/admin/users/:id/organizations
// @Summary      Get user organizations (Admin)
// @Description  Get all organizations a user belongs to
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "User ID"
// @Success      200 {object} map[string]interface{} "List of organizations"
// @Failure      400 {object} ErrorResponse "Invalid user ID"
// @Failure      404 {object} ErrorResponse "User not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/users/{id}/organizations [get]
func (h *OrganizationHandler) GetUserOrganizations(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	orgs, err := h.orgService.GetUserOrganizations(userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		h.logger.Error("Failed to get user organizations: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get organizations")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"organizations": orgs,
		"count":         len(orgs),
	})
}

// GetOrganizationUsers handles GET /api/admin/organizations/:id/users
// @Summary      Get organization users (Admin)
// @Description  Get all users belonging to an organization
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Organization ID"
// @Success      200 {object} map[string]interface{} "List of users"
// @Failure      400 {object} ErrorResponse "Invalid organization ID"
// @Failure      404 {object} ErrorResponse "Organization not found"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /admin/organizations/{id}/users [get]
func (h *OrganizationHandler) GetOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid organization ID")
		return
	}

	users, err := h.orgService.GetOrganizationUsers(orgID)
	if err != nil {
		if err == service.ErrOrganizationNotFound {
			respondError(w, http.StatusNotFound, "Organization not found")
			return
		}
		h.logger.Error("Failed to get organization users: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}
