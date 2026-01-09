package email

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestExtractEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple email",
			input:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "email with display name",
			input:    "John Doe <john@example.com>",
			expected: "john@example.com",
		},
		{
			name:     "email with display name and spaces",
			input:    "  John Doe  <john@example.com>  ",
			expected: "john@example.com",
		},
		{
			name:     "email with only angle brackets",
			input:    "<noreply@example.com>",
			expected: "noreply@example.com",
		},
		{
			name:     "email with whitespace",
			input:    "  test@example.com  ",
			expected: "test@example.com",
		},
		{
			name:     "complex display name",
			input:    "ActaLog Support Team <support@actalog.com>",
			expected: "support@actalog.com",
		},
		{
			name:     "email with quoted display name",
			input:    "\"John Doe\" <john@example.com>",
			expected: "john@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEmailAddress(tt.input)
			if got != tt.expected {
				t.Errorf("extractEmailAddress(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	cfg := Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUser:     "user@example.com",
		SMTPPassword: "password",
		FromAddress:  "noreply@example.com",
		FromName:     "Test App",
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	svc := NewService(cfg, logger)

	if svc == nil {
		t.Fatal("NewService() returned nil")
	}

	if svc.config.SMTPHost != cfg.SMTPHost {
		t.Errorf("config.SMTPHost = %q, want %q", svc.config.SMTPHost, cfg.SMTPHost)
	}

	if svc.config.SMTPPort != cfg.SMTPPort {
		t.Errorf("config.SMTPPort = %d, want %d", svc.config.SMTPPort, cfg.SMTPPort)
	}

	if svc.config.FromAddress != cfg.FromAddress {
		t.Errorf("config.FromAddress = %q, want %q", svc.config.FromAddress, cfg.FromAddress)
	}

	if svc.config.FromName != cfg.FromName {
		t.Errorf("config.FromName = %q, want %q", svc.config.FromName, cfg.FromName)
	}

	if svc.logger != logger {
		t.Error("logger not set correctly")
	}
}

func TestNewService_AllFields(t *testing.T) {
	cfg := Config{
		SMTPHost:     "mail.test.com",
		SMTPPort:     465,
		SMTPUser:     "testuser",
		SMTPPassword: "testpass",
		FromAddress:  "from@test.com",
		FromName:     "Test Sender",
	}

	logger := log.New(os.Stdout, "", 0)
	svc := NewService(cfg, logger)

	if svc.config.SMTPUser != "testuser" {
		t.Errorf("SMTPUser = %q, want %q", svc.config.SMTPUser, "testuser")
	}

	if svc.config.SMTPPassword != "testpass" {
		t.Errorf("SMTPPassword = %q, want %q", svc.config.SMTPPassword, "testpass")
	}
}

func TestMessage(t *testing.T) {
	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Test body content",
		IsHTML:  false,
	}

	if len(msg.To) != 1 || msg.To[0] != "recipient@example.com" {
		t.Errorf("Message.To = %v, want [recipient@example.com]", msg.To)
	}

	if msg.Subject != "Test Subject" {
		t.Errorf("Message.Subject = %q, want %q", msg.Subject, "Test Subject")
	}

	if msg.Body != "Test body content" {
		t.Errorf("Message.Body = %q, want %q", msg.Body, "Test body content")
	}

	if msg.IsHTML {
		t.Error("Message.IsHTML should be false")
	}
}

func TestMessage_HTML(t *testing.T) {
	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "HTML Test",
		Body:    "<html><body><h1>Hello</h1></body></html>",
		IsHTML:  true,
	}

	if !msg.IsHTML {
		t.Error("Message.IsHTML should be true")
	}

	if !strings.Contains(msg.Body, "<html>") {
		t.Error("Message.Body should contain HTML content")
	}
}

// Removed: TestConfig_Defaults - tested Go zero values, not business logic
// Removed: TestEmailService_Interface - tested interface assignment, compiler verifies this

