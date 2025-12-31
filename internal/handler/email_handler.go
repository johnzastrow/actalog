package handler

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/johnzastrow/actalog/internal/service"
	"github.com/johnzastrow/actalog/pkg/email"
	"github.com/johnzastrow/actalog/pkg/logger"
	"github.com/johnzastrow/actalog/pkg/middleware"
)

// EmailHandler handles email-related endpoints
type EmailHandler struct {
	emailService    *email.Service
	emailLogService *service.EmailLogService
	logger          *logger.Logger
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService *email.Service, emailLogService *service.EmailLogService, logger *logger.Logger) *EmailHandler {
	return &EmailHandler{
		emailService:    emailService,
		emailLogService: emailLogService,
		logger:          logger,
	}
}

// TestEmailRequest is the request body for test email
type TestEmailRequest struct {
	RecipientEmail string `json:"recipient_email"`
}

// TestEmailResponse includes detailed debug info
type TestEmailResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	DebugInfo *SMTPDebugResponse `json:"debug_info"`
}

// SMTPDebugResponse is the response format for SMTP debug info
type SMTPDebugResponse struct {
	TotalDuration   string   `json:"total_duration"`
	ConnectionHost  string   `json:"connection_host"`
	ConnectionPort  int      `json:"connection_port"`
	ConnectionTLS   string   `json:"connection_tls"`
	ConnectionTime  string   `json:"connection_time"`
	ConnectionError string   `json:"connection_error,omitempty"`
	AuthMethod      string   `json:"auth_method"`
	AuthTime        string   `json:"auth_time"`
	AuthError       string   `json:"auth_error,omitempty"`
	SendTime        string   `json:"send_time"`
	SendError       string   `json:"send_error,omitempty"`
	SMTPResponses   []string `json:"smtp_responses"`
	FinalError      string   `json:"final_error,omitempty"`
}

// EmailConfigResponse is the response format for email configuration
type EmailConfigResponse struct {
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    int    `json:"smtp_port"`
	SMTPUser    string `json:"smtp_user"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
	Enabled     bool   `json:"enabled"`
	TLSMode     string `json:"tls_mode"`
}

// GetEmailConfig handles GET /api/admin/email/config
// Returns the current email configuration (with masked password)
func (h *EmailHandler) GetEmailConfig(w http.ResponseWriter, r *http.Request) {
	// Verify admin authorization
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if email service is available
	if h.emailService == nil {
		response := EmailConfigResponse{
			Enabled: false,
		}
		respondJSON(w, http.StatusOK, response)
		return
	}

	config := h.emailService.GetConfig()

	response := EmailConfigResponse{
		SMTPHost:    config.SMTPHost,
		SMTPPort:    config.SMTPPort,
		SMTPUser:    config.SMTPUser,
		FromAddress: config.FromAddress,
		FromName:    config.FromName,
		Enabled:     config.Enabled,
		TLSMode:     config.TLSMode,
	}

	respondJSON(w, http.StatusOK, response)
}

// SendTestEmail handles POST /api/admin/email/test
// Sends a test email with detailed debug output
func (h *EmailHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from JWT token in context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if email service is available
	if h.emailService == nil {
		response := TestEmailResponse{
			Success: false,
			Message: "Email service is not configured. Please configure SMTP settings in your environment.",
		}
		respondJSON(w, http.StatusOK, response)
		return
	}

	// Parse request body
	var req TestEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email address
	if req.RecipientEmail == "" {
		respondError(w, http.StatusBadRequest, "Recipient email is required")
		return
	}

	if !isValidEmail(req.RecipientEmail) {
		respondError(w, http.StatusBadRequest, "Invalid email address format")
		return
	}

	// Send test email
	result := h.emailService.SendTestEmail(req.RecipientEmail)

	// Log the email attempt
	if h.emailLogService != nil {
		subject := "ActaLog Test Email"
		if err := h.emailLogService.LogTestEmail(req.RecipientEmail, subject, result, userID); err != nil {
			if h.logger != nil {
				h.logger.Error("Failed to log test email attempt: %v", err)
			}
		}
	}

	// Build response
	response := TestEmailResponse{
		Success: result.Success,
	}

	if result.Success {
		response.Message = "Test email sent successfully"
	} else {
		response.Message = "Failed to send test email"
		if result.Error != nil {
			response.Message = result.Error.Error()
		}
	}

	// Convert debug info to response format
	if result.DebugInfo != nil {
		response.DebugInfo = &SMTPDebugResponse{
			TotalDuration:   result.DebugInfo.TotalDuration.String(),
			ConnectionHost:  result.DebugInfo.ConnectionHost,
			ConnectionPort:  result.DebugInfo.ConnectionPort,
			ConnectionTLS:   result.DebugInfo.ConnectionTLS,
			ConnectionTime:  result.DebugInfo.ConnectionTime.String(),
			ConnectionError: result.DebugInfo.ConnectionError,
			AuthMethod:      result.DebugInfo.AuthMethod,
			AuthTime:        result.DebugInfo.AuthTime.String(),
			AuthError:       result.DebugInfo.AuthError,
			SendTime:        result.DebugInfo.SendTime.String(),
			SendError:       result.DebugInfo.SendError,
			SMTPResponses:   result.DebugInfo.SMTPResponses,
			FinalError:      result.DebugInfo.FinalError,
		}
	}

	if result.Success {
		respondJSON(w, http.StatusOK, response)
	} else {
		respondJSON(w, http.StatusOK, response) // Still return 200 with error details in body
	}
}

// isValidEmail validates email format using a simple regex
func isValidEmail(email string) bool {
	// Simple email regex - not perfect but catches most invalid formats
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
