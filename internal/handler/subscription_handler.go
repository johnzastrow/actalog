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

// SubscriptionHandler handles subscription-related endpoints
type SubscriptionHandler struct {
	subscriptionService *service.SubscriptionService
	logger              *logger.Logger
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(subscriptionService *service.SubscriptionService, l *logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		logger:              l,
	}
}

// CreateUserSubscriptionRequest represents a request to create a user subscription
type CreateUserSubscriptionRequest struct {
	UserID           int64  `json:"user_id"`
	SubscriptionType string `json:"subscription_type"` // "free", "monthly", "annual"
	IsPermanentFree  bool   `json:"is_permanent_free"`
	Notes            string `json:"notes"`
}

// CreateOrganizationSubscriptionRequest represents a request to create an organization subscription
type CreateOrganizationSubscriptionRequest struct {
	OrganizationID   int64  `json:"organization_id"`
	SubscriptionType string `json:"subscription_type"`
	IsPermanentFree  bool   `json:"is_permanent_free"`
	Notes            string `json:"notes"`
}

// CancelSubscriptionRequest represents a request to cancel a subscription
type CancelSubscriptionRequest struct {
	Reason string `json:"reason"`
}

// GetMySubscriptionStatus returns the current user's subscription status
func (h *SubscriptionHandler) GetMySubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription status
	status, err := h.subscriptionService.GetUserSubscriptionStatus(userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_subscription_status outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to get subscription status")
		return
	}

	respondJSON(w, http.StatusOK, status)
}

// CreateUserSubscription creates a new user subscription (admin only)
func (h *SubscriptionHandler) CreateUserSubscription(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateUserSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	if req.SubscriptionType == "" {
		respondError(w, http.StatusBadRequest, "subscription_type is required")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=create_user_subscription admin_id=%d user_id=%d type=%s permanent_free=%v",
			adminUserID, req.UserID, req.SubscriptionType, req.IsPermanentFree)
	}

	// Create subscription
	sub, err := h.subscriptionService.CreateUserSubscription(
		adminUserID,
		req.UserID,
		req.SubscriptionType,
		req.IsPermanentFree,
		req.Notes,
	)
	if err != nil {
		switch err {
		case service.ErrCannotModifyOwnSubscription:
			respondError(w, http.StatusForbidden, "Admins cannot modify their own subscriptions")
		case service.ErrActiveSubscriptionExists:
			respondError(w, http.StatusConflict, "User already has an active subscription")
		default:
			if h.logger != nil {
				h.logger.Error("action=create_user_subscription outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to create subscription")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=create_user_subscription outcome=success admin_id=%d subscription_id=%d", adminUserID, sub.ID)
	}

	respondJSON(w, http.StatusCreated, sub)
}

// CreateOrganizationSubscription creates a new organization subscription (admin only)
func (h *SubscriptionHandler) CreateOrganizationSubscription(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req CreateOrganizationSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.OrganizationID == 0 {
		respondError(w, http.StatusBadRequest, "organization_id is required")
		return
	}
	if req.SubscriptionType == "" {
		respondError(w, http.StatusBadRequest, "subscription_type is required")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=create_org_subscription admin_id=%d org_id=%d type=%s permanent_free=%v",
			adminUserID, req.OrganizationID, req.SubscriptionType, req.IsPermanentFree)
	}

	// Create subscription
	sub, err := h.subscriptionService.CreateOrganizationSubscription(
		adminUserID,
		req.OrganizationID,
		req.SubscriptionType,
		req.IsPermanentFree,
		req.Notes,
	)
	if err != nil {
		switch err {
		case service.ErrActiveOrgSubscriptionExists:
			respondError(w, http.StatusConflict, "Organization already has an active subscription")
		default:
			if h.logger != nil {
				h.logger.Error("action=create_org_subscription outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to create subscription")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=create_org_subscription outcome=success admin_id=%d subscription_id=%d", adminUserID, sub.ID)
	}

	respondJSON(w, http.StatusCreated, sub)
}

// GetUserSubscriptions gets all subscriptions for a user (admin only)
func (h *SubscriptionHandler) GetUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL parameter
	userIDStr := chi.URLParam(r, "user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	// Get subscriptions
	subscriptions, err := h.subscriptionService.GetUserSubscriptions(userID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_user_subscriptions outcome=failure user_id=%d error=%v", userID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to get subscriptions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"count":         len(subscriptions),
	})
}

// GetOrganizationSubscriptions gets all subscriptions for an organization (admin only)
func (h *SubscriptionHandler) GetOrganizationSubscriptions(w http.ResponseWriter, r *http.Request) {
	// Get organization ID from URL parameter
	orgIDStr := chi.URLParam(r, "org_id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid org_id")
		return
	}

	// Get subscriptions
	subscriptions, err := h.subscriptionService.GetOrganizationSubscriptions(orgID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=get_org_subscriptions outcome=failure org_id=%d error=%v", orgID, err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to get subscriptions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"count":         len(subscriptions),
	})
}

