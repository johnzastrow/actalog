package service

import (
	"strings"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

// userImportEmailMock is a mock email service for user import tests
type userImportEmailMock struct {
	sentEmails      []string
	sendError       error
	passwordResetTo []string
}

func newUserImportEmailMock() *userImportEmailMock {
	return &userImportEmailMock{
		sentEmails:      []string{},
		passwordResetTo: []string{},
	}
}

func (m *userImportEmailMock) SendVerificationEmail(to, token string) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.sentEmails = append(m.sentEmails, to)
	return nil
}

func (m *userImportEmailMock) SendPasswordResetEmail(to, resetURL string) error {
	if m.sendError != nil {
		return m.sendError
	}
	m.passwordResetTo = append(m.passwordResetTo, to)
	return nil
}

func (m *userImportEmailMock) SendWelcomeEmail(to, userName, appURL string) error {
	return nil
}

func (m *userImportEmailMock) SendAdminNotificationEmail(to, subject, body string) error {
	return nil
}

func (m *userImportEmailMock) SendHTMLEmail(to, subject, htmlBody string) error {
	return nil
}

func TestNewUserImportService(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	if svc == nil {
		t.Fatal("NewUserImportService returned nil")
	}
}

func TestUserImportService_PreviewUserImport_ValidCSV(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
john@example.com,John Doe,password123
jane@example.com,Jane Smith,securepass99`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result == nil {
		t.Fatal("PreviewUserImport() returned nil")
	}
	if result.TotalRows != 2 {
		t.Errorf("TotalRows = %d, want 2", result.TotalRows)
	}
	if result.ValidRows != 2 {
		t.Errorf("ValidRows = %d, want 2", result.ValidRows)
	}
	if result.InvalidRows != 0 {
		t.Errorf("InvalidRows = %d, want 0", result.InvalidRows)
	}
	if result.DuplicateRows != 0 {
		t.Errorf("DuplicateRows = %d, want 0", result.DuplicateRows)
	}
}

func TestUserImportService_PreviewUserImport_InvalidHeader(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `wrong,header,columns
john@example.com,John Doe,password123`

	_, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err == nil {
		t.Error("PreviewUserImport() should return error for invalid header")
	}
	if !strings.Contains(err.Error(), "invalid CSV header") {
		t.Errorf("Error should mention invalid header, got: %v", err)
	}
}

func TestUserImportService_PreviewUserImport_MissingEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
,John Doe,password123`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if result.ValidRows != 0 {
		t.Errorf("ValidRows = %d, want 0", result.ValidRows)
	}
	if len(result.Rows[0].Errors) == 0 {
		t.Error("Expected errors for invalid row")
	}
}

func TestUserImportService_PreviewUserImport_InvalidEmailFormat(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
not-an-email,John Doe,password123`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if !strings.Contains(result.Rows[0].Errors[0], "invalid email") {
		t.Errorf("Expected 'invalid email' error, got: %v", result.Rows[0].Errors)
	}
}

func TestUserImportService_PreviewUserImport_MissingName(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
john@example.com,,password123`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if !strings.Contains(result.Rows[0].Errors[0], "name is required") {
		t.Errorf("Expected 'name is required' error, got: %v", result.Rows[0].Errors)
	}
}

func TestUserImportService_PreviewUserImport_MissingPassword(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
john@example.com,John Doe,`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if !strings.Contains(result.Rows[0].Errors[0], "password is required") {
		t.Errorf("Expected 'password is required' error, got: %v", result.Rows[0].Errors)
	}
}

func TestUserImportService_PreviewUserImport_PasswordTooShort(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
john@example.com,John Doe,short`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}
	if !strings.Contains(result.Rows[0].Errors[0], "at least 8 characters") {
		t.Errorf("Expected password length error, got: %v", result.Rows[0].Errors)
	}
}

func TestUserImportService_PreviewUserImport_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	// Pre-create a user
	existingUser := &domain.User{
		Email: "existing@example.com",
		Name:  "Existing User",
	}
	userRepo.Create(existingUser)

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
existing@example.com,New Name,password123`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result.DuplicateRows != 1 {
		t.Errorf("DuplicateRows = %d, want 1", result.DuplicateRows)
	}
	if result.ValidRows != 0 {
		t.Errorf("ValidRows = %d, want 0 (duplicates don't count as valid)", result.ValidRows)
	}
	if !result.Rows[0].IsDuplicate {
		t.Error("Row should be marked as duplicate")
	}
}

func TestUserImportService_ConfirmUserImport_CreateUsers(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
john@example.com,John Doe,password123
jane@example.com,Jane Smith,securepass99`

	result, err := svc.ConfirmUserImport(strings.NewReader(csv), true)
	if err != nil {
		t.Fatalf("ConfirmUserImport() error = %v", err)
	}
	if result.CreatedCount != 2 {
		t.Errorf("CreatedCount = %d, want 2", result.CreatedCount)
	}
	if result.SkippedCount != 0 {
		t.Errorf("SkippedCount = %d, want 0", result.SkippedCount)
	}

	// Verify users were created
	user1, _ := userRepo.GetByEmail("john@example.com")
	if user1 == nil {
		t.Error("User john@example.com was not created")
	} else if user1.Role != "athlete" {
		t.Errorf("User role = %s, want 'athlete'", user1.Role)
	}

	user2, _ := userRepo.GetByEmail("jane@example.com")
	if user2 == nil {
		t.Error("User jane@example.com was not created")
	}
}