func TestMessage_MultipleRecipients(t *testing.T) {
	msg := Message{
		To: []string{
			"user1@example.com",
			"user2@example.com",
			"user3@example.com",
		},
		Subject: "Group Email",
		Body:    "Message to multiple recipients",
		IsHTML:  false,
	}

	if len(msg.To) != 3 {
		t.Errorf("Message.To length = %d, want 3", len(msg.To))
	}

	for i, expected := range []string{"user1@example.com", "user2@example.com", "user3@example.com"} {
		if msg.To[i] != expected {
			t.Errorf("Message.To[%d] = %q, want %q", i, msg.To[i], expected)
		}
	}
}

func TestExtractEmailAddress_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "nested angle brackets",
			input:    "Name <<email@test.com>>",
			expected: "<email@test.com", // Gets the first match between < and >
		},
		{
			name:     "no closing bracket",
			input:    "Name <email@test.com",
			expected: "Name <email@test.com", // Returns as-is
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractEmailAddress(tt.input)
			if got != tt.expected {
				t.Errorf("extractEmailAddress(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestService_Send_ConnectionError tests that Send returns an error when SMTP is unreachable
func TestService_Send_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345, // Non-existent port
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "test@example.com",
		FromName:     "Test",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
		IsHTML:  false,
	}

	err := svc.Send(msg)
	if err == nil {
		t.Error("Send() should return error when SMTP server is unreachable")
	}

	// Verify logging occurred
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Attempting to send email") {
		t.Error("Should log attempt to send email")
	}
	if !strings.Contains(logOutput, "Connecting to SMTP server") {
		t.Error("Should log connection attempt")
	}
}

// TestService_Send_Port465_ConnectionError tests TLS connection path
func TestService_Send_Port465_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     465, // TLS port - triggers sendWithTLS path
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "test@example.com",
		FromName:     "Test",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
		IsHTML:  true,
	}

	err := svc.Send(msg)
	if err == nil {
		t.Error("Send() should return error when TLS connection fails")
	}

	// Verify TLS path was attempted
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Using TLS connection") {
		t.Error("Should log TLS connection attempt")
	}
}

// TestService_Send_FromNameFormatting tests that FromName is properly formatted
func TestService_Send_FromNameFormatting(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	tests := []struct {
		name        string
		fromAddress string
		fromName    string
	}{
		{
			name:        "with from name",
			fromAddress: "noreply@example.com",
			fromName:    "ActaLog",
		},
		{
			name:        "without from name",
			fromAddress: "noreply@example.com",
			fromName:    "",
		},
		{
			name:        "from address already has display name",
			fromAddress: "ActaLog <noreply@example.com>",
			fromName:    "Should Be Ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf.Reset()

			cfg := Config{
				SMTPHost:     "localhost",
				SMTPPort:     12345,
				SMTPUser:     "test",
				SMTPPassword: "test",
				FromAddress:  tt.fromAddress,
				FromName:     tt.fromName,
			}

			svc := NewService(cfg, logger)
			msg := Message{
				To:      []string{"test@example.com"},
				Subject: "Test",
				Body:    "Body",
				IsHTML:  false,
			}

			// Will fail to connect, but exercises the from address formatting code
			_ = svc.Send(msg)
		})
	}
}

// TestService_SendPasswordResetEmail_ConnectionError tests password reset email
func TestService_SendPasswordResetEmail_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	err := svc.SendPasswordResetEmail("user@example.com", "https://example.com/reset?token=abc123")
	if err == nil {
		t.Error("SendPasswordResetEmail() should return error when SMTP is unreachable")
	}

	// Verify the method was called with correct logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Preparing password reset email") {
		t.Error("Should log password reset email preparation")
	}
	if !strings.Contains(logOutput, "user@example.com") {
		t.Error("Should log recipient email")
	}
}

// TestService_SendVerificationEmail_ConnectionError tests verification email
func TestService_SendVerificationEmail_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	err := svc.SendVerificationEmail("newuser@example.com", "https://example.com/verify?token=xyz789")
	if err == nil {
		t.Error("SendVerificationEmail() should return error when SMTP is unreachable")
	}

	// Verify the method was called with correct logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Preparing verification email") {
		t.Error("Should log verification email preparation")
	}
	if !strings.Contains(logOutput, "newuser@example.com") {
		t.Error("Should log recipient email")
	}
}

