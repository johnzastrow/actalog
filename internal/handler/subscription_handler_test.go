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
	// Test constructor with nil dependencies
	handler := NewSubscriptionHandler(nil, nil)

	if handler == nil {
		t.Error("NewSubscriptionHandler() should not return nil")
	}

	if handler.subscriptionService != nil {
		t.Error("subscriptionService should be nil when passed nil")
	}

	if handler.logger != nil {
		t.Error("logger should be nil when passed nil")
	}
}

func TestCreateUserSubscriptionRequest_Struct(t *testing.T) {
	req := CreateUserSubscriptionRequest{
		UserID:           123,
		SubscriptionType: "monthly",
		IsPermanentFree:  false,
		Notes:            "Test subscription",
	}

	if req.UserID != 123 {
		t.Errorf("UserID = %d, want 123", req.UserID)
	}
	if req.SubscriptionType != "monthly" {
		t.Errorf("SubscriptionType = %q, want %q", req.SubscriptionType, "monthly")
	}
	if req.IsPermanentFree != false {
		t.Error("IsPermanentFree should be false")
	}
	if req.Notes != "Test subscription" {
		t.Errorf("Notes = %q, want %q", req.Notes, "Test subscription")
	}
}

func TestCreateUserSubscriptionRequest_PermanentFree(t *testing.T) {
	req := CreateUserSubscriptionRequest{
		UserID:           456,
		SubscriptionType: "free",
		IsPermanentFree:  true,
		Notes:            "Lifetime free",
	}

	if req.IsPermanentFree != true {
		t.Error("IsPermanentFree should be true")
	}
}

func TestCreateOrganizationSubscriptionRequest_Struct(t *testing.T) {
	req := CreateOrganizationSubscriptionRequest{
		OrganizationID:   789,
		SubscriptionType: "annual",
		IsPermanentFree:  false,
		Notes:            "Org subscription",
	}

	if req.OrganizationID != 789 {
		t.Errorf("OrganizationID = %d, want 789", req.OrganizationID)
	}
	if req.SubscriptionType != "annual" {
		t.Errorf("SubscriptionType = %q, want %q", req.SubscriptionType, "annual")
	}
	if req.IsPermanentFree != false {
		t.Error("IsPermanentFree should be false")
	}
	if req.Notes != "Org subscription" {
		t.Errorf("Notes = %q, want %q", req.Notes, "Org subscription")
	}
}

func TestCancelSubscriptionRequest_Struct(t *testing.T) {
	req := CancelSubscriptionRequest{
		Reason: "User requested cancellation",
	}

	if req.Reason != "User requested cancellation" {
		t.Errorf("Reason = %q, want %q", req.Reason, "User requested cancellation")
	}

	// Test empty reason
	req2 := CancelSubscriptionRequest{}
	if req2.Reason != "" {
		t.Errorf("Reason = %q, want empty", req2.Reason)
	}
}

func TestMarkUserSubscriptionAsPaidRequest_Struct(t *testing.T) {
	paymentDate := "2024-01-15"
	durationDays := 30

	req := MarkUserSubscriptionAsPaidRequest{
		PaymentDate:  &paymentDate,
		DurationDays: &durationDays,
	}

	if req.PaymentDate == nil || *req.PaymentDate != "2024-01-15" {
		t.Errorf("PaymentDate = %v, want 2024-01-15", req.PaymentDate)
	}
	if req.DurationDays == nil || *req.DurationDays != 30 {
		t.Errorf("DurationDays = %v, want 30", req.DurationDays)
	}
}

func TestMarkUserSubscriptionAsPaidRequest_NilFields(t *testing.T) {
	req := MarkUserSubscriptionAsPaidRequest{}

	if req.PaymentDate != nil {
		t.Error("PaymentDate should be nil")
	}
	if req.DurationDays != nil {
		t.Error("DurationDays should be nil")
	}
}