func TestUserImportService_ConfirmUserImport_SkipDuplicates(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	// Pre-create a user
	existingUser := &domain.User{
		Email: "existing@example.com",
		Name:  "Existing User",
	}
	userRepo.Create(existingUser)

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
existing@example.com,New Name,password123
newuser@example.com,New User,password456`

	result, err := svc.ConfirmUserImport(strings.NewReader(csv), true)
	if err != nil {
		t.Fatalf("ConfirmUserImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}
	if result.SkippedCount != 1 {
		t.Errorf("SkippedCount = %d, want 1", result.SkippedCount)
	}
}

func TestUserImportService_ConfirmUserImport_SkipsInvalidRows(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
valid@example.com,Valid User,password123
,Missing Email,password123
invalid-email,Bad Format,password123`

	result, err := svc.ConfirmUserImport(strings.NewReader(csv), true)
	if err != nil {
		t.Fatalf("ConfirmUserImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}
	if result.SkippedCount != 2 {
		t.Errorf("SkippedCount = %d, want 2", result.SkippedCount)
	}
}

func TestUserImportService_ConfirmUserImport_CreatesSubscription(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password
john@example.com,John Doe,password123`

	_, err := svc.ConfirmUserImport(strings.NewReader(csv), true)
	if err != nil {
		t.Fatalf("ConfirmUserImport() error = %v", err)
	}

	// Verify subscription was created
	if len(userSubRepo.subscriptions) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(userSubRepo.subscriptions))
	}

	// Check subscription properties
	for _, sub := range userSubRepo.subscriptions {
		if sub.SubscriptionType != domain.SubscriptionTypeFree {
			t.Errorf("SubscriptionType = %s, want 'free'", sub.SubscriptionType)
		}
		if !sub.IsPermanentFree {
			t.Error("Subscription should be permanent free")
		}
		if sub.Status != domain.SubscriptionStatusActive {
			t.Errorf("Status = %s, want 'active'", sub.Status)
		}
	}
}

func TestUserImportService_ExportUsersToCSV(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	// Create some users
	userRepo.Create(&domain.User{Email: "user1@example.com", Name: "User One"})
	userRepo.Create(&domain.User{Email: "user2@example.com", Name: "User Two"})

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csvBytes, err := svc.ExportUsersToCSV()
	if err != nil {
		t.Fatalf("ExportUsersToCSV() error = %v", err)
	}

	csvStr := string(csvBytes)

	// Verify header
	if !strings.Contains(csvStr, "email,name") {
		t.Error("CSV should contain header 'email,name'")
	}

	// Verify users are exported
	if !strings.Contains(csvStr, "user1@example.com") {
		t.Error("CSV should contain user1@example.com")
	}
	if !strings.Contains(csvStr, "User One") {
		t.Error("CSV should contain 'User One'")
	}
	if !strings.Contains(csvStr, "user2@example.com") {
		t.Error("CSV should contain user2@example.com")
	}
}

func TestUserImportService_ListUsersWithFilter(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	// Create some users
	userRepo.Create(&domain.User{Email: "user1@example.com", Name: "User One"})
	userRepo.Create(&domain.User{Email: "user2@example.com", Name: "User Two"})
	userRepo.Create(&domain.User{Email: "user3@example.com", Name: "User Three"})

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	filter := domain.UserListFilter{}
	users, count, err := svc.ListUsersWithFilter(filter, 10, 0)
	if err != nil {
		t.Fatalf("ListUsersWithFilter() error = %v", err)
	}

	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	if len(users) != 3 {
		t.Errorf("len(users) = %d, want 3", len(users))
	}
}

func TestUserImportService_SendBatchPasswordResetEmails_Success(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	// Create users
	user1 := &domain.User{Email: "user1@example.com", Name: "User One"}
	user2 := &domain.User{Email: "user2@example.com", Name: "User Two"}
	userRepo.Create(user1)
	userRepo.Create(user2)

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	result, err := svc.SendBatchPasswordResetEmails([]int64{user1.ID, user2.ID})
	if err != nil {
		t.Fatalf("SendBatchPasswordResetEmails() error = %v", err)
	}

	if result.TotalRequested != 2 {
		t.Errorf("TotalRequested = %d, want 2", result.TotalRequested)
	}
	if result.SuccessfullySent != 2 {
		t.Errorf("SuccessfullySent = %d, want 2", result.SuccessfullySent)
	}
	if result.FailedCount != 0 {
		t.Errorf("FailedCount = %d, want 0", result.FailedCount)
	}

	// Verify emails were sent
	if len(emailSvc.passwordResetTo) != 2 {
		t.Errorf("Expected 2 password reset emails, got %d", len(emailSvc.passwordResetTo))
	}

	// Verify reset tokens were set
	updatedUser1, _ := userRepo.GetByID(user1.ID)
	if updatedUser1.ResetToken == nil {
		t.Error("User1 should have reset token set")
	}
	if updatedUser1.ResetTokenExpiresAt == nil {
		t.Error("User1 should have reset token expiry set")
	}
}

func TestUserImportService_SendBatchPasswordResetEmails_UserNotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	result, err := svc.SendBatchPasswordResetEmails([]int64{999})
	if err != nil {
		t.Fatalf("SendBatchPasswordResetEmails() error = %v", err)
	}

	if result.FailedCount != 1 {
		t.Errorf("FailedCount = %d, want 1", result.FailedCount)
	}
	if result.SuccessfullySent != 0 {
		t.Errorf("SuccessfullySent = %d, want 0", result.SuccessfullySent)
	}
	if len(result.Errors) == 0 {
		t.Error("Expected error message for not found user")
	}
}

func TestUserImportService_SendBatchPasswordResetEmails_NoEmailService(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()

	// Create service without email service
	svc := NewUserImportService(userRepo, userSubRepo, nil, "http://localhost:3000")

	_, err := svc.SendBatchPasswordResetEmails([]int64{1})
	if err == nil {
		t.Error("Expected error when email service is not configured")
	}
	if !strings.Contains(err.Error(), "email service is not configured") {
		t.Errorf("Expected 'email service is not configured' error, got: %v", err)
	}
}

func TestUserImportService_PreviewUserImport_MultipleErrors(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	// Row with multiple validation errors
	csv := `email,name,password
not-valid-email,,short`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}

	if result.InvalidRows != 1 {
		t.Errorf("InvalidRows = %d, want 1", result.InvalidRows)
	}

	// Should have 3 errors: invalid email, missing name, password too short
	if len(result.Rows[0].Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d: %v", len(result.Rows[0].Errors), result.Rows[0].Errors)
	}
}

func TestUserImportService_PreviewUserImport_HeaderCaseInsensitive(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	// Test with uppercase header
	csv := `EMAIL,NAME,PASSWORD
john@example.com,John Doe,password123`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v (header should be case-insensitive)", err)
	}
	if result.ValidRows != 1 {
		t.Errorf("ValidRows = %d, want 1", result.ValidRows)
	}
}

func TestUserImportService_PreviewUserImport_EmptyCSV(t *testing.T) {
	userRepo := newMockUserRepo()
	userSubRepo := newMockUserSubscriptionRepo()
	emailSvc := newUserImportEmailMock()

	svc := NewUserImportService(userRepo, userSubRepo, emailSvc, "http://localhost:3000")

	csv := `email,name,password`

	result, err := svc.PreviewUserImport(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("PreviewUserImport() error = %v", err)
	}
	if result.TotalRows != 0 {
		t.Errorf("TotalRows = %d, want 0", result.TotalRows)
	}
}

func TestUserImportService_ConfirmUserImport_NilSubscriptionRepo(t *testing.T) {
	userRepo := newMockUserRepo()
	emailSvc := newUserImportEmailMock()

	// Create service without subscription repo (should still work, just skip subscription creation)
	svc := NewUserImportService(userRepo, nil, emailSvc, "http://localhost:3000")

	csv := `email,name,password
john@example.com,John Doe,password123`

	result, err := svc.ConfirmUserImport(strings.NewReader(csv), true)
	if err != nil {
		t.Fatalf("ConfirmUserImport() error = %v", err)
	}
	if result.CreatedCount != 1 {
		t.Errorf("CreatedCount = %d, want 1", result.CreatedCount)
	}

	// User should still be created
	user, _ := userRepo.GetByEmail("john@example.com")
	if user == nil {
		t.Error("User should have been created even without subscription repo")
	}
}