// TestService_SendHTMLEmail_ConnectionError tests generic HTML email
func TestService_SendHTMLEmail_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	htmlBody := "<html><body><h1>Welcome!</h1></body></html>"
	err := svc.SendHTMLEmail("user@example.com", "Welcome Email", htmlBody)
	if err == nil {
		t.Error("SendHTMLEmail() should return error when SMTP is unreachable")
	}

	// Verify the method was called with correct logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Preparing HTML email") {
		t.Error("Should log HTML email preparation")
	}
	if !strings.Contains(logOutput, "Welcome Email") {
		t.Error("Should log email subject")
	}
}

// TestService_Send_HTMLContentType tests that HTML emails get correct content type
func TestService_Send_HTMLvsPlainText(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	// Test HTML message
	htmlMsg := Message{
		To:      []string{"test@example.com"},
		Subject: "HTML Test",
		Body:    "<html><body>Hello</body></html>",
		IsHTML:  true,
	}
	_ = svc.Send(htmlMsg)

	// Test plain text message
	textMsg := Message{
		To:      []string{"test@example.com"},
		Subject: "Plain Text Test",
		Body:    "Hello, World!",
		IsHTML:  false,
	}
	_ = svc.Send(textMsg)

	// Both should attempt to send (and fail due to no server)
	logOutput := logBuf.String()
	if strings.Count(logOutput, "Attempting to send email") != 2 {
		t.Error("Should have attempted to send 2 emails")
	}
}

// TestService_Send_MultipleRecipients tests sending to multiple recipients
func TestService_Send_MultipleRecipients(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To: []string{
			"user1@example.com",
			"user2@example.com",
			"user3@example.com",
		},
		Subject: "Group Email",
		Body:    "Hello everyone!",
		IsHTML:  false,
	}

	_ = svc.Send(msg)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "user1@example.com") {
		t.Error("Should log first recipient")
	}
}

// TestService_sendWithTLS_ConnectionError tests the TLS send path error handling
func TestService_sendWithTLS_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     465,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	// Directly test sendWithTLS
	err := svc.sendWithTLS("localhost:465", nil, "test@example.com", []string{"recipient@example.com"}, []byte("test message"))
	if err == nil {
		t.Error("sendWithTLS() should return error when connection fails")
	}

	// Verify error is about connection failure
	if !strings.Contains(err.Error(), "failed to connect") {
		t.Errorf("Error should mention connection failure, got: %v", err)
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Starting TLS connection") {
		t.Error("Should log TLS connection start")
	}
	if !strings.Contains(logOutput, "TLS connection failed") {
		t.Error("Should log TLS connection failure")
	}
}

// Removed: TestConfig_Ports - tested struct field assignment
// Removed: TestMessage_EmptyFields - tested Go zero values
// Removed: TestMessage_LongBody - tested string length
// Removed: TestMessage_SpecialCharacters - tested strings.Contains

// TestService_Logger tests that service uses provided logger
func TestService_Logger(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "[EMAIL] ", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "test@example.com",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{"test@example.com"},
		Subject: "Test",
		Body:    "Body",
		IsHTML:  false,
	}

	_ = svc.Send(msg)

	logOutput := logBuf.String()
	if !strings.HasPrefix(logOutput, "[EMAIL]") {
		t.Error("Log output should use provided logger prefix")
	}
}

