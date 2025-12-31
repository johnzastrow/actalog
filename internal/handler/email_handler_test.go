package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/email"
)

func TestNewEmailHandler(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(nil, emailLogService, nil)

	if handler == nil {
		t.Error("NewEmailHandler returned nil")
	}
}

func TestEmailHandler_GetEmailConfig_Unauthorized(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(nil, emailLogService, nil)

	req := createTestRequest(http.MethodGet, "/api/admin/email/config", "")
	rr := httptest.NewRecorder()

	handler.GetEmailConfig(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestEmailHandler_GetEmailConfig_NilService(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(nil, emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email/config", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetEmailConfig(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, `"enabled":false`)
}

func TestEmailHandler_GetEmailConfig_WithService(t *testing.T) {
	mockEmailService := &email.Service{}
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(mockEmailService, emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodGet, "/api/admin/email/config", "", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.GetEmailConfig(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
}

func TestEmailHandler_SendTestEmail_Unauthorized(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(nil, emailLogService, nil)

	req := createTestRequest(http.MethodPost, "/api/admin/email/test", `{"recipient_email":"test@example.com"}`)
	rr := httptest.NewRecorder()

	handler.SendTestEmail(rr, req)

	assertStatusCode(t, rr, http.StatusUnauthorized)
}

func TestEmailHandler_SendTestEmail_NilService(t *testing.T) {
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(nil, emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/email/test", `{"recipient_email":"test@example.com"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SendTestEmail(rr, req)

	assertStatusCode(t, rr, http.StatusOK)
	assertBodyContains(t, rr, "Email service is not configured")
}

func TestEmailHandler_SendTestEmail_InvalidBody(t *testing.T) {
	mockEmailService := &email.Service{}
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(mockEmailService, emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/email/test", "invalid json", 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SendTestEmail(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
}

func TestEmailHandler_SendTestEmail_EmptyEmail(t *testing.T) {
	mockEmailService := &email.Service{}
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(mockEmailService, emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/email/test", `{"recipient_email":""}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SendTestEmail(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Recipient email is required")
}

func TestEmailHandler_SendTestEmail_InvalidEmailFormat(t *testing.T) {
	mockEmailService := &email.Service{}
	emailLogService := service.NewEmailLogService(NewMockEmailLogRepository())
	handler := NewEmailHandler(mockEmailService, emailLogService, nil)

	req := createAuthenticatedRequest(http.MethodPost, "/api/admin/email/test", `{"recipient_email":"not-an-email"}`, 1, "admin@example.com", "admin")
	rr := httptest.NewRecorder()

	handler.SendTestEmail(rr, req)

	assertStatusCode(t, rr, http.StatusBadRequest)
	assertBodyContains(t, rr, "Invalid email address format")
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected bool
	}{
		{"test@example.com", true},
		{"user.name@domain.org", true},
		{"user+tag@example.com", true},
		{"invalid", false},
		{"@example.com", false},
		{"test@", false},
		{"test@.com", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}
