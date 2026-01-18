package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

type Phase4Handler struct {
	phase4Service *service.Phase4Service
	logger        *logger.Logger
}

func NewPhase4Handler(phase4Service *service.Phase4Service, logger *logger.Logger) *Phase4Handler {
	return &Phase4Handler{
		phase4Service: phase4Service,
		logger:        logger,
	}
}

// ============================================
// DOCUMENT HANDLERS (Admin)
// ============================================

// CreateDocument handles POST /api/gyms/{gym_id}/documents
func (h *Phase4Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	var req struct {
		Name             string  `json:"name"`
		Description      *string `json:"description"`
		DocumentType     string  `json:"document_type"`
		URL              *string `json:"url"`
		IsRequired       bool    `json:"is_required"`
		ExpiresAfterDays *int    `json:"expires_after_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Document name is required")
		return
	}

	if req.DocumentType == "" {
		req.DocumentType = domain.DocumentTypeOther
	}

	doc := &domain.Document{
		OrganizationID:   gymID,
		Name:             req.Name,
		Description:      req.Description,
		DocumentType:     req.DocumentType,
		URL:              req.URL,
		IsRequired:       req.IsRequired,
		ExpiresAfterDays: req.ExpiresAfterDays,
		IsActive:         true,
	}

	if err := h.phase4Service.CreateDocument(userID, doc); err != nil {
		h.logger.Error("Failed to create document: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create document")
		return
	}

	respondJSON(w, http.StatusCreated, doc)
}

// ListDocuments handles GET /api/gyms/{gym_id}/documents
func (h *Phase4Handler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	docs, err := h.phase4Service.GetDocumentsByOrganization(gymID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list documents: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list documents")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"documents": docs,
		"count":     len(docs),
	})
}

// GetDocument handles GET /api/gyms/{gym_id}/documents/{id}
func (h *Phase4Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	doc, err := h.phase4Service.GetDocumentByID(id)
	if err != nil {
		if err == service.ErrDocumentNotFound {
			respondError(w, http.StatusNotFound, "Document not found")
			return
		}
		h.logger.Error("Failed to get document: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get document")
		return
	}

	respondJSON(w, http.StatusOK, doc)
}

// UpdateDocument handles PUT /api/gyms/{gym_id}/documents/{id}
func (h *Phase4Handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	doc, err := h.phase4Service.GetDocumentByID(id)
	if err != nil {
		if err == service.ErrDocumentNotFound {
			respondError(w, http.StatusNotFound, "Document not found")
			return
		}
		h.logger.Error("Failed to get document: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get document")
		return
	}

	var req struct {
		Name             string  `json:"name"`
		Description      *string `json:"description"`
		DocumentType     string  `json:"document_type"`
		URL              *string `json:"url"`
		IsRequired       bool    `json:"is_required"`
		ExpiresAfterDays *int    `json:"expires_after_days"`
		IsActive         bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	doc.Name = req.Name
	doc.Description = req.Description
	doc.DocumentType = req.DocumentType
	doc.URL = req.URL
	doc.IsRequired = req.IsRequired
	doc.ExpiresAfterDays = req.ExpiresAfterDays
	doc.IsActive = req.IsActive

	if err := h.phase4Service.UpdateDocument(userID, doc); err != nil {
		h.logger.Error("Failed to update document: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update document")
		return
	}

	respondJSON(w, http.StatusOK, doc)
}

// DeleteDocument handles DELETE /api/gyms/{gym_id}/documents/{id}
func (h *Phase4Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid document ID")
		return
	}

	if err := h.phase4Service.DeleteDocument(userID, id); err != nil {
		if err == service.ErrDocumentNotFound {
			respondError(w, http.StatusNotFound, "Document not found")
			return
		}
		h.logger.Error("Failed to delete document: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete document")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Document deleted"})
}

// ============================================
// USER DOCUMENT HANDLERS
// ============================================

// GetUserDocuments handles GET /api/users/me/documents
func (h *Phase4Handler) GetUserDocuments(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	docs, err := h.phase4Service.GetUserDocuments(userID)
	if err != nil {
		h.logger.Error("Failed to get user documents: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user documents")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_documents": docs,
		"count":          len(docs),
	})
}

// GetUserDocumentsByOrganization handles GET /api/gyms/{gym_id}/users/me/documents
func (h *Phase4Handler) GetUserDocumentsByOrganization(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	docs, err := h.phase4Service.GetUserDocumentsByOrganization(userID, gymID)
	if err != nil {
		h.logger.Error("Failed to get user documents: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user documents")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_documents": docs,
		"count":          len(docs),
	})
}

// GetPendingDocuments handles GET /api/gyms/{gym_id}/users/me/documents/pending
func (h *Phase4Handler) GetPendingDocuments(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	docs, err := h.phase4Service.GetPendingUserDocuments(userID, gymID)
	if err != nil {
		h.logger.Error("Failed to get pending documents: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get pending documents")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pending_documents": docs,
		"count":             len(docs),
	})
}

// MarkDocumentCompleted handles POST /api/gyms/{gym_id}/user-documents/{id}/complete
func (h *Phase4Handler) MarkDocumentCompleted(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user document ID")
		return
	}

	if err := h.phase4Service.MarkDocumentCompleted(adminUserID, id); err != nil {
		if err == service.ErrUserDocumentNotFound {
			respondError(w, http.StatusNotFound, "User document not found")
			return
		}
		h.logger.Error("Failed to mark document completed: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to mark document completed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Document marked as completed"})
}

// InitializeUserDocuments handles POST /api/gyms/{gym_id}/users/{user_id}/documents/init
func (h *Phase4Handler) InitializeUserDocuments(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	targetUserIDStr := chi.URLParam(r, "user_id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.phase4Service.InitializeUserDocuments(targetUserID, gymID); err != nil {
		h.logger.Error("Failed to initialize user documents: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to initialize user documents")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User documents initialized"})
}

// ============================================
// CLASS PACKAGE HANDLERS (Admin)
// ============================================

// CreateClassPackage handles POST /api/gyms/{gym_id}/packages
func (h *Phase4Handler) CreateClassPackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	var req struct {
		Name         string  `json:"name"`
		Description  *string `json:"description"`
		Credits      int     `json:"credits"`
		PriceCents   int     `json:"price_cents"`
		ValidityDays *int    `json:"validity_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Package name is required")
		return
	}

	if req.Credits <= 0 {
		respondError(w, http.StatusBadRequest, "Credits must be positive")
		return
	}

	pkg := &domain.ClassPackage{
		OrganizationID: gymID,
		Name:           req.Name,
		Description:    req.Description,
		Credits:        req.Credits,
		PriceCents:     req.PriceCents,
		ValidityDays:   req.ValidityDays,
		IsActive:       true,
	}

	if err := h.phase4Service.CreatePackage(userID, pkg); err != nil {
		h.logger.Error("Failed to create class package: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create class package")
		return
	}

	respondJSON(w, http.StatusCreated, pkg)
}