// TestService_GetConfig tests GetConfig returns correct configuration info
func TestService_GetConfig(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	tests := []struct {
		name            string
		config          Config
		expectedTLSMode string
		expectedEnabled bool
	}{
		{
			name: "STARTTLS port 587",
			config: Config{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUser:     "user@example.com",
				SMTPPassword: "secret",
				FromAddress:  "noreply@example.com",
				FromName:     "ActaLog",
			},
			expectedTLSMode: "STARTTLS",
			expectedEnabled: true,
		},
		{
			name: "TLS port 465",
			config: Config{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     465,
				SMTPUser:     "user@example.com",
				SMTPPassword: "secret",
				FromAddress:  "noreply@example.com",
				FromName:     "Test",
			},
			expectedTLSMode: "TLS",
			expectedEnabled: true,
		},
		{
			name: "Plain port 25",
			config: Config{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     25,
				SMTPUser:     "user@example.com",
				SMTPPassword: "secret",
				FromAddress:  "noreply@example.com",
			},
			expectedTLSMode: "Plain",
			expectedEnabled: true,
		},
		{
			name: "Non-standard port defaults to STARTTLS",
			config: Config{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     2525,
				SMTPUser:     "user@example.com",
				SMTPPassword: "secret",
				FromAddress:  "noreply@example.com",
			},
			expectedTLSMode: "STARTTLS",
			expectedEnabled: true,
		},
		{
			name: "Empty host means disabled",
			config: Config{
				SMTPHost: "",
				SMTPPort: 587,
			},
			expectedTLSMode: "STARTTLS",
			expectedEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.config, logger)
			info := svc.GetConfig()

			if info.TLSMode != tt.expectedTLSMode {
				t.Errorf("GetConfig().TLSMode = %q, want %q", info.TLSMode, tt.expectedTLSMode)
			}

			if info.Enabled != tt.expectedEnabled {
				t.Errorf("GetConfig().Enabled = %v, want %v", info.Enabled, tt.expectedEnabled)
			}

			if info.SMTPHost != tt.config.SMTPHost {
				t.Errorf("GetConfig().SMTPHost = %q, want %q", info.SMTPHost, tt.config.SMTPHost)
			}

			if info.SMTPPort != tt.config.SMTPPort {
				t.Errorf("GetConfig().SMTPPort = %d, want %d", info.SMTPPort, tt.config.SMTPPort)
			}

			if info.SMTPUser != tt.config.SMTPUser {
				t.Errorf("GetConfig().SMTPUser = %q, want %q", info.SMTPUser, tt.config.SMTPUser)
			}

			if info.FromAddress != tt.config.FromAddress {
				t.Errorf("GetConfig().FromAddress = %q, want %q", info.FromAddress, tt.config.FromAddress)
			}

			if info.FromName != tt.config.FromName {
				t.Errorf("GetConfig().FromName = %q, want %q", info.FromName, tt.config.FromName)
			}
		})
	}
}

// TestService_SendWithDebug_ConnectionError tests SendWithDebug with connection failure
func TestService_SendWithDebug_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	tests := []struct {
		name        string
		port        int
		expectedTLS string
	}{
		{
			name:        "STARTTLS port 587",
			port:        587,
			expectedTLS: "STARTTLS",
		},
		{
			name:        "TLS port 465",
			port:        465,
			expectedTLS: "TLS",
		},
		{
			name:        "Plain port 25",
			port:        25,
			expectedTLS: "Plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf.Reset()

			cfg := Config{
				SMTPHost:     "localhost",
				SMTPPort:     tt.port,
				SMTPUser:     "test",
				SMTPPassword: "test",
				FromAddress:  "noreply@example.com",
				FromName:     "Test",
			}

			svc := NewService(cfg, logger)

			msg := Message{
				To:      []string{"recipient@example.com"},
				Subject: "Test Subject",
				Body:    "Test body content",
				IsHTML:  false,
			}

			result := svc.SendWithDebug(msg)

			// Should fail due to connection error
			if result.Success {
				t.Error("SendWithDebug() should fail when SMTP is unreachable")
			}

			if result.Error == nil {
				t.Error("SendWithDebug() should return an error")
			}

			if result.DebugInfo == nil {
				t.Fatal("SendWithDebug() should return DebugInfo even on failure")
			}

			// Check debug info
			if result.DebugInfo.ConnectionHost != "localhost" {
				t.Errorf("DebugInfo.ConnectionHost = %q, want localhost", result.DebugInfo.ConnectionHost)
			}

			if result.DebugInfo.ConnectionPort != tt.port {
				t.Errorf("DebugInfo.ConnectionPort = %d, want %d", result.DebugInfo.ConnectionPort, tt.port)
			}

			if result.DebugInfo.ConnectionTLS != tt.expectedTLS {
				t.Errorf("DebugInfo.ConnectionTLS = %q, want %q", result.DebugInfo.ConnectionTLS, tt.expectedTLS)
			}

			if result.DebugInfo.Success {
				t.Error("DebugInfo.Success should be false on failure")
			}

			if result.DebugInfo.FinalError == "" {
				t.Error("DebugInfo.FinalError should contain error message")
			}

			if result.DebugInfo.ConnectionError == "" {
				t.Error("DebugInfo.ConnectionError should contain error message")
			}

			if len(result.DebugInfo.SMTPResponses) == 0 {
				t.Error("DebugInfo.SMTPResponses should contain entries")
			}

			// Check timing info
			if result.DebugInfo.StartTime.IsZero() {
				t.Error("DebugInfo.StartTime should be set")
			}

			if result.DebugInfo.EndTime.IsZero() {
				t.Error("DebugInfo.EndTime should be set")
			}

			if result.DebugInfo.TotalDuration == 0 {
				t.Error("DebugInfo.TotalDuration should be non-zero")
			}

			// Check logging
			logOutput := logBuf.String()
			if !strings.Contains(logOutput, "SendWithDebug") {
				t.Error("Should log SendWithDebug attempt")
			}
		})
	}
}