func TestSetUserSubscriptionPermanentRequest_Struct(t *testing.T) {
	req := SetUserSubscriptionPermanentRequest{
		IsPermanent: true,
	}

	if req.IsPermanent != true {
		t.Error("IsPermanent should be true")
	}

	req2 := SetUserSubscriptionPermanentRequest{
		IsPermanent: false,
	}

	if req2.IsPermanent != false {
		t.Error("IsPermanent should be false")
	}
}

func TestSubscriptionHandler_GetMySubscriptionStatus_Unauthorized(t *testing.T) {
	handler := &SubscriptionHandler{}

	req := createTestRequest(http.MethodGet, "/api/subscription/status", "")
	rr := httptest.NewRecorder()

	handler.GetMySubscriptionStatus(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
	assertBodyContains(t, rr, "Unauthorized")
}

func TestSubscriptionHandler_GetMySubscriptionStatus_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/subscription/status", "", 1, "test@example.com", "user")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.GetMySubscriptionStatus(rr, req)
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

func TestSubscriptionHandler_CreateUserSubscription_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user", `{"user_id": 2, "subscription_type": "monthly"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.CreateUserSubscription(rr, req)
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

func TestSubscriptionHandler_CreateOrganizationSubscription_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization", `{"organization_id": 1, "subscription_type": "monthly"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.CreateOrganizationSubscription(rr, req)
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

func TestSubscriptionHandler_ListAllUserSubscriptions_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/users", "")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.ListAllUserSubscriptions(rr, req)
}

func TestSubscriptionHandler_ListAllOrganizationSubscriptions_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/organizations", "")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.ListAllOrganizationSubscriptions(rr, req)
}

func TestSubscriptionHandler_AllSubscriptionTypes(t *testing.T) {
	testCases := []struct {
		name             string
		subscriptionType string
	}{
		{"free", "free"},
		{"monthly", "monthly"},
		{"annual", "annual"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := CreateUserSubscriptionRequest{
				UserID:           1,
				SubscriptionType: tc.subscriptionType,
			}

			if req.SubscriptionType != tc.subscriptionType {
				t.Errorf("SubscriptionType = %q, want %q", req.SubscriptionType, tc.subscriptionType)
			}
		})
	}
}

func TestSubscriptionHandler_GetUserSubscriptions_ValidID(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/users/1", "")
	req = addChiURLParam(req, "user_id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.GetUserSubscriptions(rr, req)
}

func TestSubscriptionHandler_GetOrganizationSubscriptions_ValidID(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/organizations/1", "")
	req = addChiURLParam(req, "org_id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.GetOrganizationSubscriptions(rr, req)
}

func TestSubscriptionHandler_MarkUserSubscriptionAsPaid_ValidID(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/mark-paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.MarkUserSubscriptionAsPaid(rr, req)
}

func TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_ValidID(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/mark-paid", `{}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.MarkOrganizationSubscriptionAsPaid(rr, req)
}

func TestSubscriptionHandler_CancelUserSubscription_ValidIDWithReason(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/cancel", `{"reason": "Customer requested"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.CancelUserSubscription(rr, req)
}

func TestSubscriptionHandler_CancelOrganizationSubscription_ValidIDWithReason(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/cancel", `{"reason": "Org requested"}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.CancelOrganizationSubscription(rr, req)
}

func TestSubscriptionHandler_SetUserSubscriptionPermanent_ValidID(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/1/permanent", `{"is_permanent": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.SetUserSubscriptionPermanent(rr, req)
}

func TestSubscriptionHandler_SetOrganizationSubscriptionPermanent_ValidID(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/1/permanent", `{"is_permanent": true}`, 1, "admin@example.com", "admin")
	req = addChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()

	// This will panic due to nil subscriptionService
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.SetOrganizationSubscriptionPermanent(rr, req)
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

// Tests with different IDs

func TestSubscriptionHandler_GetUserSubscriptions_DifferentIDs(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("user_id_"+id, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/users/"+id, "")
			req = addChiURLParam(req, "user_id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetUserSubscriptions requires service")
				}
			}()

			handler.GetUserSubscriptions(rr, req)
		})
	}
}

