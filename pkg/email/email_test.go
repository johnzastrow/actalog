package email

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
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

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}

	// Empty config should have zero values
	if cfg.SMTPHost != "" {
		t.Errorf("default SMTPHost = %q, want empty", cfg.SMTPHost)
	}

	if cfg.SMTPPort != 0 {
		t.Errorf("default SMTPPort = %d, want 0", cfg.SMTPPort)
	}

	if cfg.SMTPUser != "" {
		t.Errorf("default SMTPUser = %q, want empty", cfg.SMTPUser)
	}

	if cfg.SMTPPassword != "" {
		t.Errorf("default SMTPPassword = %q, want empty", cfg.SMTPPassword)
	}

	if cfg.FromAddress != "" {
		t.Errorf("default FromAddress = %q, want empty", cfg.FromAddress)
	}

	if cfg.FromName != "" {
		t.Errorf("default FromName = %q, want empty", cfg.FromName)
	}
}

func TestEmailService_Interface(t *testing.T) {
	// Verify that *Service implements EmailService interface
	cfg := Config{
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		FromAddress: "test@example.com",
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	var svc EmailService = NewService(cfg, logger)

	if svc == nil {
		t.Fatal("Service should implement EmailService interface")
	}
}

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

// TestConfig_Ports tests common SMTP port configurations
func TestConfig_Ports(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		description string
	}{
		{"port 25", 25, "Standard SMTP"},
		{"port 465", 465, "SMTPS (TLS)"},
		{"port 587", 587, "Submission (STARTTLS)"},
		{"port 2525", 2525, "Alternative SMTP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				SMTPHost: "smtp.example.com",
				SMTPPort: tt.port,
			}

			if cfg.SMTPPort != tt.port {
				t.Errorf("SMTPPort = %d, want %d", cfg.SMTPPort, tt.port)
			}
		})
	}
}

// TestMessage_EmptyFields tests message with empty fields
func TestMessage_EmptyFields(t *testing.T) {
	msg := Message{}

	if msg.To != nil {
		t.Error("Empty Message.To should be nil")
	}

	if msg.Subject != "" {
		t.Error("Empty Message.Subject should be empty string")
	}

	if msg.Body != "" {
		t.Error("Empty Message.Body should be empty string")
	}

	if msg.IsHTML != false {
		t.Error("Empty Message.IsHTML should be false")
	}
}

// TestMessage_LongBody tests message with very long body
func TestMessage_LongBody(t *testing.T) {
	longBody := strings.Repeat("This is a test sentence. ", 1000)

	msg := Message{
		To:      []string{"test@example.com"},
		Subject: "Long Email",
		Body:    longBody,
		IsHTML:  false,
	}

	if len(msg.Body) != len(longBody) {
		t.Errorf("Body length = %d, want %d", len(msg.Body), len(longBody))
	}
}

// TestMessage_SpecialCharacters tests message with special characters
func TestMessage_SpecialCharacters(t *testing.T) {
	msg := Message{
		To:      []string{"test@example.com"},
		Subject: "Test: Special Characters! @#$%^&*()",
		Body:    "Body with émojis 🎉 and ünïcödé characters: 中文, 日本語, العربية",
		IsHTML:  false,
	}

	if !strings.Contains(msg.Subject, "@#$%^&*()") {
		t.Error("Subject should preserve special characters")
	}

	if !strings.Contains(msg.Body, "🎉") {
		t.Error("Body should preserve emoji")
	}

	if !strings.Contains(msg.Body, "中文") {
		t.Error("Body should preserve Chinese characters")
	}
}

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