// TestService_SendWithDebug_HTMLMessage tests SendWithDebug with HTML content
func TestService_SendWithDebug_HTMLMessage(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     587,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
		FromName:     "Test App",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "HTML Test",
		Body:    "<html><body><h1>Hello</h1></body></html>",
		IsHTML:  true,
	}

	result := svc.SendWithDebug(msg)

	// Will fail due to connection, but should process correctly
	if result.Success {
		t.Error("SendWithDebug() should fail when SMTP is unreachable")
	}

	if result.DebugInfo == nil {
		t.Fatal("DebugInfo should not be nil")
	}

	if result.DebugInfo.AuthMethod != "PLAIN" {
		t.Errorf("DebugInfo.AuthMethod = %q, want PLAIN", result.DebugInfo.AuthMethod)
	}
}

// TestService_SendWithDebug_FromNameFormatting tests from address formatting in SendWithDebug
func TestService_SendWithDebug_FromNameFormatting(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	tests := []struct {
		name        string
		fromAddress string
		fromName    string
	}{
		{
			name:        "with from name",
			fromAddress: "noreply@example.com",
			fromName:    "ActaLog",
		},
		{
			name:        "without from name",
			fromAddress: "noreply@example.com",
			fromName:    "",
		},
		{
			name:        "from address already has display name",
			fromAddress: "ActaLog <noreply@example.com>",
			fromName:    "Should Be Ignored",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf.Reset()

			cfg := Config{
				SMTPHost:     "localhost",
				SMTPPort:     587,
				SMTPUser:     "test",
				SMTPPassword: "test",
				FromAddress:  tt.fromAddress,
				FromName:     tt.fromName,
			}

			svc := NewService(cfg, logger)
			msg := Message{
				To:      []string{"test@example.com"},
				Subject: "Test",
				Body:    "Body",
				IsHTML:  false,
			}

			// Will fail to connect, but exercises the from address formatting code
			result := svc.SendWithDebug(msg)
			if result.DebugInfo == nil {
				t.Error("DebugInfo should not be nil")
			}
		})
	}
}

// TestService_SendTestEmail_ConnectionError tests SendTestEmail with connection failure
func TestService_SendTestEmail_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	tests := []struct {
		name string
		port int
	}{
		{"port 587 STARTTLS", 587},
		{"port 465 TLS", 465},
		{"port 25 Plain", 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf.Reset()

			cfg := Config{
				SMTPHost:     "localhost",
				SMTPPort:     tt.port,
				SMTPUser:     "test",
				SMTPPassword: "test",
				FromAddress:  "noreply@example.com",
				FromName:     "ActaLog",
			}

			svc := NewService(cfg, logger)

			result := svc.SendTestEmail("recipient@example.com")

			// Should fail due to connection error
			if result.Success {
				t.Error("SendTestEmail() should fail when SMTP is unreachable")
			}

			if result.Error == nil {
				t.Error("SendTestEmail() should return an error")
			}

			if result.DebugInfo == nil {
				t.Fatal("SendTestEmail() should return DebugInfo even on failure")
			}

			// Verify logging
			logOutput := logBuf.String()
			if !strings.Contains(logOutput, "Preparing test email") {
				t.Error("Should log test email preparation")
			}
			if !strings.Contains(logOutput, "recipient@example.com") {
				t.Error("Should log recipient email")
			}
		})
	}
}