// ListClassPackages handles GET /api/gyms/{gym_id}/packages
func (h *Phase4Handler) ListClassPackages(w http.ResponseWriter, r *http.Request) {
	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	packages, err := h.phase4Service.GetPackagesByOrganization(gymID, includeInactive)
	if err != nil {
		h.logger.Error("Failed to list class packages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list class packages")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"packages": packages,
		"count":    len(packages),
	})
}

// GetClassPackage handles GET /api/gyms/{gym_id}/packages/{id}
func (h *Phase4Handler) GetClassPackage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid package ID")
		return
	}

	pkg, err := h.phase4Service.GetPackageByID(id)
	if err != nil {
		if err == service.ErrClassPackageNotFound {
			respondError(w, http.StatusNotFound, "Class package not found")
			return
		}
		h.logger.Error("Failed to get class package: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get class package")
		return
	}

	respondJSON(w, http.StatusOK, pkg)
}

// UpdateClassPackage handles PUT /api/gyms/{gym_id}/packages/{id}
func (h *Phase4Handler) UpdateClassPackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid package ID")
		return
	}

	pkg, err := h.phase4Service.GetPackageByID(id)
	if err != nil {
		if err == service.ErrClassPackageNotFound {
			respondError(w, http.StatusNotFound, "Class package not found")
			return
		}
		h.logger.Error("Failed to get class package: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get class package")
		return
	}

	var req struct {
		Name         string  `json:"name"`
		Description  *string `json:"description"`
		Credits      int     `json:"credits"`
		PriceCents   int     `json:"price_cents"`
		ValidityDays *int    `json:"validity_days"`
		IsActive     bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pkg.Name = req.Name
	pkg.Description = req.Description
	pkg.Credits = req.Credits
	pkg.PriceCents = req.PriceCents
	pkg.ValidityDays = req.ValidityDays
	pkg.IsActive = req.IsActive

	if err := h.phase4Service.UpdatePackage(userID, pkg); err != nil {
		h.logger.Error("Failed to update class package: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update class package")
		return
	}

	respondJSON(w, http.StatusOK, pkg)
}

// DeleteClassPackage handles DELETE /api/gyms/{gym_id}/packages/{id}
func (h *Phase4Handler) DeleteClassPackage(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid package ID")
		return
	}

	if err := h.phase4Service.DeletePackage(userID, id); err != nil {
		if err == service.ErrClassPackageNotFound {
			respondError(w, http.StatusNotFound, "Class package not found")
			return
		}
		h.logger.Error("Failed to delete class package: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete class package")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Class package deleted"})
}

// ============================================
// CREDITS HANDLERS
// ============================================

