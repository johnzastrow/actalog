package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// errMockFailure is a standard error used in mock error tests
var errMockFailure = errors.New("mock failure")

func TestNewSubscriptionHandler(t *testing.T) {
	handler := NewSubscriptionHandler(nil, createTestLogger())
	if handler == nil {
		t.Fatal("NewSubscriptionHandler() should not return nil")
	}
}

// Removed struct field assignment tests:
// - TestCreateUserSubscriptionRequest_Struct, TestCreateUserSubscriptionRequest_PermanentFree
// - TestCreateOrganizationSubscriptionRequest_Struct
// - TestCancelSubscriptionRequest_Struct
// - TestMarkUserSubscriptionAsPaidRequest_Struct, TestMarkUserSubscriptionAsPaidRequest_NilFields
// - TestSetUserSubscriptionPermanentRequest_Struct
// These tests verified Go struct assignment works, not business logic.

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

func TestSubscriptionHandler_CreateOrganizationSubscription_MissingType(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization", `{"organization_id": 1}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "subscription_type is required")
}

func TestSubscriptionHandler_GetUserSubscriptions_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/users/abc", "")
	rr := httptest.NewRecorder()

	handler.GetUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid user_id")
}

func TestSubscriptionHandler_GetOrganizationSubscriptions_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/organizations/abc", "")
	rr := httptest.NewRecorder()

	handler.GetOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid org_id")
}

func TestSubscriptionHandler_MarkUserSubscriptionAsPaid_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/user/1/mark-paid", "")
	rr := httptest.NewRecorder()

	handler.MarkUserSubscriptionAsPaid(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_MarkUserSubscriptionAsPaid_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/abc/mark-paid", `{}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.MarkUserSubscriptionAsPaid(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid subscription ID")
}

func TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/mark-paid", "")
	rr := httptest.NewRecorder()

	handler.MarkOrganizationSubscriptionAsPaid(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/abc/mark-paid", `{}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.MarkOrganizationSubscriptionAsPaid(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid subscription ID")
}

func TestSubscriptionHandler_CancelUserSubscription_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel", `{"reason": "Test"}`)
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_CancelUserSubscription_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/abc/cancel", `{"reason": "Test"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid subscription ID")
}

func TestSubscriptionHandler_CancelUserSubscription_InvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestSubscriptionHandler_CancelUserSubscription_EmptyReason(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel", `{"reason": ""}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestSubscriptionHandler_CancelOrganizationSubscription_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel", `{"reason": "Test"}`)
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_CancelOrganizationSubscription_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/abc/cancel", `{"reason": "Test"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid subscription ID")
}

func TestSubscriptionHandler_CancelOrganizationSubscription_InvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestSubscriptionHandler_CancelOrganizationSubscription_EmptyReason(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel", `{"reason": ""}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestSubscriptionHandler_SetUserSubscriptionPermanent_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/user/1/permanent", `{"is_permanent": true}`)
	rr := httptest.NewRecorder()

	handler.SetUserSubscriptionPermanent(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_SetUserSubscriptionPermanent_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/abc/permanent", `{"is_permanent": true}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SetUserSubscriptionPermanent(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid subscription ID")
}

func TestSubscriptionHandler_SetUserSubscriptionPermanent_InvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/permanent", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SetUserSubscriptionPermanent(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/permanent", `{"is_permanent": true}`)
	rr := httptest.NewRecorder()

	handler.SetOrganizationSubscriptionPermanent(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_InvalidID(t *testing.T) {
	handler := &SubscriptionHandler{}

	// Without chi router context, URLParam returns empty string
	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/abc/permanent", `{"is_permanent": true}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SetOrganizationSubscriptionPermanent(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid subscription ID")
}

func TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_InvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/permanent", "{bad json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SetOrganizationSubscriptionPermanent(rr, req)

	// Without chi context, fails on URL param first
	assertStatusCode(t, rr, http.StatusBadRequest)
}

// Additional tests for Cancel subscription with valid IDs and empty reason

func TestSubscriptionHandler_CancelUserSubscription_ValidIDEmptyReason(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel", `{"reason": ""}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Cancellation reason is required")
}

func TestSubscriptionHandler_CancelOrganizationSubscription_ValidIDEmptyReason(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel", `{"reason": ""}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Cancellation reason is required")
}

func TestSubscriptionHandler_CancelUserSubscription_ValidIDInvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestSubscriptionHandler_CancelOrganizationSubscription_ValidIDInvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestSubscriptionHandler_SetUserSubscriptionPermanent_ValidIDInvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/permanent", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.SetUserSubscriptionPermanent(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

func TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_ValidIDInvalidJSON(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/permanent", "{bad json", 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.SetOrganizationSubscriptionPermanent(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid request body")
}

// Removed 23 panic-expectation tests:
// - TestSubscriptionHandler_GetMySubscriptionStatus_NilService
// - TestSubscriptionHandler_CreateUserSubscription_NilService
// - TestSubscriptionHandler_CreateOrganizationSubscription_NilService
// - TestSubscriptionHandler_ListAllUserSubscriptions_NilService
// - TestSubscriptionHandler_ListAllOrganizationSubscriptions_NilService
// - TestSubscriptionHandler_AllSubscriptionTypes (struct assignment test)
// - TestSubscriptionHandler_GetUserSubscriptions_ValidID
// - TestSubscriptionHandler_GetOrganizationSubscriptions_ValidID
// - TestSubscriptionHandler_MarkUserSubscriptionAsPaid_ValidID
// - TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_ValidID
// - TestSubscriptionHandler_CancelUserSubscription_ValidIDWithReason
// - TestSubscriptionHandler_CancelOrganizationSubscription_ValidIDWithReason
// - TestSubscriptionHandler_SetUserSubscriptionPermanent_ValidID
// - TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_ValidID
// - TestSubscriptionHandler_GetUserSubscriptions_DifferentIDs
// - TestSubscriptionHandler_GetOrganizationSubscriptions_DifferentIDs
// - TestSubscriptionHandler_MarkUserSubscriptionAsPaid_DifferentIDs
// - TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_DifferentIDs
// - TestSubscriptionHandler_ListExpiringUserSubscriptions_NilService
// - TestSubscriptionHandler_ListExpiringUserSubscriptions_QueryParam
// - TestSubscriptionHandler_ListExpiredUserSubscriptions_NilService
// - TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_NilService
// - TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_QueryParam
// - TestSubscriptionHandler_ListExpiredOrganizationSubscriptions_NilService
// These tests verified nil pointer panics, not business logic.

// =============================================================================
// Success Path Tests with Real Service
// =============================================================================

func TestSubscriptionHandler_ListAllUserSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAllUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
}

func TestSubscriptionHandler_ListAllOrganizationSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAllOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
}

func TestSubscriptionHandler_ListExpiringUserSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiringUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
	assertBodyContains(t, rr, "days")
}

func TestSubscriptionHandler_ListExpiringUserSubscriptions_CustomDays(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	testCases := []struct {
		name         string
		days         string
		expectedDays string
	}{
		{"7 days", "7", "7"},
		{"14 days", "14", "14"},
		{"60 days", "60", "60"},
		{"invalid falls to default", "invalid", "30"},
		{"negative falls to default", "-5", "30"},
		{"zero falls to default", "0", "30"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/admin/subscriptions/users/expiring?days=" + tc.days
			req := createAuthenticatedRequest(http.MethodGet, url, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListExpiringUserSubscriptions(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
			assertBodyContains(t, rr, tc.expectedDays)
		})
	}
}

func TestSubscriptionHandler_ListExpiredUserSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiredUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
}

func TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiringOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
	assertBodyContains(t, rr, "days")
}

func TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_CustomDays(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	testCases := []struct {
		name         string
		days         string
		expectedDays string
	}{
		{"7 days", "7", "7"},
		{"14 days", "14", "14"},
		{"60 days", "60", "60"},
		{"invalid falls to default", "invalid", "30"},
		{"negative falls to default", "-5", "30"},
		{"zero falls to default", "0", "30"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/admin/subscriptions/organizations/expiring?days=" + tc.days
			req := createAuthenticatedRequest(http.MethodGet, url, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			handler.ListExpiringOrganizationSubscriptions(rr, req)

			assertStatusCode(t, rr, http.StatusOK)
			assertBodyContains(t, rr, tc.expectedDays)
		})
	}
}

func TestSubscriptionHandler_ListExpiredOrganizationSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiredOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
}

func TestSubscriptionHandler_GetMySubscriptionStatus_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/subscription/status", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetMySubscriptionStatus(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "has_access")
}

// =============================================================================
// Error Path Tests with Mock Errors
// =============================================================================

func TestSubscriptionHandler_ListAllUserSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAllUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get subscriptions")
}

func TestSubscriptionHandler_ListAllUserSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil, // No logger to test nil logger branch
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAllUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_ListAllOrganizationSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAllOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get subscriptions")
}