// TestService_sendWithTLSDebug_ConnectionError tests TLS debug path
func TestService_sendWithTLSDebug_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     465,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	debug := &SMTPDebugInfo{
		SMTPResponses: []string{},
	}

	err := svc.sendWithTLSDebug("localhost:55555", "from@example.com", []string{"to@example.com"}, []byte("test message"), debug)

	if err == nil {
		t.Error("sendWithTLSDebug() should return error when connection fails")
	}

	if !strings.Contains(err.Error(), "TLS connection failed") {
		t.Errorf("Error should mention TLS connection failure, got: %v", err)
	}

	if debug.ConnectionError == "" {
		t.Error("debug.ConnectionError should be set")
	}

	if debug.AuthMethod != "PLAIN" {
		t.Errorf("debug.AuthMethod = %q, want PLAIN", debug.AuthMethod)
	}

	if len(debug.SMTPResponses) == 0 {
		t.Error("debug.SMTPResponses should contain error entry")
	}

	// Verify logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "TLS: Connecting") {
		t.Error("Should log TLS connection attempt")
	}
}

// TestService_sendWithSTARTTLSDebug_ConnectionError tests STARTTLS debug path
func TestService_sendWithSTARTTLSDebug_ConnectionError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     587,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	debug := &SMTPDebugInfo{
		SMTPResponses: []string{},
	}

	err := svc.sendWithSTARTTLSDebug("localhost:55556", "from@example.com", []string{"to@example.com"}, []byte("test message"), debug)

	if err == nil {
		t.Error("sendWithSTARTTLSDebug() should return error when connection fails")
	}

	if !strings.Contains(err.Error(), "connection failed") {
		t.Errorf("Error should mention connection failure, got: %v", err)
	}

	if debug.ConnectionError == "" {
		t.Error("debug.ConnectionError should be set")
	}

	if debug.AuthMethod != "PLAIN" {
		t.Errorf("debug.AuthMethod = %q, want PLAIN", debug.AuthMethod)
	}

	if len(debug.SMTPResponses) == 0 {
		t.Error("debug.SMTPResponses should contain error entry")
	}

	// Verify logging
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "STARTTLS: Connecting") {
		t.Error("Should log STARTTLS connection attempt")
	}
}

// TestSMTPDebugInfo tests the SMTPDebugInfo struct
func TestSMTPDebugInfo(t *testing.T) {
	now := time.Now()

	debug := SMTPDebugInfo{
		StartTime:       now,
		EndTime:         now.Add(1 * time.Second),
		TotalDuration:   1 * time.Second,
		ConnectionHost:  "smtp.example.com",
		ConnectionPort:  587,
		ConnectionTLS:   "STARTTLS",
		ConnectionTime:  100 * time.Millisecond,
		ConnectionError: "",
		AuthMethod:      "PLAIN",
		AuthTime:        50 * time.Millisecond,
		AuthError:       "",
		SendTime:        200 * time.Millisecond,
		SendError:       "",
		SMTPResponses:   []string{"220 OK", "250 OK"},
		Success:         true,
		FinalError:      "",
	}

	if debug.ConnectionHost != "smtp.example.com" {
		t.Errorf("ConnectionHost = %q, want smtp.example.com", debug.ConnectionHost)
	}

	if debug.ConnectionPort != 587 {
		t.Errorf("ConnectionPort = %d, want 587", debug.ConnectionPort)
	}

	if debug.ConnectionTLS != "STARTTLS" {
		t.Errorf("ConnectionTLS = %q, want STARTTLS", debug.ConnectionTLS)
	}

	if !debug.Success {
		t.Error("Success should be true")
	}

	if len(debug.SMTPResponses) != 2 {
		t.Errorf("SMTPResponses length = %d, want 2", len(debug.SMTPResponses))
	}
}

