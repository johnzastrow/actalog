package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/pkg/version"
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

// EmailService defines the methods used by other packages so tests can provide fakes.
// The concrete *Service type implements this interface.
type EmailService interface {
	SendPasswordResetEmail(to, resetURL string) error
	SendVerificationEmail(to, verifyURL string) error
	SendHTMLEmail(to, subject, htmlBody string) error
	SendWelcomeEmail(to, userName, appURL string) error
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

// Timeout constants for SMTP operations
const (
	ConnectionTimeout = 10 * time.Second
	AuthTimeout       = 5 * time.Second
	SendTimeout       = 30 * time.Second
	OverallTimeout    = 60 * time.Second
)

// SMTPDebugInfo captures detailed SMTP conversation info for debugging
type SMTPDebugInfo struct {
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	TotalDuration time.Duration `json:"total_duration"`

	// Connection phase
	ConnectionHost  string        `json:"connection_host"`
	ConnectionPort  int           `json:"connection_port"`
	ConnectionTLS   string        `json:"connection_tls"` // "STARTTLS", "TLS", "Plain"
	ConnectionTime  time.Duration `json:"connection_time"`
	ConnectionError string        `json:"connection_error,omitempty"`

	// Authentication phase
	AuthMethod string        `json:"auth_method"`
	AuthTime   time.Duration `json:"auth_time"`
	AuthError  string        `json:"auth_error,omitempty"`

	// Send phase
	SendTime  time.Duration `json:"send_time"`
	SendError string        `json:"send_error,omitempty"`

	// SMTP responses captured during the conversation
	SMTPResponses []string `json:"smtp_responses"`

	// Overall status
	Success    bool   `json:"success"`
	FinalError string `json:"final_error,omitempty"`
}

// SendResult wraps the result of a send operation with debug info
type SendResult struct {
	Success   bool
	Error     error
	DebugInfo *SMTPDebugInfo
}

// EmailConfigInfo represents email configuration for display (with masked password)
type EmailConfigInfo struct {
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    int    `json:"smtp_port"`
	SMTPUser    string `json:"smtp_user"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
	Enabled     bool   `json:"enabled"`
	TLSMode     string `json:"tls_mode"` // "TLS", "STARTTLS", "Plain"
}

// extractEmailAddress extracts the email address from a string that may contain a display name
// Examples: "Name <email@example.com>" -> "email@example.com"
//
//	"email@example.com" -> "email@example.com"
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

// SendHTMLEmail sends a generic HTML email
func (s *Service) SendHTMLEmail(to, subject, htmlBody string) error {
	s.logger.Printf("[INFO] Preparing HTML email for %s with subject: %s", to, subject)

	return s.Send(Message{
		To:      []string{to},
		Subject: subject,
		Body:    htmlBody,
		IsHTML:  true,
	})
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *Service) SendWelcomeEmail(to, userName, appURL string) error {
	s.logger.Printf("[INFO] Preparing welcome email for %s", to)
	subject := "Welcome to ActaLog - Let's Get Started!"

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
        .step { background-color: white; padding: 15px; margin: 10px 0; border-radius: 4px; border-left: 4px solid #00bcd4; }
        .step-number { display: inline-block; background-color: #00bcd4; color: white; width: 24px; height: 24px; border-radius: 50%%; text-align: center; line-height: 24px; margin-right: 10px; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to ActaLog!</h1>
        </div>
        <div class="content">
            <h2>Hi %s! 👋</h2>
            <p>Welcome to ActaLog - your personal CrossFit workout tracker. We're excited to have you on board!</p>

            <h3>Getting Started</h3>
            <p>Here's how to make the most of ActaLog:</p>

            <div class="step">
                <span class="step-number">1</span>
                <strong>Log Your First Workout</strong>
                <p>Tap the <strong>Quick Log</strong> button on the dashboard to record your workout. Enter your movements, weights, reps, and times.</p>
            </div>

            <div class="step">
                <span class="step-number">2</span>
                <strong>Track Your Progress</strong>
                <p>Visit the <strong>Performance</strong> tab to see your personal records and progress over time. ActaLog automatically detects PRs!</p>
            </div>

            <div class="step">
                <span class="step-number">3</span>
                <strong>Browse WODs</strong>
                <p>Check out the <strong>WODs</strong> section to find benchmark workouts like Fran, Murph, and more. Use these as templates for your training.</p>
            </div>

            <p style="text-align: center; margin: 30px 0;">
                <a href="%s" class="button">Open ActaLog</a>
            </p>

            <h3>Need Help?</h3>
            <p>If you have any questions or need assistance, contact your administrator or reply to this email. We're here to help!</p>

            <p><strong>Happy training! 💪</strong></p>
        </div>
        <div class="footer">
            <p>&copy; 2024 ActaLog. All rights reserved.</p>
            <p>This is an automated welcome email from ActaLog.</p>
        </div>
    </div>
</body>
</html>
`, userName, appURL)

	return s.Send(Message{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}

// GetConfig returns the email configuration info (with masked password)
func (s *Service) GetConfig() EmailConfigInfo {
	tlsMode := "STARTTLS"
	if s.config.SMTPPort == 465 {
		tlsMode = "TLS"
	} else if s.config.SMTPPort == 25 {
		tlsMode = "Plain"
	}

	return EmailConfigInfo{
		SMTPHost:    s.config.SMTPHost,
		SMTPPort:    s.config.SMTPPort,
		SMTPUser:    s.config.SMTPUser,
		FromAddress: s.config.FromAddress,
		FromName:    s.config.FromName,
		Enabled:     s.config.SMTPHost != "",
		TLSMode:     tlsMode,
	}
}

// SendWithDebug sends an email message and captures detailed debug information
func (s *Service) SendWithDebug(msg Message) SendResult {
	debug := &SMTPDebugInfo{
		StartTime:      time.Now(),
		ConnectionHost: s.config.SMTPHost,
		ConnectionPort: s.config.SMTPPort,
		SMTPResponses:  []string{},
	}

	// Determine TLS mode
	if s.config.SMTPPort == 465 {
		debug.ConnectionTLS = "TLS"
	} else if s.config.SMTPPort == 587 {
		debug.ConnectionTLS = "STARTTLS"
	} else {
		debug.ConnectionTLS = "Plain"
	}

	s.logger.Printf("[INFO] SendWithDebug: Attempting to send email to %v, subject: %s", msg.To, msg.Subject)

	// Extract the actual email address from FromAddress
	fromEmail := extractEmailAddress(s.config.FromAddress)

	// Build from header
	from := s.config.FromAddress
	if s.config.FromName != "" && !strings.Contains(s.config.FromAddress, "<") {
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

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)

	var err error
	if s.config.SMTPPort == 465 {
		err = s.sendWithTLSDebug(addr, fromEmail, msg.To, []byte(message), debug)
	} else {
		err = s.sendWithSTARTTLSDebug(addr, fromEmail, msg.To, []byte(message), debug)
	}

	debug.EndTime = time.Now()
	debug.TotalDuration = debug.EndTime.Sub(debug.StartTime)

	if err != nil {
		debug.Success = false
		debug.FinalError = err.Error()
		s.logger.Printf("[ERROR] SendWithDebug: Failed to send email: %v", err)
		return SendResult{Success: false, Error: err, DebugInfo: debug}
	}

	debug.Success = true
	s.logger.Printf("[INFO] SendWithDebug: Email sent successfully in %v", debug.TotalDuration)
	return SendResult{Success: true, Error: nil, DebugInfo: debug}
}

// sendWithTLSDebug sends email using TLS (port 465) with debug capture
func (s *Service) sendWithTLSDebug(addr string, fromEmail string, to []string, msg []byte, debug *SMTPDebugInfo) error {
	debug.AuthMethod = "PLAIN"

	// Connection phase with timeout
	connStart := time.Now()
	s.logger.Printf("[INFO] TLS: Connecting to %s", addr)

	dialer := &net.Dialer{Timeout: ConnectionTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName: s.config.SMTPHost,
	})
	debug.ConnectionTime = time.Since(connStart)

	if err != nil {
		debug.ConnectionError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("CONNECTION ERROR: %v", err))
		return fmt.Errorf("TLS connection failed: %w", err)
	}
	defer conn.Close()

	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("Connected to %s via TLS in %v", addr, debug.ConnectionTime))
	s.logger.Printf("[INFO] TLS: Connection established in %v", debug.ConnectionTime)

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		debug.ConnectionError = err.Error()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authentication phase with timeout
	authStart := time.Now()
	s.logger.Printf("[INFO] TLS: Authenticating as %s", s.config.SMTPUser)

	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		debug.AuthTime = time.Since(authStart)
		debug.AuthError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("AUTH ERROR: %v", err))
		return fmt.Errorf("authentication failed: %w", err)
	}
	debug.AuthTime = time.Since(authStart)
	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("Authenticated as %s in %v", s.config.SMTPUser, debug.AuthTime))
	s.logger.Printf("[INFO] TLS: Authentication successful in %v", debug.AuthTime)

	// Send phase
	sendStart := time.Now()

	// Set sender
	if err := client.Mail(fromEmail); err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("MAIL FROM ERROR: %v", err))
		return fmt.Errorf("failed to set sender: %w", err)
	}
	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("MAIL FROM: %s - OK", fromEmail))

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			debug.SendTime = time.Since(sendStart)
			debug.SendError = err.Error()
			debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("RCPT TO ERROR: %v", err))
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("RCPT TO: %s - OK", recipient))
	}

	// Send message data
	w, err := client.Data()
	if err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("DATA ERROR: %v", err))
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	debug.SMTPResponses = append(debug.SMTPResponses, "DATA - OK")

	if _, err = w.Write(msg); err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = w.Close(); err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		return fmt.Errorf("failed to close writer: %w", err)
	}

	debug.SendTime = time.Since(sendStart)
	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("Message sent in %v", debug.SendTime))
	s.logger.Printf("[INFO] TLS: Message sent in %v", debug.SendTime)

	// Quit
	if err = client.Quit(); err != nil {
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("QUIT: %v (non-fatal)", err))
	} else {
		debug.SMTPResponses = append(debug.SMTPResponses, "QUIT - OK")
	}

	return nil
}