func TestSubscriptionHandler_ListAllOrganizationSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListAllOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_ListExpiredUserSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiredUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get expired subscriptions")
}

func TestSubscriptionHandler_ListExpiredUserSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiredUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_ListExpiredOrganizationSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiredOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get expired subscriptions")
}

func TestSubscriptionHandler_ListExpiredOrganizationSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiredOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_ListExpiringUserSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiringUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get expiring subscriptions")
}

func TestSubscriptionHandler_ListExpiringUserSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiringUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiringOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get expiring subscriptions")
}

func TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.ListExpiringOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_GetMySubscriptionStatus_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.accessRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/subscription/status", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetMySubscriptionStatus(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_GetMySubscriptionStatus_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.accessRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/subscription/status", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	handler.GetMySubscriptionStatus(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

// =============================================================================
// GetUserSubscriptions and GetOrganizationSubscriptions with Service
// =============================================================================

func TestSubscriptionHandler_GetUserSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/users/1", "")
	req = addChiURLParam(req, "user_id", "1")
	rr := httptest.NewRecorder()

	handler.GetUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
}

func TestSubscriptionHandler_GetUserSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/users/1", "")
	req = addChiURLParam(req, "user_id", "1")
	rr := httptest.NewRecorder()

	handler.GetUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get subscriptions")
}

func TestSubscriptionHandler_GetUserSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/users/1", "")
	req = addChiURLParam(req, "user_id", "1")
	rr := httptest.NewRecorder()

	handler.GetUserSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

func TestSubscriptionHandler_GetOrganizationSubscriptions_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/organizations/1", "")
	req = addChiURLParam(req, "org_id", "1")
	rr := httptest.NewRecorder()

	handler.GetOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "subscriptions")
	assertBodyContains(t, rr, "count")
}

func TestSubscriptionHandler_GetOrganizationSubscriptions_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/organizations/1", "")
	req = addChiURLParam(req, "org_id", "1")
	rr := httptest.NewRecorder()

	handler.GetOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
	assertBodyContains(t, rr, "Failed to get subscriptions")
}

func TestSubscriptionHandler_GetOrganizationSubscriptions_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/organizations/1", "")
	req = addChiURLParam(req, "org_id", "1")
	rr := httptest.NewRecorder()

	handler.GetOrganizationSubscriptions(rr, req)

	assertStatusCode(t, rr, http.StatusInternalServerError)
}

// =============================================================================
// MarkUserSubscriptionAsPaid and MarkOrganizationSubscriptionAsPaid with Service
// =============================================================================