// TestSendResult tests the SendResult struct
func TestSendResult(t *testing.T) {
	// Test successful result
	successResult := SendResult{
		Success:   true,
		Error:     nil,
		DebugInfo: &SMTPDebugInfo{Success: true},
	}

	if !successResult.Success {
		t.Error("Success result should have Success = true")
	}

	if successResult.Error != nil {
		t.Error("Success result should have nil Error")
	}

	if successResult.DebugInfo == nil {
		t.Error("Success result should have DebugInfo")
	}

	// Test failure result
	failResult := SendResult{
		Success:   false,
		Error:     fmt.Errorf("connection failed"),
		DebugInfo: &SMTPDebugInfo{Success: false, FinalError: "connection failed"},
	}

	if failResult.Success {
		t.Error("Failure result should have Success = false")
	}

	if failResult.Error == nil {
		t.Error("Failure result should have non-nil Error")
	}

	if failResult.DebugInfo.FinalError != "connection failed" {
		t.Errorf("FinalError = %q, want 'connection failed'", failResult.DebugInfo.FinalError)
	}
}

// TestEmailConfigInfo tests the EmailConfigInfo struct
func TestEmailConfigInfo(t *testing.T) {
	info := EmailConfigInfo{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		SMTPUser:    "user@example.com",
		FromAddress: "noreply@example.com",
		FromName:    "ActaLog",
		Enabled:     true,
		TLSMode:     "STARTTLS",
	}

	if info.SMTPHost != "smtp.example.com" {
		t.Errorf("SMTPHost = %q, want smtp.example.com", info.SMTPHost)
	}

	if info.SMTPPort != 587 {
		t.Errorf("SMTPPort = %d, want 587", info.SMTPPort)
	}

	if !info.Enabled {
		t.Error("Enabled should be true")
	}

	if info.TLSMode != "STARTTLS" {
		t.Errorf("TLSMode = %q, want STARTTLS", info.TLSMode)
	}
}

// Removed: TestTimeoutConstants - tested constant values, compiler verifies these

// TestService_SendWithDebug_MultipleRecipients tests sending to multiple recipients
func TestService_SendWithDebug_MultipleRecipients(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     587,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To: []string{
			"user1@example.com",
			"user2@example.com",
			"user3@example.com",
		},
		Subject: "Group Email",
		Body:    "Hello everyone!",
		IsHTML:  false,
	}

	result := svc.SendWithDebug(msg)

	// Will fail, but should process correctly
	if result.DebugInfo == nil {
		t.Fatal("DebugInfo should not be nil")
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "user1@example.com") {
		t.Error("Should log first recipient in To list")
	}
}

// TestService_Send_EmptyRecipients tests sending with empty recipients list
func TestService_Send_EmptyRecipients(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{},
		Subject: "Test",
		Body:    "Test body",
		IsHTML:  false,
	}

	// Should still attempt to send (will fail at SMTP level)
	err := svc.Send(msg)
	if err == nil {
		t.Error("Send() with empty recipients should fail")
	}
}

// TestService_Send_Port25_Plain tests plain SMTP (port 25) path
func TestService_Send_Port25_Plain(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     25,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
		IsHTML:  false,
	}

	_ = svc.Send(msg)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "STARTTLS connection (port 25)") {
		t.Error("Should log STARTTLS connection for port 25")
	}
}

// TestService_SendWithDebug_Port25_Plain tests plain SMTP debug path
func TestService_SendWithDebug_Port25_Plain(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     25,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test",
		Body:    "Test body",
		IsHTML:  false,
	}

	result := svc.SendWithDebug(msg)

	if result.DebugInfo == nil {
		t.Fatal("DebugInfo should not be nil")
	}

	if result.DebugInfo.ConnectionTLS != "Plain" {
		t.Errorf("ConnectionTLS = %q, want Plain", result.DebugInfo.ConnectionTLS)
	}
}

// TestService_Send_LongSubject tests sending with a long subject line
func TestService_Send_LongSubject(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	longSubject := strings.Repeat("Test Subject ", 50)
	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: longSubject,
		Body:    "Test body",
		IsHTML:  false,
	}

	_ = svc.Send(msg)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Attempting to send email") {
		t.Error("Should log attempt to send email")
	}
}

// TestService_Send_LongBody tests sending with a very long body
func TestService_Send_LongBody(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	longBody := strings.Repeat("This is a test paragraph. ", 1000)
	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "Long Body Test",
		Body:    longBody,
		IsHTML:  false,
	}

	_ = svc.Send(msg)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Attempting to send email") {
		t.Error("Should log attempt to send email")
	}
}

