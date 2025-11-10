package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
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
}

// NewService creates a new email service
func NewService(config Config) *Service {
	return &Service{config: config}
}

// Message represents an email message
type Message struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// Send sends an email message
func (s *Service) Send(msg Message) error {
	// Build from header
	from := s.config.FromAddress
	if s.config.FromName != "" {
		from = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress)
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

	// Setup authentication
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)

	// For TLS connections (port 465)
	if s.config.SMTPPort == 465 {
		return s.sendWithTLS(addr, auth, msg.To, []byte(message))
	}

	// For STARTTLS connections (port 587) or plain (port 25)
	return smtp.SendMail(addr, auth, s.config.FromAddress, msg.To, []byte(message))
}

// sendWithTLS sends email using TLS (for port 465)
func (s *Service) sendWithTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	// TLS config
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}

	// Connect
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(s.config.FromAddress); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return client.Quit()
}

// SendPasswordResetEmail sends a password reset email
func (s *Service) SendPasswordResetEmail(to, resetURL string) error {
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
