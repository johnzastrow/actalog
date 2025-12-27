package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubscriptionHandler_GetMySubscriptionStatus_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodGet, "/api/subscription/status", "")
	rr := httptest.NewRecorder()

	handler.GetMySubscriptionStatus(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_CreateUserSubscription_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/user", `{"user_id": 1, "subscription_type": "monthly"}`)
	rr := httptest.NewRecorder()

	handler.CreateUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_CreateUserSubscription_InvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestSubscriptionHandler_CreateUserSubscription_MissingUserID(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user", `{"subscription_type": "monthly"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "user_id is required")
}

func TestSubscriptionHandler_CreateUserSubscription_MissingType(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user", `{"user_id": 2}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "subscription_type is required")
}

func TestSubscriptionHandler_CreateOrganizationSubscription_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/organization", `{"organization_id": 1, "subscription_type": "monthly"}`)
	rr := httptest.NewRecorder()

	handler.CreateOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_CreateOrganizationSubscription_InvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestSubscriptionHandler_CreateOrganizationSubscription_MissingOrgID(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization", `{"subscription_type": "monthly"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "organization_id is required")
}