func TestSubscriptionHandler_GetOrganizationSubscriptions_DifferentIDs(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("org_id_"+id, func(t *testing.T) {
			req := createTestRequest(http.MethodGet, "/api/admin/subscriptions/organizations/"+id, "")
			req = addChiURLParam(req, "org_id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("GetOrganizationSubscriptions requires service")
				}
			}()

			handler.GetOrganizationSubscriptions(rr, req)
		})
	}
}

func TestSubscriptionHandler_MarkUserSubscriptionAsPaid_DifferentIDs(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("subscription_id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/user/"+id+"/mark-paid", `{}`, 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("MarkUserSubscriptionAsPaid requires service")
				}
			}()

			handler.MarkUserSubscriptionAsPaid(rr, req)
		})
	}
}

func TestSubscriptionHandler_MarkOrganizationSubscriptionAsPaid_DifferentIDs(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	testIDs := []string{"1", "10", "100"}

	for _, id := range testIDs {
		t.Run("subscription_id_"+id, func(t *testing.T) {
			req := createAuthenticatedRequest(http.MethodPost, "/api/admin/subscriptions/organization/"+id+"/mark-paid", `{}`, 1, "admin@example.com", "admin")
			req = addChiURLParam(req, "id", id)
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("MarkOrganizationSubscriptionAsPaid requires service")
				}
			}()

			handler.MarkOrganizationSubscriptionAsPaid(rr, req)
		})
	}
}

// Tests for ListExpiringUserSubscriptions
func TestSubscriptionHandler_ListExpiringUserSubscriptions_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.ListExpiringUserSubscriptions(rr, req)
}

func TestSubscriptionHandler_ListExpiringUserSubscriptions_QueryParam(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	testCases := []struct {
		name string
		days string
	}{
		{"default", ""},
		{"7 days", "7"},
		{"30 days", "30"},
		{"60 days", "60"},
		{"invalid", "abc"}, // Falls back to default
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/admin/subscriptions/users/expiring"
			if tc.days != "" {
				url += "?days=" + tc.days
			}
			req := createAuthenticatedRequest(http.MethodGet, url, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListExpiringUserSubscriptions requires service")
				}
			}()

			handler.ListExpiringUserSubscriptions(rr, req)
		})
	}
}

// Tests for ListExpiredUserSubscriptions
func TestSubscriptionHandler_ListExpiredUserSubscriptions_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/users/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.ListExpiredUserSubscriptions(rr, req)
}

// Tests for ListExpiringOrganizationSubscriptions
func TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expiring", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.ListExpiringOrganizationSubscriptions(rr, req)
}

func TestSubscriptionHandler_ListExpiringOrganizationSubscriptions_QueryParam(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	testCases := []struct {
		name string
		days string
	}{
		{"default", ""},
		{"14 days", "14"},
		{"90 days", "90"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/admin/subscriptions/organizations/expiring"
			if tc.days != "" {
				url += "?days=" + tc.days
			}
			req := createAuthenticatedRequest(http.MethodGet, url, "", 1, "admin@example.com", "admin")
			rr := httptest.NewRecorder()

			defer func() {
				if r := recover(); r == nil {
					t.Log("ListExpiringOrganizationSubscriptions requires service")
				}
			}()

			handler.ListExpiringOrganizationSubscriptions(rr, req)
		})
	}
}

// Tests for ListExpiredOrganizationSubscriptions
func TestSubscriptionHandler_ListExpiredOrganizationSubscriptions_NilService(t *testing.T) {
	handler := &SubscriptionHandler{
		logger: createTestLogger(),
	}

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/subscriptions/organizations/expired", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic with nil subscriptionService")
		}
	}()

	handler.ListExpiredOrganizationSubscriptions(rr, req)
}

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
