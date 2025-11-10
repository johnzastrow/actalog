package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"regexp"
	"strings"
)

// Config holds email configuration
type Config struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

// Service handles email sending
type Service struct {
	config Config
	logger *log.Logger
}

// NewService creates a new email service
func NewService(config Config, logger *log.Logger) *Service {
	return &Service{
		config: config,
		logger: logger,
	}
}

// Message represents an email message
type Message struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// extractEmailAddress extracts the email address from a string that may contain a display name
// Examples: "Name <email@example.com>" -> "email@example.com"
//           "email@example.com" -> "email@example.com"
func extractEmailAddress(addr string) string {
	// Try to extract email from "Name <email@example.com>" format
	re := regexp.MustCompile(`<([^>]+)>`)
	matches := re.FindStringSubmatch(addr)
	if len(matches) > 1 {
		return matches[1]
	}
	// Return as-is if no angle brackets found (assumes it's just an email)
	return strings.TrimSpace(addr)
}

// Send sends an email message
func (s *Service) Send(msg Message) error {
	s.logger.Printf("[INFO] Attempting to send email to %v, subject: %s", msg.To, msg.Subject)

	// Extract the actual email address from FromAddress (removes display name if present)
	fromEmail := extractEmailAddress(s.config.FromAddress)

	// Build from header (can include display name for email headers)
	from := s.config.FromAddress
	if s.config.FromName != "" && !strings.Contains(s.config.FromAddress, "<") {
		// Only add display name if FromAddress doesn't already include it
		from = fmt.Sprintf("%s <%s>", s.config.FromName, fromEmail)
	}

	// Build headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = strings.Join(msg.To, ", ")
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"

	if msg.IsHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + msg.Body

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	s.logger.Printf("[INFO] Connecting to SMTP server: %s (using user: %s)", addr, s.config.SMTPUser)

	// Setup authentication
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)

	var err error
	// For TLS connections (port 465)
	if s.config.SMTPPort == 465 {
		s.logger.Printf("[INFO] Using TLS connection (port 465)")
		// Pass the extracted email address (not the display name version)
		err = s.sendWithTLS(addr, auth, fromEmail, msg.To, []byte(message))
	} else {
		// For STARTTLS connections (port 587) or plain (port 25)
		s.logger.Printf("[INFO] Using STARTTLS connection (port %d)", s.config.SMTPPort)
		// Use extracted email address for SMTP envelope
		err = smtp.SendMail(addr, auth, fromEmail, msg.To, []byte(message))
	}

	if err != nil {
		s.logger.Printf("[ERROR] Failed to send email to %v: %v", msg.To, err)
		return err
	}

	s.logger.Printf("[INFO] Email sent successfully to %v", msg.To)
	return nil
}

// sendWithTLS sends email using TLS (for port 465)
func (s *Service) sendWithTLS(addr string, auth smtp.Auth, fromEmail string, to []string, msg []byte) error {
	s.logger.Printf("[INFO] Starting TLS connection to %s", addr)

	// TLS config
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	// Connect
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		s.logger.Printf("[ERROR] TLS connection failed: %v", err)
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()
	s.logger.Printf("[INFO] TLS connection established")

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		s.logger.Printf("[ERROR] Failed to create SMTP client: %v", err)
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()
	s.logger.Printf("[INFO] SMTP client created successfully")

	// Authenticate
	if auth != nil {
		s.logger.Printf("[INFO] Authenticating as %s", s.config.SMTPUser)
		if err := client.Auth(auth); err != nil {
			s.logger.Printf("[ERROR] SMTP authentication failed: %v", err)
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		s.logger.Printf("[INFO] SMTP authentication successful")
	}

	// Set sender (use extracted email address without display name)
	s.logger.Printf("[INFO] Setting sender: %s", fromEmail)
	if err := client.Mail(fromEmail); err != nil {
		s.logger.Printf("[ERROR] Failed to set sender: %v", err)
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	s.logger.Printf("[INFO] Setting recipients: %v", to)
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			s.logger.Printf("[ERROR] Failed to set recipient %s: %v", recipient, err)
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Send message
	s.logger.Printf("[INFO] Sending message data")
	w, err := client.Data()
	if err != nil {
		s.logger.Printf("[ERROR] Failed to get data writer: %v", err)
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		s.logger.Printf("[ERROR] Failed to write message: %v", err)
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		s.logger.Printf("[ERROR] Failed to close writer: %v", err)
		return fmt.Errorf("failed to close writer: %w", err)
	}

	s.logger.Printf("[INFO] Message data sent successfully, closing connection")
	return client.Quit()
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(to, resetURL string) error {
	s.logger.Printf("[INFO] Preparing password reset email for %s with URL: %s", to, resetURL)
	subject := "ActaLog - Password Reset Request"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #00bcd4; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f7fa; }
        .button { display: inline-block; padding: 12px 24px; background-color: #ffc107; color: #1a1a1a; text-decoration: none; border-radius: 4px; font-weight: bold; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ActaLog</h1>
        </div>
        <div class="content">
            <h2>Password Reset Request</h2>
            <p>You requested to reset your password for your ActaLog account.</p>
            <p>Click the button below to reset your password. This link will expire in 1 hour.</p>
            <p style="text-align: center; margin: 30px 0;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this URL into your browser:</p>
            <p style="word-break: break-all; background-color: white; padding: 10px; border-radius: 4px;">%s</p>
            <p><strong>If you didn't request this password reset, you can safely ignore this email.</strong></p>
        </div>
        <div class="footer">
            <p>&copy; 2024 ActaLog. All rights reserved.</p>
            <p>This is an automated email. Please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, resetURL, resetURL)

	return s.Send(Message{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// SendVerificationEmail sends an email verification email
func (s *Service) SendVerificationEmail(to, verifyURL string) error {
	s.logger.Printf("[INFO] Preparing verification email for %s with URL: %s", to, verifyURL)
	subject := "ActaLog - Verify Your Email Address"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #00bcd4; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f7fa; }
        .button { display: inline-block; padding: 12px 24px; background-color: #ffc107; color: #1a1a1a; text-decoration: none; border-radius: 4px; font-weight: bold; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ActaLog</h1>
        </div>
        <div class="content">
            <h2>Welcome to ActaLog!</h2>
            <p>Thanks for signing up! Please verify your email address to get started.</p>
            <p>Click the button below to verify your email. This link will expire in 24 hours.</p>
            <p style="text-align: center; margin: 30px 0;">
                <a href="%s" class="button">Verify Email</a>
            </p>
            <p>Or copy and paste this URL into your browser:</p>
            <p style="word-break: break-all; background-color: white; padding: 10px; border-radius: 4px;">%s</p>
            <p><strong>If you didn't create an ActaLog account, you can safely ignore this email.</strong></p>
        </div>
        <div class="footer">
            <p>&copy; 2024 ActaLog. All rights reserved.</p>
            <p>This is an automated email. Please do not reply.</p>
        </div>
    </div>
</body>
</html>
`, verifyURL, verifyURL)

	return s.Send(Message{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}