// MarkUserSubscriptionAsPaidRequest represents a request to mark a subscription as paid
type MarkUserSubscriptionAsPaidRequest struct {
	PaymentDate  *string `json:"payment_date"`  // Optional: defaults to now
	DurationDays *int    `json:"duration_days"` // Optional: custom duration in days
}

// MarkUserSubscriptionAsPaid marks a user subscription as paid (admin only)
func (h *SubscriptionHandler) MarkUserSubscriptionAsPaid(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription ID from URL parameter
	subIDStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(subIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	// Parse request body (optional)
	var req MarkUserSubscriptionAsPaidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is invalid, use defaults
		req = MarkUserSubscriptionAsPaidRequest{}
	}

	if h.logger != nil {
		h.logger.Info("action=mark_user_subscription_paid admin_id=%d subscription_id=%d duration_days=%v",
			adminUserID, subID, req.DurationDays)
	}

	// Mark as paid
	err = h.subscriptionService.MarkUserSubscriptionAsPaid(adminUserID, subID, req.PaymentDate, req.DurationDays)
	if err != nil {
		switch err {
		case service.ErrSubscriptionNotFound:
			respondError(w, http.StatusNotFound, "Subscription not found")
		case service.ErrCannotModifyOwnSubscription:
			respondError(w, http.StatusForbidden, "Admins cannot modify their own subscriptions")
		case service.ErrCannotMarkFreeSubscriptionPaid:
			respondError(w, http.StatusBadRequest, "Cannot mark free subscription as paid")
		default:
			if h.logger != nil {
				h.logger.Error("action=mark_user_subscription_paid outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to mark subscription as paid")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=mark_user_subscription_paid outcome=success admin_id=%d subscription_id=%d", adminUserID, subID)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription marked as paid successfully",
	})
}

// MarkOrganizationSubscriptionAsPaid marks an organization subscription as paid (admin only)
func (h *SubscriptionHandler) MarkOrganizationSubscriptionAsPaid(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription ID from URL parameter
	subIDStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(subIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	// Parse request body (optional)
	var req MarkUserSubscriptionAsPaidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If body is invalid, use defaults
		req = MarkUserSubscriptionAsPaidRequest{}
	}

	if h.logger != nil {
		h.logger.Info("action=mark_org_subscription_paid admin_id=%d subscription_id=%d duration_days=%v",
			adminUserID, subID, req.DurationDays)
	}

	// Mark as paid
	err = h.subscriptionService.MarkOrganizationSubscriptionAsPaid(adminUserID, subID, req.PaymentDate, req.DurationDays)
	if err != nil {
		switch err {
		case service.ErrSubscriptionNotFound:
			respondError(w, http.StatusNotFound, "Subscription not found")
		case service.ErrCannotMarkFreeSubscriptionPaid:
			respondError(w, http.StatusBadRequest, "Cannot mark free subscription as paid")
		default:
			if h.logger != nil {
				h.logger.Error("action=mark_org_subscription_paid outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to mark subscription as paid")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=mark_org_subscription_paid outcome=success admin_id=%d subscription_id=%d", adminUserID, subID)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription marked as paid successfully",
	})
}

// CancelUserSubscription cancels a user subscription (admin only)
func (h *SubscriptionHandler) CancelUserSubscription(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription ID from URL parameter
	subIDStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(subIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var req CancelSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "Cancellation reason is required")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=cancel_user_subscription admin_id=%d subscription_id=%d reason=%s", adminUserID, subID, req.Reason)
	}

	// Cancel subscription
	err = h.subscriptionService.CancelUserSubscription(adminUserID, subID, req.Reason)
	if err != nil {
		switch err {
		case service.ErrSubscriptionNotFound:
			respondError(w, http.StatusNotFound, "Subscription not found")
		case service.ErrCannotModifyOwnSubscription:
			respondError(w, http.StatusForbidden, "Admins cannot modify their own subscriptions")
		default:
			if h.logger != nil {
				h.logger.Error("action=cancel_user_subscription outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to cancel subscription")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=cancel_user_subscription outcome=success admin_id=%d subscription_id=%d", adminUserID, subID)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription cancelled successfully",
	})
}

// SetUserSubscriptionPermanentRequest represents a request to set permanent status
type SetUserSubscriptionPermanentRequest struct {
	IsPermanent bool `json:"is_permanent"`
}

// SetUserSubscriptionPermanent sets the permanent free status of a user subscription (admin only)
func (h *SubscriptionHandler) SetUserSubscriptionPermanent(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription ID from URL parameter
	subIDStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(subIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	// Parse request body
	var req SetUserSubscriptionPermanentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=set_user_subscription_permanent admin_id=%d subscription_id=%d is_permanent=%v", adminUserID, subID, req.IsPermanent)
	}

	// Set permanent status
	err = h.subscriptionService.SetUserSubscriptionPermanent(adminUserID, subID, req.IsPermanent)
	if err != nil {
		switch err {
		case service.ErrSubscriptionNotFound:
			respondError(w, http.StatusNotFound, "Subscription not found")
		case service.ErrCannotModifyOwnSubscription:
			respondError(w, http.StatusForbidden, "Admins cannot modify their own subscriptions")
		default:
			if h.logger != nil {
				h.logger.Error("action=set_user_subscription_permanent outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to update subscription")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=set_user_subscription_permanent outcome=success admin_id=%d subscription_id=%d", adminUserID, subID)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription permanent status updated successfully",
	})
}

// CancelOrganizationSubscription cancels an organization subscription (admin only)
func (h *SubscriptionHandler) CancelOrganizationSubscription(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription ID from URL parameter
	subIDStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(subIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	var req CancelSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "Cancellation reason is required")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=cancel_org_subscription admin_id=%d subscription_id=%d reason=%s", adminUserID, subID, req.Reason)
	}

	// Cancel subscription
	err = h.subscriptionService.CancelOrganizationSubscription(adminUserID, subID, req.Reason)
	if err != nil {
		switch err {
		case service.ErrSubscriptionNotFound:
			respondError(w, http.StatusNotFound, "Subscription not found")
		default:
			if h.logger != nil {
				h.logger.Error("action=cancel_org_subscription outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to cancel subscription")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=cancel_org_subscription outcome=success admin_id=%d subscription_id=%d", adminUserID, subID)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription cancelled successfully",
	})
}

// SetOrganizationSubscriptionPermanent sets the permanent free status of an organization subscription (admin only)
func (h *SubscriptionHandler) SetOrganizationSubscriptionPermanent(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get subscription ID from URL parameter
	subIDStr := chi.URLParam(r, "id")
	subID, err := strconv.ParseInt(subIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid subscription ID")
		return
	}

	// Parse request body
	var req SetUserSubscriptionPermanentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if h.logger != nil {
		h.logger.Info("action=set_org_subscription_permanent admin_id=%d subscription_id=%d is_permanent=%v", adminUserID, subID, req.IsPermanent)
	}

	// Set permanent status
	err = h.subscriptionService.SetOrganizationSubscriptionPermanent(adminUserID, subID, req.IsPermanent)
	if err != nil {
		switch err {
		case service.ErrSubscriptionNotFound:
			respondError(w, http.StatusNotFound, "Subscription not found")
		default:
			if h.logger != nil {
				h.logger.Error("action=set_org_subscription_permanent outcome=failure admin_id=%d error=%v", adminUserID, err)
			}
			respondError(w, http.StatusInternalServerError, "Failed to update subscription")
		}
		return
	}

	if h.logger != nil {
		h.logger.Info("action=set_org_subscription_permanent outcome=success admin_id=%d subscription_id=%d", adminUserID, subID)
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription permanent status updated successfully",
	})
}

// ListAllUserSubscriptions lists all user subscriptions (admin only)
func (h *SubscriptionHandler) ListAllUserSubscriptions(w http.ResponseWriter, r *http.Request) {
	// Get all user subscriptions
	subscriptions, err := h.subscriptionService.ListAllUserSubscriptions()
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=list_all_user_subscriptions outcome=failure error=%v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to get subscriptions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"count":         len(subscriptions),
	})
}

// ListAllOrganizationSubscriptions lists all organization subscriptions (admin only)
func (h *SubscriptionHandler) ListAllOrganizationSubscriptions(w http.ResponseWriter, r *http.Request) {
	// Get all organization subscriptions
	subscriptions, err := h.subscriptionService.ListAllOrganizationSubscriptions()
	if err != nil {
		if h.logger != nil {
			h.logger.Error("action=list_all_org_subscriptions outcome=failure error=%v", err)
		}
		respondError(w, http.StatusInternalServerError, "Failed to get subscriptions")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"count":         len(subscriptions),
	})
}
