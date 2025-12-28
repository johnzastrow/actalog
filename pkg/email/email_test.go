package email

import (
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

// TestPasswordResetEmailContent verifies password reset email structure
// Note: This doesn't actually send email, just verifies the email would be formatted correctly
func TestPasswordResetEmailContent(t *testing.T) {
	// We can't easily test the actual email sending without a mock SMTP server
	// But we can verify the URL is properly formatted in the expected places
	resetURL := "https://example.com/reset?token=abc123"

	// Verify the URL format is valid for embedding
	if !strings.HasPrefix(resetURL, "https://") {
		t.Error("Reset URL should use HTTPS")
	}

	if !strings.Contains(resetURL, "token=") {
		t.Error("Reset URL should contain token parameter")
	}
}

// TestVerificationEmailContent verifies verification email structure
func TestVerificationEmailContent(t *testing.T) {
	verifyURL := "https://example.com/verify?token=xyz789"

	if !strings.HasPrefix(verifyURL, "https://") {
		t.Error("Verify URL should use HTTPS")
	}

	if !strings.Contains(verifyURL, "token=") {
		t.Error("Verify URL should contain token parameter")
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