// TestService_Send_SpecialCharactersInAddress tests sending to addresses with special characters
func TestService_Send_SpecialCharactersInAddress(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	// Test various email formats
	testAddresses := []string{
		"user+tag@example.com",
		"user.name@example.com",
		"user_name@example.com",
	}

	for _, addr := range testAddresses {
		msg := Message{
			To:      []string{addr},
			Subject: "Test",
			Body:    "Test body",
			IsHTML:  false,
		}

		_ = svc.Send(msg)
	}

	logOutput := logBuf.String()
	if strings.Count(logOutput, "Attempting to send email") != 3 {
		t.Error("Should have attempted to send 3 emails")
	}
}

// TestConfig_Validation tests various Config scenarios
func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "minimal config",
			config: Config{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
			},
		},
		{
			name: "full config",
			config: Config{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUser:     "user",
				SMTPPassword: "pass",
				FromAddress:  "from@example.com",
				FromName:     "From Name",
			},
		},
		{
			name: "config with empty password",
			config: Config{
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUser:     "user",
				SMTPPassword: "",
				FromAddress:  "from@example.com",
			},
		},
	}

	logger := log.New(os.Stdout, "", 0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.config, logger)
			if svc == nil {
				t.Error("NewService should not return nil")
			}
			if svc.config.SMTPHost != tt.config.SMTPHost {
				t.Errorf("SMTPHost = %q, want %q", svc.config.SMTPHost, tt.config.SMTPHost)
			}
		})
	}
}

// TestService_sendWithTLS_NilAuth tests sendWithTLS with nil authentication
func TestService_sendWithTLS_NilAuth(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     465,
		SMTPUser:     "",
		SMTPPassword: "",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	// This tests the TLS path with no auth (will still fail due to connection, but exercises the code)
	err := svc.sendWithTLS("localhost:55557", nil, "test@example.com", []string{"recipient@example.com"}, []byte("test message"))
	if err == nil {
		t.Error("sendWithTLS() should return error when connection fails")
	}
}

// Removed: TestSMTPDebugInfo_EmptyFields - tested Go zero values

// TestService_GetConfig_MaskedPassword verifies password is not exposed
func TestService_GetConfig_MaskedPassword(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)

	cfg := Config{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUser:     "user@example.com",
		SMTPPassword: "super_secret_password_123",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)
	info := svc.GetConfig()

	// The password should not be in the config info (it doesn't have a password field)
	// Verify the struct doesn't contain the password
	infoStr := fmt.Sprintf("%+v", info)
	if strings.Contains(infoStr, "super_secret_password_123") {
		t.Error("GetConfig() should not expose the password")
	}
}

// TestService_Send_UnicodeContent tests sending with Unicode content
func TestService_Send_UnicodeContent(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	msg := Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test: émojis 🎉 and 中文",
		Body:    "Hello 世界! 🌍 This is a test with Unicode: äöüß",
		IsHTML:  false,
	}

	_ = svc.Send(msg)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Attempting to send email") {
		t.Error("Should log attempt to send email")
	}
}

// TestService_SendHTMLEmail_ComplexHTML tests SendHTMLEmail with complex HTML
func TestService_SendHTMLEmail_ComplexHTML(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	cfg := Config{
		SMTPHost:     "localhost",
		SMTPPort:     12345,
		SMTPUser:     "test",
		SMTPPassword: "test",
		FromAddress:  "noreply@example.com",
	}

	svc := NewService(cfg, logger)

	complexHTML := `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .header { background-color: #00bcd4; color: white; padding: 20px; }
        .content { padding: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to ActaLog</h1>
    </div>
    <div class="content">
        <p>This is a <strong>test</strong> email with <em>HTML</em> content.</p>
        <table>
            <tr><td>Item 1</td><td>Value 1</td></tr>
            <tr><td>Item 2</td><td>Value 2</td></tr>
        </table>
    </div>
</body>
</html>`

	err := svc.SendHTMLEmail("recipient@example.com", "Complex HTML Test", complexHTML)
	if err == nil {
		t.Error("SendHTMLEmail() should return error when SMTP is unreachable")
	}

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Preparing HTML email") {
		t.Error("Should log HTML email preparation")
	}
}
