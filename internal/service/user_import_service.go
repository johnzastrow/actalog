package service

import (
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/pkg/auth"
	"github.com/johnzastrow/actalog/pkg/email"
)

// UserImportService handles user import/export and batch operations
type UserImportService struct {
	userRepo     domain.UserRepository
	userSubRepo  domain.UserSubscriptionRepository
	emailService email.EmailService
	appURL       string
}

// NewUserImportService creates a new user import service
func NewUserImportService(
	userRepo domain.UserRepository,
	userSubRepo domain.UserSubscriptionRepository,
	emailService email.EmailService,
	appURL string,
) *UserImportService {
	return &UserImportService{
		userRepo:     userRepo,
		userSubRepo:  userSubRepo,
		emailService: emailService,
		appURL:       appURL,
	}
}

// UserImportRow represents a single row during user import preview
type UserImportRow struct {
	RowNumber   int      `json:"row_number"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Password    string   `json:"-"` // Never serialize
	IsValid     bool     `json:"is_valid"`
	IsDuplicate bool     `json:"is_duplicate"`
	Errors      []string `json:"errors,omitempty"`
}

// UserImportResult represents the result of a user import operation
type UserImportResult struct {
	TotalRows        int             `json:"total_rows"`
	ValidRows        int             `json:"valid_rows"`
	InvalidRows      int             `json:"invalid_rows"`
	DuplicateRows    int             `json:"duplicate_rows"`
	CreatedCount     int             `json:"created_count"`
	SkippedCount     int             `json:"skipped_count"`
	Rows             []UserImportRow `json:"rows"`
	ValidationErrors []string        `json:"validation_errors,omitempty"`
}

// BatchPasswordResetResult represents the result of batch password reset emails
type BatchPasswordResetResult struct {
	TotalRequested   int      `json:"total_requested"`
	SuccessfullySent int      `json:"successfully_sent"`
	FailedCount      int      `json:"failed_count"`
	Errors           []string `json:"errors,omitempty"`
}

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// PreviewUserImport validates CSV data and returns preview without saving
func (s *UserImportService) PreviewUserImport(csvData io.Reader) (*UserImportResult, error) {
	reader := csv.NewReader(csvData)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Normalize header (lowercase, trim spaces)
	for i, h := range header {
		header[i] = strings.ToLower(strings.TrimSpace(h))
	}

	// Validate header
	expectedHeader := []string{"email", "name", "password"}
	if !equalStringSlices(header, expectedHeader) {
		return nil, fmt.Errorf("invalid CSV header. Expected: %v, Got: %v", expectedHeader, header)
	}

	result := &UserImportResult{
		Rows: []UserImportRow{},
	}

	rowNumber := 1 // Start at 1 (header is 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row %d: %w", rowNumber+1, err)
		}

		rowNumber++
		row := s.parseUserRow(record, rowNumber)

		// Validate the row
		s.validateUserRow(&row)

		// Check for duplicate by email
		if row.IsValid {
			existingUser, err := s.userRepo.GetByEmail(row.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to check for duplicate user: %w", err)
			}
			if existingUser != nil {
				row.IsDuplicate = true
				result.DuplicateRows++
			}
		}

		if row.IsValid && !row.IsDuplicate {
			result.ValidRows++
		} else if !row.IsValid {
			result.InvalidRows++
		}

		result.Rows = append(result.Rows, row)
		result.TotalRows++
	}

	return result, nil
}

// parseUserRow parses a CSV record into a UserImportRow
func (s *UserImportService) parseUserRow(record []string, rowNumber int) UserImportRow {
	row := UserImportRow{
		RowNumber: rowNumber,
		IsValid:   true,
		Errors:    []string{},
	}

	if len(record) >= 1 {
		row.Email = strings.TrimSpace(record[0])
	}
	if len(record) >= 2 {
		row.Name = strings.TrimSpace(record[1])
	}
	if len(record) >= 3 {
		row.Password = record[2] // Don't trim password to preserve spaces if intentional
	}

	return row
}

// validateUserRow validates a user import row
func (s *UserImportService) validateUserRow(row *UserImportRow) {
	// Validate email
	if row.Email == "" {
		row.IsValid = false
		row.Errors = append(row.Errors, "email is required")
	} else if !emailRegex.MatchString(row.Email) {
		row.IsValid = false
		row.Errors = append(row.Errors, "invalid email format")
	}

	// Validate name
	if row.Name == "" {
		row.IsValid = false
		row.Errors = append(row.Errors, "name is required")
	}

	// Validate password
	if row.Password == "" {
		row.IsValid = false
		row.Errors = append(row.Errors, "password is required")
	} else if len(row.Password) < 8 {
		row.IsValid = false
		row.Errors = append(row.Errors, "password must be at least 8 characters")
	}
}

// ConfirmUserImport imports users after preview
func (s *UserImportService) ConfirmUserImport(csvData io.Reader, skipDuplicates bool) (*UserImportResult, error) {
	// First, run preview to validate
	preview, err := s.PreviewUserImport(csvData)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Process each valid row
	for i, row := range preview.Rows {
		if !row.IsValid {
			preview.SkippedCount++
			continue
		}

		if row.IsDuplicate {
			if skipDuplicates {
				preview.SkippedCount++
				continue
			}
			// If not skipping duplicates and row is duplicate, skip anyway (no update mode for users)
			preview.SkippedCount++
			continue
		}

		// Hash the password
		hashedPassword, err := auth.HashPassword(row.Password)
		if err != nil {
			preview.Rows[i].IsValid = false
			preview.Rows[i].Errors = append(preview.Rows[i].Errors, fmt.Sprintf("failed to hash password: %v", err))
			preview.SkippedCount++
			continue
		}

		// Create the user
		user := &domain.User{
			Email:         row.Email,
			PasswordHash:  hashedPassword,
			Name:          row.Name,
			Role:          "user", // Always create as regular user
			EmailVerified: false,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		err = s.userRepo.Create(user)
		if err != nil {
			preview.Rows[i].IsValid = false
			preview.Rows[i].Errors = append(preview.Rows[i].Errors, fmt.Sprintf("failed to create user: %v", err))
			preview.SkippedCount++
			continue
		}

		// Create permanent free subscription
		if s.userSubRepo != nil {
			subscription := &domain.UserSubscription{
				UserID:           user.ID,
				SubscriptionType: domain.SubscriptionTypeFree,
				Status:           domain.SubscriptionStatusActive,
				IsPermanentFree:  true,
				StartDate:        now,
				CreatedAt:        now,
				UpdatedAt:        now,
			}
			err = s.userSubRepo.Create(subscription)
			if err != nil {
				// Log warning but don't fail - admin can create subscription later
				fmt.Printf("warning: failed to create subscription for imported user %d: %v\n", user.ID, err)
			}
		}

		preview.CreatedCount++
	}

	return preview, nil
}

// ExportUsersToCSV exports all users to CSV format
func (s *UserImportService) ExportUsersToCSV() ([]byte, error) {
	// Get all users
	users, err := s.userRepo.List(10000, 0) // Get up to 10000 users
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"email", "name"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write each user
	for _, user := range users {
		record := []string{user.Email, user.Name}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

// ListUsersWithFilter returns users matching the filter criteria
func (s *UserImportService) ListUsersWithFilter(filter domain.UserListFilter, limit, offset int) ([]*domain.User, int64, error) {
	users, err := s.userRepo.ListWithFilter(filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	count, err := s.userRepo.CountWithFilter(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, count, nil
}

// SendBatchPasswordResetEmails sends password reset emails to selected users
func (s *UserImportService) SendBatchPasswordResetEmails(userIDs []int64) (*BatchPasswordResetResult, error) {
	result := &BatchPasswordResetResult{
		TotalRequested: len(userIDs),
		Errors:         []string{},
	}

	if s.emailService == nil {
		return nil, fmt.Errorf("email service is not configured")
	}

	for _, userID := range userIDs {
		// Get user
		user, err := s.userRepo.GetByID(userID)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("user %d: failed to fetch: %v", userID, err))
			continue
		}
		if user == nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("user %d: not found", userID))
			continue
		}

		// Generate reset token
		token, err := generateSecureToken()
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("user %d (%s): failed to generate token: %v", userID, user.Email, err))
			continue
		}

		// Set token expiration (1 hour from now)
		expiresAt := time.Now().Add(1 * time.Hour)

		// Update user with reset token
		user.ResetToken = &token
		user.ResetTokenExpiresAt = &expiresAt

		err = s.userRepo.Update(user)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("user %d (%s): failed to save token: %v", userID, user.Email, err))
			continue
		}

		// Send password reset email
		resetURL := fmt.Sprintf("%s/reset-password/%s", s.appURL, token)
		err = s.emailService.SendPasswordResetEmail(user.Email, resetURL)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("user %d (%s): failed to send email: %v", userID, user.Email, err))
			continue
		}

		result.SuccessfullySent++
	}

	return result, nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