// sendWithSTARTTLSDebug sends email using STARTTLS (port 587) with debug capture
func (s *Service) sendWithSTARTTLSDebug(addr string, fromEmail string, to []string, msg []byte, debug *SMTPDebugInfo) error {
	debug.AuthMethod = "PLAIN"

	// Connection phase with timeout
	connStart := time.Now()
	s.logger.Printf("[INFO] STARTTLS: Connecting to %s", addr)

	dialer := &net.Dialer{Timeout: ConnectionTimeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		debug.ConnectionTime = time.Since(connStart)
		debug.ConnectionError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("CONNECTION ERROR: %v", err))
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		debug.ConnectionTime = time.Since(connStart)
		debug.ConnectionError = err.Error()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Say EHLO
	if err = client.Hello("localhost"); err != nil {
		debug.ConnectionTime = time.Since(connStart)
		debug.ConnectionError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("EHLO ERROR: %v", err))
		return fmt.Errorf("EHLO failed: %w", err)
	}
	debug.SMTPResponses = append(debug.SMTPResponses, "EHLO localhost - OK")

	// Start TLS
	tlsConfig := &tls.Config{ServerName: s.config.SMTPHost}
	if err = client.StartTLS(tlsConfig); err != nil {
		debug.ConnectionTime = time.Since(connStart)
		debug.ConnectionError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("STARTTLS ERROR: %v", err))
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	debug.ConnectionTime = time.Since(connStart)
	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("STARTTLS - OK, connected in %v", debug.ConnectionTime))
	s.logger.Printf("[INFO] STARTTLS: Connection established in %v", debug.ConnectionTime)

	// Authentication phase
	authStart := time.Now()
	s.logger.Printf("[INFO] STARTTLS: Authenticating as %s", s.config.SMTPUser)

	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	if err := client.Auth(auth); err != nil {
		debug.AuthTime = time.Since(authStart)
		debug.AuthError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("AUTH ERROR: %v", err))
		return fmt.Errorf("authentication failed: %w", err)
	}
	debug.AuthTime = time.Since(authStart)
	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("AUTH PLAIN - OK in %v", debug.AuthTime))
	s.logger.Printf("[INFO] STARTTLS: Authentication successful in %v", debug.AuthTime)

	// Send phase
	sendStart := time.Now()

	// Set sender
	if err := client.Mail(fromEmail); err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("MAIL FROM ERROR: %v", err))
		return fmt.Errorf("failed to set sender: %w", err)
	}
	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("MAIL FROM: %s - OK", fromEmail))

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			debug.SendTime = time.Since(sendStart)
			debug.SendError = err.Error()
			debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("RCPT TO ERROR: %v", err))
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("RCPT TO: %s - OK", recipient))
	}

	// Send message data
	w, err := client.Data()
	if err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("DATA ERROR: %v", err))
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	debug.SMTPResponses = append(debug.SMTPResponses, "DATA - OK")

	if _, err = w.Write(msg); err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = w.Close(); err != nil {
		debug.SendTime = time.Since(sendStart)
		debug.SendError = err.Error()
		return fmt.Errorf("failed to close writer: %w", err)
	}

	debug.SendTime = time.Since(sendStart)
	debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("Message sent in %v", debug.SendTime))
	s.logger.Printf("[INFO] STARTTLS: Message sent in %v", debug.SendTime)

	// Quit
	if err = client.Quit(); err != nil {
		debug.SMTPResponses = append(debug.SMTPResponses, fmt.Sprintf("QUIT: %v (non-fatal)", err))
	} else {
		debug.SMTPResponses = append(debug.SMTPResponses, "QUIT - OK")
	}

	return nil
}