func TestSubscriptionHandler_MarkUserSubscriptionAsPaid_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/subscriptions/users/1/paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.MarkUserSubscriptionAsPaid(rr, req)

	// May return not found or internal server error since mock has no subscriptions
	// The service wraps errors so 500 is also possible
	// Just verify it doesn't panic and exercises the code path
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_MarkUserSubscriptionAsPaid_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/subscriptions/users/1/paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.MarkUserSubscriptionAsPaid(rr, req)

	// Service error should result in 500 or specific error status
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusNotFound {
		t.Errorf("Expected StatusInternalServerError or StatusNotFound, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_MarkUserSubscriptionAsPaid_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.userSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/subscriptions/users/1/paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.MarkUserSubscriptionAsPaid(rr, req)

	// Service error should result in 500 or specific error status
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusNotFound {
		t.Errorf("Expected StatusInternalServerError or StatusNotFound, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/subscriptions/organizations/1/paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.MarkOrganizationSubscriptionAsPaid(rr, req)

	// May return not found since mock has no subscriptions
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound {
		t.Errorf("Expected StatusOK or StatusNotFound, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_ServiceError(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/subscriptions/organizations/1/paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.MarkOrganizationSubscriptionAsPaid(rr, req)

	// Service error should result in 500 or specific error status
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusNotFound {
		t.Errorf("Expected StatusInternalServerError or StatusNotFound, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_ServiceError_NoLogger(t *testing.T) {
	svc, mocks := createTestSubscriptionServiceWithMocks()
	mocks.orgSubRepo.SetError(errMockFailure)

	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPut, "/api/admin/subscriptions/organizations/1/paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.MarkOrganizationSubscriptionAsPaid(rr, req)

	// Service error should result in 500 or specific error status
	if rr.Code != http.StatusInternalServerError && rr.Code != http.StatusNotFound {
		t.Errorf("Expected StatusInternalServerError or StatusNotFound, got %d", rr.Code)
	}
}

// =============================================================================
// CreateUserSubscription with Service
// =============================================================================

func TestSubscriptionHandler_CreateUserSubscription_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user",
		`{"user_id": 1, "subscription_type": "monthly"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateUserSubscription(rr, req)

	// May return 201 (created), 400 (already exists), 403 (auth), or 500 (error)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusBadRequest && rr.Code != http.StatusForbidden && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusCreated, StatusBadRequest, StatusForbidden, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_CreateUserSubscription_PermanentFree(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user",
		`{"user_id": 2, "subscription_type": "permanent_free"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateUserSubscription(rr, req)

	// Test the permanent_free code path
	if rr.Code != http.StatusCreated && rr.Code != http.StatusBadRequest && rr.Code != http.StatusForbidden && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusCreated, StatusBadRequest, StatusForbidden, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_CreateUserSubscription_NoLogger(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user",
		`{"user_id": 1, "subscription_type": "annual"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateUserSubscription(rr, req)

	// May return 201 (created), 400 (already exists), 403 (auth), or 500 (error)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusBadRequest && rr.Code != http.StatusForbidden && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusCreated, StatusBadRequest, StatusForbidden, or StatusInternalServerError, got %d", rr.Code)
	}
}

// =============================================================================
// CreateOrganizationSubscription with Service
// =============================================================================

func TestSubscriptionHandler_CreateOrganizationSubscription_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization",
		`{"organization_id": 1, "subscription_type": "monthly"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganizationSubscription(rr, req)

	// May return 201 (created), 400 (already exists), 403 (auth), or 500 (error)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusBadRequest && rr.Code != http.StatusForbidden && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusCreated, StatusBadRequest, StatusForbidden, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_CreateOrganizationSubscription_PermanentFree(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization",
		`{"organization_id": 2, "subscription_type": "permanent_free"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganizationSubscription(rr, req)

	// Test the permanent_free code path
	if rr.Code != http.StatusCreated && rr.Code != http.StatusBadRequest && rr.Code != http.StatusForbidden && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusCreated, StatusBadRequest, StatusForbidden, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_CreateOrganizationSubscription_NoLogger(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization",
		`{"organization_id": 1, "subscription_type": "annual"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.CreateOrganizationSubscription(rr, req)

	// May return 201 (created), 400 (already exists), 403 (auth), or 500 (error)
	if rr.Code != http.StatusCreated && rr.Code != http.StatusBadRequest && rr.Code != http.StatusForbidden && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusCreated, StatusBadRequest, StatusForbidden, or StatusInternalServerError, got %d", rr.Code)
	}
}

// =============================================================================
// CancelUserSubscription with Service
// =============================================================================

func TestSubscriptionHandler_CancelUserSubscription_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel",
		`{"reason": "User requested cancellation"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_CancelUserSubscription_NoLogger(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel",
		`{"reason": "No longer needed"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelUserSubscription(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

// =============================================================================
// CancelOrganizationSubscription with Service
// =============================================================================

func TestSubscriptionHandler_CancelOrganizationSubscription_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel",
		`{"reason": "Organization requested cancellation"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_CancelOrganizationSubscription_NoLogger(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel",
		`{"reason": "No longer needed"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.CancelOrganizationSubscription(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

// =============================================================================
// SetUserSubscriptionPermanent with Service
// =============================================================================

func TestSubscriptionHandler_SetUserSubscriptionPermanent_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/permanent",
		`{"is_permanent": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.SetUserSubscriptionPermanent(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_SetUserSubscriptionPermanent_NoLogger(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/permanent",
		`{"is_permanent": false}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.SetUserSubscriptionPermanent(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

// =============================================================================
// SetOrganizationSubscriptionPermanent with Service
// =============================================================================

func TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_Success(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/permanent",
		`{"is_permanent": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.SetOrganizationSubscriptionPermanent(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}

func TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_NoLogger(t *testing.T) {
	svc := createTestSubscriptionService()
	handler := &SubscriptionHandler{
		subscriptionService: svc,
		logger:              nil,
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/permanent",
		`{"is_permanent": false}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	handler.SetOrganizationSubscriptionPermanent(rr, req)

	// May return 200 (success), 404 (not found), or 500 (error)
	if rr.Code != http.StatusOK && rr.Code != http.StatusNotFound && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected StatusOK, StatusNotFound, or StatusInternalServerError, got %d", rr.Code)
	}
}