// PurchaseCredits handles POST /api/gyms/{gym_id}/users/{user_id}/credits
func (h *Phase4Handler) PurchaseCredits(w http.ResponseWriter, r *http.Request) {
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	targetUserIDStr := chi.URLParam(r, "user_id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		PackageID *int64 `json:"package_id"`
		Credits   int    `json:"credits"`
		Notes     string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get package credits for response if using a package
	creditsAdded := req.Credits
	if req.PackageID != nil && req.Credits == 0 {
		pkg, err := h.phase4Service.GetPackageByID(*req.PackageID)
		if err == nil && pkg != nil {
			creditsAdded = pkg.Credits
		}
	}

	if err := h.phase4Service.PurchaseCredits(adminUserID, targetUserID, gymID, req.PackageID, req.Credits, req.Notes); err != nil {
		if err == service.ErrClassPackageNotFound {
			respondError(w, http.StatusNotFound, "Class package not found")
			return
		}
		h.logger.Error("Failed to purchase credits: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to purchase credits")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Credits added successfully",
		"credits": creditsAdded,
	})
}

// GetUserCredits handles GET /api/gyms/{gym_id}/users/me/credits
func (h *Phase4Handler) GetUserCredits(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	credits, err := h.phase4Service.GetUserCreditsByOrganization(userID, gymID)
	if err != nil {
		h.logger.Error("Failed to get user credits: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user credits")
		return
	}

	available, err := h.phase4Service.GetAvailableCredits(userID, gymID)
	if err != nil {
		h.logger.Error("Failed to get available credits: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get available credits")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"credit_records":    credits,
		"count":             len(credits),
		"available_credits": available,
	})
}

// GetAvailableCredits handles GET /api/gyms/{gym_id}/users/me/credits/available
func (h *Phase4Handler) GetAvailableCredits(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	gymIDStr := chi.URLParam(r, "gym_id")
	gymID, err := strconv.ParseInt(gymIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid gym ID")
		return
	}

	available, err := h.phase4Service.GetAvailableCredits(userID, gymID)
	if err != nil {
		h.logger.Error("Failed to get available credits: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get available credits")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"available_credits": available,
	})
}

// ============================================
// WAITLIST HANDLERS
// ============================================

// JoinWaitlist handles POST /api/sessions/{session_id}/waitlist
func (h *Phase4Handler) JoinWaitlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	entry, err := h.phase4Service.JoinWaitlist(userID, sessionID)
	if err != nil {
		if err == service.ErrAlreadyOnWaitlist {
			respondError(w, http.StatusConflict, "Already on waitlist")
			return
		}
		h.logger.Error("Failed to join waitlist: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to join waitlist")
		return
	}

	respondJSON(w, http.StatusCreated, entry)
}

// LeaveWaitlist handles DELETE /api/sessions/{session_id}/waitlist
func (h *Phase4Handler) LeaveWaitlist(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	if err := h.phase4Service.CancelWaitlistBySession(userID, sessionID); err != nil {
		if err == service.ErrWaitlistEntryNotFound {
			respondError(w, http.StatusNotFound, "Not on waitlist")
			return
		}
		h.logger.Error("Failed to leave waitlist: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to leave waitlist")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Removed from waitlist"})
}

// GetSessionWaitlist handles GET /api/sessions/{session_id}/waitlist
func (h *Phase4Handler) GetSessionWaitlist(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	entries, err := h.phase4Service.GetSessionWaitlist(sessionID)
	if err != nil {
		h.logger.Error("Failed to get session waitlist: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get session waitlist")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"waitlist": entries,
		"count":    len(entries),
	})
}

// GetUserWaitlistPosition handles GET /api/sessions/{session_id}/waitlist/position
func (h *Phase4Handler) GetUserWaitlistPosition(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	sessionIDStr := chi.URLParam(r, "session_id")
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	position, err := h.phase4Service.GetWaitlistPosition(sessionID, userID)
	if err != nil {
		if err == service.ErrWaitlistEntryNotFound {
			respondError(w, http.StatusNotFound, "Not on waitlist")
			return
		}
		h.logger.Error("Failed to get waitlist position: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get waitlist position")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"position": position,
	})
}

// GetUserWaitlistEntries handles GET /api/users/me/waitlist
func (h *Phase4Handler) GetUserWaitlistEntries(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	entries, err := h.phase4Service.GetUserWaitlistEntries(userID)
	if err != nil {
		h.logger.Error("Failed to get user waitlist entries: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user waitlist entries")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"waitlist_entries": entries,
		"count":            len(entries),
	})
}

// ============================================
// NOTIFICATION HANDLERS
// ============================================

// GetUserNotifications handles GET /api/users/me/notifications
func (h *Phase4Handler) GetUserNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	notifications, err := h.phase4Service.GetUserNotifications(userID, limit)
	if err != nil {
		h.logger.Error("Failed to get user notifications: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get user notifications")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"count":         len(notifications),
	})
}