// SendTestEmail sends a test email with system information
func (s *Service) SendTestEmail(to string) SendResult {
	s.logger.Printf("[INFO] Preparing test email for %s", to)

	tlsMode := "STARTTLS"
	if s.config.SMTPPort == 465 {
		tlsMode = "TLS"
	} else if s.config.SMTPPort == 25 {
		tlsMode = "Plain"
	}

	subject := "ActaLog Test Email"
	timestamp := time.Now().UTC().Format(time.RFC3339)

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #00bcd4; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f5f7fa; }
        .info-table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
        .info-table td { padding: 10px; border-bottom: 1px solid #ddd; }
        .info-table td:first-child { font-weight: bold; width: 40%%; }
        .success { color: #4caf50; font-weight: bold; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ActaLog</h1>
        </div>
        <div class="content">
            <h2>This is a test email from ActaLog</h2>
            <p class="success">✓ Your email configuration is working correctly!</p>

            <h3>System Information</h3>
            <table class="info-table">
                <tr><td>Server Version</td><td>%s</td></tr>
                <tr><td>Timestamp (UTC)</td><td>%s</td></tr>
                <tr><td>SMTP Server</td><td>%s:%d</td></tr>
                <tr><td>TLS Mode</td><td>%s</td></tr>
                <tr><td>From Address</td><td>%s</td></tr>
            </table>

            <p>If you received this email, your email configuration is working correctly.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 ActaLog. All rights reserved.</p>
            <p>This is an automated test email.</p>
        </div>
    </div>
</body>
</html>
`, version.FullVersion(), timestamp, s.config.SMTPHost, s.config.SMTPPort, tlsMode, s.config.FromAddress)

	return s.SendWithDebug(Message{
		To:      []string{to},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})
}
