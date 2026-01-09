package handler

import (
	"errors"
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/service"
)

// Common errors for mocks
var (
	ErrMockNotFound      = errors.New("not found")
	ErrMockUnauthorized  = errors.New("unauthorized")
	ErrMockInternalError = errors.New("internal error")
)

// Helper function to create test AuditLogService
func createTestAuditLogService() *service.AuditLogService {
	mockAuditLogRepo := NewMockAuditLogRepository()
	return service.NewAuditLogService(mockAuditLogRepo)
}

// MockMovementRepository is a mock implementation of MovementRepository
type MockMovementRepository struct {
	movements     []*domain.Movement
	shouldError   bool
	errorToReturn error
	searchResults []*domain.Movement
}

func NewMockMovementRepository() *MockMovementRepository {
	return &MockMovementRepository{
		movements: []*domain.Movement{
			{ID: 1, Name: "Back Squat", Type: "weightlifting", IsStandard: true},
			{ID: 2, Name: "Deadlift", Type: "weightlifting", IsStandard: true},
			{ID: 3, Name: "Bench Press", Type: "weightlifting", IsStandard: true},
		},
	}
}

func (m *MockMovementRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockMovementRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockMovementRepository) Create(movement *domain.Movement) error {
	if m.shouldError {
		return m.errorToReturn
	}
	movement.ID = int64(len(m.movements) + 1)
	m.movements = append(m.movements, movement)
	return nil
}

func (m *MockMovementRepository) GetByID(id int64) (*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, mov := range m.movements {
		if mov.ID == id {
			return mov, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockMovementRepository) GetByName(name string) (*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, mov := range m.movements {
		if mov.Name == name {
			return mov, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockMovementRepository) ListAll() ([]*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.movements, nil
}

func (m *MockMovementRepository) ListStandard() ([]*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Movement
	for _, mov := range m.movements {
		if mov.IsStandard {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *MockMovementRepository) ListByUser(userID int64) ([]*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Movement
	for _, mov := range m.movements {
		if mov.CreatedBy != nil && *mov.CreatedBy == userID {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *MockMovementRepository) ListAllUserCreated() ([]*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Movement
	for _, mov := range m.movements {
		if !mov.IsStandard {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *MockMovementRepository) ListAllUserCreatedWithUserInfo() ([]*domain.MovementWithCreator, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return []*domain.MovementWithCreator{}, nil
}

func (m *MockMovementRepository) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, movementType, creator string) ([]*domain.MovementWithCreator, int64, error) {
	if m.shouldError {
		return nil, 0, m.errorToReturn
	}
	return []*domain.MovementWithCreator{}, 0, nil
}

func (m *MockMovementRepository) CountAllUserCreated() (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	count := int64(0)
	for _, mov := range m.movements {
		if !mov.IsStandard {
			count++
		}
	}
	return count, nil
}

func (m *MockMovementRepository) Update(movement *domain.Movement) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for i, mov := range m.movements {
		if mov.ID == movement.ID {
			m.movements[i] = movement
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockMovementRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for i, mov := range m.movements {
		if mov.ID == id {
			m.movements = append(m.movements[:i], m.movements[i+1:]...)
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockMovementRepository) Search(query string, limit int) ([]*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	if m.searchResults != nil {
		return m.searchResults, nil
	}
	// Simple search implementation
	var result []*domain.Movement
	for _, mov := range m.movements {
		if len(result) >= limit {
			break
		}
		result = append(result, mov)
	}
	return result, nil
}

func (m *MockMovementRepository) CopyToStandard(id int64, newName string) (*domain.Movement, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	mov, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}
	newMov := &domain.Movement{
		ID:         int64(len(m.movements) + 1),
		Name:       newName,
		Type:       mov.Type,
		IsStandard: true,
	}
	m.movements = append(m.movements, newMov)
	return newMov, nil
}

// MockNotificationRepository is a mock implementation of NotificationRepository
type MockNotificationRepository struct {
	notifications []*domain.Notification
	shouldError   bool
	errorToReturn error
	unreadCount   int64
}

func NewMockNotificationRepository() *MockNotificationRepository {
	now := time.Now()
	readAt := now.Add(-time.Hour) // 1 hour ago for read notifications
	return &MockNotificationRepository{
		notifications: []*domain.Notification{
			{ID: 1, UserID: 1, Type: "pr", Title: "New PR!", Message: "You set a new PR", ReadAt: nil, CreatedAt: now},
			{ID: 2, UserID: 1, Type: "info", Title: "Welcome", Message: "Welcome to ActaLog", ReadAt: &readAt, CreatedAt: now},
		},
		unreadCount: 1,
	}
}

func (m *MockNotificationRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockNotificationRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockNotificationRepository) Create(notification *domain.Notification) error {
	if m.shouldError {
		return m.errorToReturn
	}
	notification.ID = int64(len(m.notifications) + 1)
	m.notifications = append(m.notifications, notification)
	return nil
}

func (m *MockNotificationRepository) GetByID(id int64) (*domain.Notification, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, n := range m.notifications {
		if n.ID == id {
			return n, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockNotificationRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	// Apply offset and limit
	if offset >= len(result) {
		return []*domain.Notification{}, nil
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockNotificationRepository) GetUnreadByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			result = append(result, n)
		}
	}
	if offset >= len(result) {
		return []*domain.Notification{}, nil
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockNotificationRepository) CountUnreadByUserID(userID int64) (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	count := int64(0)
	for _, n := range m.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

func (m *MockNotificationRepository) MarkAsRead(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	now := time.Now()
	for _, n := range m.notifications {
		if n.ID == id {
			n.ReadAt = &now
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockNotificationRepository) MarkAsUnread(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for _, n := range m.notifications {
		if n.ID == id {
			n.ReadAt = nil
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockNotificationRepository) MarkAllAsReadForUser(userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	now := time.Now()
	for _, n := range m.notifications {
		if n.UserID == userID {
			n.ReadAt = &now
		}
	}
	return nil
}

func (m *MockNotificationRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for i, n := range m.notifications {
		if n.ID == id {
			m.notifications = append(m.notifications[:i], m.notifications[i+1:]...)
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockNotificationRepository) DeleteOlderThan(cutoff time.Time) (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	count := int64(0)
	var remaining []*domain.Notification
	for _, n := range m.notifications {
		if n.CreatedAt.Before(cutoff) {
			count++
		} else {
			remaining = append(remaining, n)
		}
	}
	m.notifications = remaining
	return count, nil
}

// MockWODRepository is a mock implementation of WODRepository
type MockWODRepository struct {
	wods          []*domain.WOD
	shouldError   bool
	errorToReturn error
}

func NewMockWODRepository() *MockWODRepository {
	userID := int64(1)
	return &MockWODRepository{
		wods: []*domain.WOD{
			{ID: 1, Name: "Fran", ScoreType: "Time", IsStandard: true, Source: "CrossFit", Type: "Girl"},
			{ID: 2, Name: "Cindy", ScoreType: "Rounds+Reps", IsStandard: true, Source: "CrossFit", Type: "Girl"},
			{ID: 3, Name: "Murph", ScoreType: "Time", IsStandard: true, Source: "CrossFit", Type: "Hero"},
			{ID: 4, Name: "My Custom WOD", ScoreType: "Time", IsStandard: false, CreatedBy: &userID, Source: "Self-recorded", Type: "Self-created"},
		},
	}
}

func (m *MockWODRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockWODRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockWODRepository) Create(wod *domain.WOD) error {
	if m.shouldError {
		return m.errorToReturn
	}
	wod.ID = int64(len(m.wods) + 1)
	m.wods = append(m.wods, wod)
	return nil
}

func (m *MockWODRepository) GetByID(id int64) (*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, wod := range m.wods {
		if wod.ID == id {
			return wod, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockWODRepository) GetByName(name string) (*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, wod := range m.wods {
		if wod.Name == name {
			return wod, nil
		}
	}
	// Return nil, nil when not found (no duplicate)
	return nil, nil
}

func (m *MockWODRepository) List(filters map[string]interface{}, limit, offset int) ([]*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	if offset >= len(m.wods) {
		return []*domain.WOD{}, nil
	}
	result := m.wods[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockWODRepository) ListStandard(limit, offset int) ([]*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.WOD
	for _, wod := range m.wods {
		if wod.IsStandard {
			result = append(result, wod)
		}
	}
	if offset >= len(result) {
		return []*domain.WOD{}, nil
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockWODRepository) ListByUser(userID int64, limit, offset int) ([]*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.WOD
	for _, wod := range m.wods {
		if wod.CreatedBy != nil && *wod.CreatedBy == userID {
			result = append(result, wod)
		}
	}
	if offset >= len(result) {
		return []*domain.WOD{}, nil
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockWODRepository) ListAllUserCreated(limit, offset int) ([]*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.WOD
	for _, wod := range m.wods {
		if !wod.IsStandard {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *MockWODRepository) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WODWithCreator, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return []*domain.WODWithCreator{}, nil
}

func (m *MockWODRepository) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, scoreType, creator string) ([]*domain.WODWithCreator, int64, error) {
	if m.shouldError {
		return nil, 0, m.errorToReturn
	}
	return []*domain.WODWithCreator{}, 0, nil
}

func (m *MockWODRepository) CountAllUserCreated() (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	count := int64(0)
	for _, wod := range m.wods {
		if !wod.IsStandard {
			count++
		}
	}
	return count, nil
}

func (m *MockWODRepository) Update(wod *domain.WOD) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for i, w := range m.wods {
		if w.ID == wod.ID {
			m.wods[i] = wod
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockWODRepository) UpdateStandard(wod *domain.WOD) error {
	return m.Update(wod)
}

func (m *MockWODRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for i, wod := range m.wods {
		if wod.ID == id {
			m.wods = append(m.wods[:i], m.wods[i+1:]...)
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockWODRepository) Search(query string, limit int) ([]*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.WOD
	for _, wod := range m.wods {
		if len(result) >= limit {
			break
		}
		result = append(result, wod)
	}
	return result, nil
}

func (m *MockWODRepository) CopyToStandard(id int64, newName string) (*domain.WOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	wod, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}
	newWOD := &domain.WOD{
		ID:         int64(len(m.wods) + 1),
		Name:       newName,
		ScoreType:  wod.ScoreType,
		IsStandard: true,
	}
	m.wods = append(m.wods, newWOD)
	return newWOD, nil
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	users         []*domain.User
	shouldError   bool
	errorToReturn error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: []*domain.User{
			{ID: 1, Email: "admin@example.com", Name: "Admin User", Role: "admin"},
			{ID: 2, Email: "test@example.com", Name: "Test User", Role: "user"},
		},
	}
}

func (m *MockUserRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockUserRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockUserRepository) Create(user *domain.User) error {
	if m.shouldError {
		return m.errorToReturn
	}
	user.ID = int64(len(m.users) + 1)
	m.users = append(m.users, user)
	return nil
}

func (m *MockUserRepository) GetByID(id int64) (*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockUserRepository) GetByEmail(email string) (*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	// Return nil, nil when not found (no user exists with this email)
	return nil, nil
}

func (m *MockUserRepository) GetByResetToken(token string) (*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return nil, ErrMockNotFound
}

func (m *MockUserRepository) GetByVerificationToken(token string) (*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return nil, ErrMockNotFound
}

func (m *MockUserRepository) Update(user *domain.User) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for i, u := range m.users {
		if u.ID == user.ID {
			m.users[i] = user
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockUserRepository) UpdatePassword(userID int64, hashedPassword string) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for _, u := range m.users {
		if u.ID == userID {
			u.PasswordHash = hashedPassword
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockUserRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for i, u := range m.users {
		if u.ID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockUserRepository) List(limit, offset int) ([]*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	if offset >= len(m.users) {
		return []*domain.User{}, nil
	}
	result := m.users[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockUserRepository) Count() (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return int64(len(m.users)), nil
}

func (m *MockUserRepository) IncrementFailedAttempts(userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserRepository) ResetFailedAttempts(userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserRepository) LockAccount(userID int64, lockDuration time.Duration) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserRepository) UnlockAccount(userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserRepository) IsAccountLocked(userID int64) (bool, *time.Time, error) {
	if m.shouldError {
		return false, nil, m.errorToReturn
	}
	return false, nil, nil
}

func (m *MockUserRepository) DisableAccount(userID int64, disabledBy int64, reason string) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserRepository) EnableAccount(userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserRepository) ListAdmins() ([]*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var admins []*domain.User
	for _, u := range m.users {
		if u.Role == "admin" {
			admins = append(admins, u)
		}
	}
	return admins, nil
}

// MockOrganizationRepository is a mock implementation of OrganizationRepository
type MockOrganizationRepository struct {
	organizations []*domain.Organization
	shouldError   bool
	errorToReturn error
}

func NewMockOrganizationRepository() *MockOrganizationRepository {
	return &MockOrganizationRepository{
		organizations: []*domain.Organization{
			{ID: 1, Name: "Test Org"},
		},
	}
}

func (m *MockOrganizationRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockOrganizationRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockOrganizationRepository) Create(org *domain.Organization) error {
	if m.shouldError {
		return m.errorToReturn
	}
	org.ID = int64(len(m.organizations) + 1)
	m.organizations = append(m.organizations, org)
	return nil
}

func (m *MockOrganizationRepository) GetByID(id int64) (*domain.Organization, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, org := range m.organizations {
		if org.ID == id {
			return org, nil
		}
	}
	return nil, nil // Not found returns nil, nil per repository pattern
}

func (m *MockOrganizationRepository) GetByOwner(ownerID int64) (*domain.Organization, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	// Return first organization for simplicity in tests
	if len(m.organizations) > 0 {
		return m.organizations[0], nil
	}
	return nil, ErrMockNotFound
}

func (m *MockOrganizationRepository) Update(org *domain.Organization) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockOrganizationRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockOrganizationRepository) List(limit, offset int) ([]*domain.Organization, int64, error) {
	if m.shouldError {
		return nil, 0, m.errorToReturn
	}
	return m.organizations, int64(len(m.organizations)), nil
}

func (m *MockOrganizationRepository) GetMemberIDs(orgID int64) ([]int64, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return []int64{1, 2}, nil
}

func (m *MockOrganizationRepository) GetByName(name string) (*domain.Organization, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, org := range m.organizations {
		if org.Name == name {
			return org, nil
		}
	}
	return nil, nil // Not found returns nil, nil per repository pattern
}

func (m *MockOrganizationRepository) Count() (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return int64(len(m.organizations)), nil
}

func (m *MockOrganizationRepository) AddUserToOrganization(userID, orgID int64, role string) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockOrganizationRepository) RemoveUserFromOrganization(userID, orgID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockOrganizationRepository) GetUserOrganizations(userID int64) ([]*domain.Organization, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.organizations, nil
}

func (m *MockOrganizationRepository) GetOrganizationUsers(orgID int64) ([]*domain.User, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return []*domain.User{{ID: 1, Email: "test@example.com"}}, nil
}

func (m *MockOrganizationRepository) IsUserInOrganization(userID, orgID int64) (bool, error) {
	if m.shouldError {
		return false, m.errorToReturn
	}
	return true, nil
}

func (m *MockOrganizationRepository) GetUserOrganizationIDs(userID int64) ([]int64, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return []int64{1}, nil
}

// MockUserSettingsRepository is a mock implementation of UserSettingsRepository
type MockUserSettingsRepository struct {
	settings      map[int64]*domain.UserSettings
	shouldError   bool
	errorToReturn error
}

func NewMockUserSettingsRepository() *MockUserSettingsRepository {
	return &MockUserSettingsRepository{
		settings: make(map[int64]*domain.UserSettings),
	}
}

func (m *MockUserSettingsRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockUserSettingsRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockUserSettingsRepository) Create(settings *domain.UserSettings) error {
	if m.shouldError {
		return m.errorToReturn
	}
	m.settings[settings.UserID] = settings
	return nil
}

func (m *MockUserSettingsRepository) GetByUserID(userID int64) (*domain.UserSettings, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	if s, ok := m.settings[userID]; ok {
		return s, nil
	}
	// Return default settings
	return &domain.UserSettings{
		UserID:                  userID,
		NotificationPreferences: `{"email_on_pr":true,"email_on_announcement":true}`,
		DataExportFormat:        "JSON",
		Theme:                   "light",
		WeightUnit:              "lbs",
		DistanceUnit:            "miles",
	}, nil
}

func (m *MockUserSettingsRepository) Update(settings *domain.UserSettings) error {
	if m.shouldError {
		return m.errorToReturn
	}
	m.settings[settings.UserID] = settings
	return nil
}

func (m *MockUserSettingsRepository) Delete(userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	delete(m.settings, userID)
	return nil
}

// MockEmailService is a mock implementation of email.EmailService
type MockEmailService struct {
	shouldError   bool
	errorToReturn error
	sentEmails    []string
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		sentEmails: []string{},
	}
}

func (m *MockEmailService) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockEmailService) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockEmailService) SendPasswordResetEmail(to, resetURL string) error {
	if m.shouldError {
		return m.errorToReturn
	}
	m.sentEmails = append(m.sentEmails, to)
	return nil
}

func (m *MockEmailService) SendVerificationEmail(to, verifyURL string) error {
	if m.shouldError {
		return m.errorToReturn
	}
	m.sentEmails = append(m.sentEmails, to)
	return nil
}

func (m *MockEmailService) SendHTMLEmail(to, subject, htmlBody string) error {
	if m.shouldError {
		return m.errorToReturn
	}
	m.sentEmails = append(m.sentEmails, to)
	return nil
}

// MockAuditLogRepository is a mock implementation of AuditLogRepository
type MockAuditLogRepository struct {
	logs          []*domain.AuditLog
	shouldError   bool
	errorToReturn error
}

func NewMockAuditLogRepository() *MockAuditLogRepository {
	return &MockAuditLogRepository{
		logs: []*domain.AuditLog{},
	}
}

func (m *MockAuditLogRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockAuditLogRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockAuditLogRepository) Create(log *domain.AuditLog) error {
	if m.shouldError {
		return m.errorToReturn
	}
	log.ID = int64(len(m.logs) + 1)
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockAuditLogRepository) GetByID(id int64) (*domain.AuditLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockAuditLogRepository) List(filters domain.AuditLogFilters, limit, offset int) ([]*domain.AuditLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.logs, nil
}

func (m *MockAuditLogRepository) Count(filters domain.AuditLogFilters) (int, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return len(m.logs), nil
}

func (m *MockAuditLogRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.AuditLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.logs, nil
}

func (m *MockAuditLogRepository) GetByTargetUserID(targetUserID int64, limit, offset int) ([]*domain.AuditLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.logs, nil
}

func (m *MockAuditLogRepository) DeleteOlderThan(cutoff time.Time) (int, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return 0, nil
}

// MockDataChangeLogRepository is a mock implementation of DataChangeLogRepository
type MockDataChangeLogRepository struct {
	logs          []*domain.DataChangeLog
	shouldError   bool
	errorToReturn error
}

func NewMockDataChangeLogRepository() *MockDataChangeLogRepository {
	return &MockDataChangeLogRepository{
		logs: []*domain.DataChangeLog{},
	}
}

func (m *MockDataChangeLogRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockDataChangeLogRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockDataChangeLogRepository) Create(log *domain.DataChangeLog) error {
	if m.shouldError {
		return m.errorToReturn
	}
	log.ID = int64(len(m.logs) + 1)
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockDataChangeLogRepository) GetByID(id int64) (*domain.DataChangeLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockDataChangeLogRepository) List(filters domain.DataChangeLogFilters, limit, offset int) ([]*domain.DataChangeLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.logs, nil
}

func (m *MockDataChangeLogRepository) Count(filters domain.DataChangeLogFilters) (int, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return len(m.logs), nil
}

func (m *MockDataChangeLogRepository) GetByEntityID(entityType string, entityID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.logs, nil
}

func (m *MockDataChangeLogRepository) GetByUserID(userID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.logs, nil
}

func (m *MockDataChangeLogRepository) DeleteOlderThan(cutoff time.Time) (int, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return 0, nil
}

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository
type MockRefreshTokenRepository struct {
	tokens        []*domain.RefreshToken
	shouldError   bool
	errorToReturn error
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		tokens: []*domain.RefreshToken{},
	}
}

func (m *MockRefreshTokenRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockRefreshTokenRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockRefreshTokenRepository) Create(token *domain.RefreshToken) error {
	if m.shouldError {
		return m.errorToReturn
	}
	token.ID = int64(len(m.tokens) + 1)
	m.tokens = append(m.tokens, token)
	return nil
}

func (m *MockRefreshTokenRepository) GetByToken(token string) (*domain.RefreshToken, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, t := range m.tokens {
		if t.Token == token {
			return t, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockRefreshTokenRepository) GetByUserID(userID int64) ([]*domain.RefreshToken, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.RefreshToken
	for _, t := range m.tokens {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *MockRefreshTokenRepository) Revoke(tokenID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for _, t := range m.tokens {
		if t.ID == tokenID {
			now := time.Now()
			t.RevokedAt = &now
			return nil
		}
	}
	return ErrMockNotFound
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	now := time.Now()
	for _, t := range m.tokens {
		if t.UserID == userID {
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpired() error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockRefreshTokenRepository) Delete(tokenID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

// MockUserSubscriptionRepository is a mock implementation of UserSubscriptionRepository
type MockUserSubscriptionRepository struct {
	subscriptions []*domain.UserSubscription
	shouldError   bool
	errorToReturn error
}

func NewMockUserSubscriptionRepository() *MockUserSubscriptionRepository {
	return &MockUserSubscriptionRepository{
		subscriptions: []*domain.UserSubscription{},
	}
}

func (m *MockUserSubscriptionRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockUserSubscriptionRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockUserSubscriptionRepository) Create(sub *domain.UserSubscription) error {
	if m.shouldError {
		return m.errorToReturn
	}
	sub.ID = int64(len(m.subscriptions) + 1)
	m.subscriptions = append(m.subscriptions, sub)
	return nil
}

func (m *MockUserSubscriptionRepository) GetByID(id int64) (*domain.UserSubscription, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, sub := range m.subscriptions {
		if sub.ID == id {
			return sub, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockUserSubscriptionRepository) GetActiveByUserID(userID int64) (*domain.UserSubscription, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, sub := range m.subscriptions {
		if sub.UserID == userID && sub.Status == "active" {
			return sub, nil
		}
	}
	return nil, nil
}

func (m *MockUserSubscriptionRepository) GetByUserID(userID int64) ([]*domain.UserSubscription, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.UserSubscription
	for _, sub := range m.subscriptions {
		if sub.UserID == userID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *MockUserSubscriptionRepository) Update(sub *domain.UserSubscription) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserSubscriptionRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserSubscriptionRepository) ListAll() ([]*domain.UserSubscription, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.subscriptions, nil
}

func (m *MockUserSubscriptionRepository) MarkAsPaid(id int64, paymentDate time.Time, adminUserID int64, durationDays *int) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserSubscriptionRepository) MarkAsExpired(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserSubscriptionRepository) Cancel(id int64, reason string, adminUserID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserSubscriptionRepository) ListExpiring(days int) ([]*domain.UserSubscription, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	now := time.Now()
	futureDate := now.AddDate(0, 0, days)
	var result []*domain.UserSubscription
	for _, sub := range m.subscriptions {
		if sub.Status == domain.SubscriptionStatusActive &&
			!sub.IsPermanentFree &&
			sub.EndDate != nil &&
			sub.EndDate.After(now) &&
			sub.EndDate.Before(futureDate) {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *MockUserSubscriptionRepository) ListExpired() ([]*domain.UserSubscription, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	now := time.Now()
	var result []*domain.UserSubscription
	for _, sub := range m.subscriptions {
		if sub.Status == domain.SubscriptionStatusExpired ||
			(sub.Status == domain.SubscriptionStatusActive &&
				!sub.IsPermanentFree &&
				sub.EndDate != nil &&
				sub.EndDate.Before(now)) {
			result = append(result, sub)
		}
	}
	return result, nil
}

// MockUserWorkoutWODRepository is a mock implementation of UserWorkoutWODRepository
type MockUserWorkoutWODRepository struct {
	wods          []*domain.UserWorkoutWOD
	shouldError   bool
	errorToReturn error
}

func NewMockUserWorkoutWODRepository() *MockUserWorkoutWODRepository {
	return &MockUserWorkoutWODRepository{
		wods: []*domain.UserWorkoutWOD{},
	}
}

func (m *MockUserWorkoutWODRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockUserWorkoutWODRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockUserWorkoutWODRepository) Create(uww *domain.UserWorkoutWOD) error {
	if m.shouldError {
		return m.errorToReturn
	}
	uww.ID = int64(len(m.wods) + 1)
	m.wods = append(m.wods, uww)
	return nil
}

func (m *MockUserWorkoutWODRepository) CreateBatch(wods []*domain.UserWorkoutWOD) error {
	if m.shouldError {
		return m.errorToReturn
	}
	for _, wod := range wods {
		wod.ID = int64(len(m.wods) + 1)
		m.wods = append(m.wods, wod)
	}
	return nil
}

func (m *MockUserWorkoutWODRepository) GetByID(id int64) (*domain.UserWorkoutWOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, wod := range m.wods {
		if wod.ID == id {
			return wod, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockUserWorkoutWODRepository) GetByUserWorkoutID(userWorkoutID int64) ([]*domain.UserWorkoutWOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.UserWorkoutWOD
	for _, wod := range m.wods {
		if wod.UserWorkoutID == userWorkoutID {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *MockUserWorkoutWODRepository) Update(uww *domain.UserWorkoutWOD) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserWorkoutWODRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserWorkoutWODRepository) DeleteByUserWorkoutID(userWorkoutID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockUserWorkoutWODRepository) GetBestTimeForWOD(userID, wodID int64) (*int, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	bestTime := 300 // 5 minutes in seconds
	return &bestTime, nil
}

func (m *MockUserWorkoutWODRepository) GetBestRoundsRepsForWOD(userID, wodID int64) (rounds *int, reps *int, err error) {
	if m.shouldError {
		return nil, nil, m.errorToReturn
	}
	r, p := 10, 5
	return &r, &p, nil
}

func (m *MockUserWorkoutWODRepository) GetPRWODs(userID int64, limit int) ([]*domain.UserWorkoutWOD, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return []*domain.UserWorkoutWOD{}, nil
}

func (m *MockUserWorkoutWODRepository) UpdatePRFlag(id int64, isPR bool) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

// MockWorkoutRepository is a mock implementation of WorkoutRepository
type MockWorkoutRepository struct {
	workouts      []*domain.Workout
	shouldError   bool
	errorToReturn error
}

func NewMockWorkoutRepository() *MockWorkoutRepository {
	userID := int64(1)
	return &MockWorkoutRepository{
		workouts: []*domain.Workout{
			{ID: 1, Name: "Push Day"},
			{ID: 2, Name: "Pull Day"},
			{ID: 3, Name: "My Custom Workout", CreatedBy: &userID},
		},
	}
}

func (m *MockWorkoutRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockWorkoutRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockWorkoutRepository) Create(workout *domain.Workout) error {
	if m.shouldError {
		return m.errorToReturn
	}
	workout.ID = int64(len(m.workouts) + 1)
	m.workouts = append(m.workouts, workout)
	return nil
}

func (m *MockWorkoutRepository) GetByID(id int64) (*domain.Workout, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, w := range m.workouts {
		if w.ID == id {
			return w, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockWorkoutRepository) GetByIDWithDetails(id int64) (*domain.Workout, error) {
	return m.GetByID(id)
}

func (m *MockWorkoutRepository) List(filters map[string]interface{}, limit, offset int) ([]*domain.Workout, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.workouts, nil
}

func (m *MockWorkoutRepository) ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy != nil && *w.CreatedBy == userID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *MockWorkoutRepository) ListStandard(limit, offset int) ([]*domain.Workout, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy == nil { // Standard workouts have no creator
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *MockWorkoutRepository) ListAllUserCreated(limit, offset int) ([]*domain.Workout, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy != nil { // User-created workouts have a creator
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *MockWorkoutRepository) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WorkoutWithCreator, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return []*domain.WorkoutWithCreator{}, nil
}

func (m *MockWorkoutRepository) CountAllUserCreated() (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	count := int64(0)
	for _, w := range m.workouts {
		if w.CreatedBy != nil { // User-created workouts have a creator
			count++
		}
	}
	return count, nil
}

func (m *MockWorkoutRepository) Count(userID *int64) (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return int64(len(m.workouts)), nil
}

func (m *MockWorkoutRepository) GetUsageStats(workoutID int64) (*domain.WorkoutWithUsageStats, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	w, err := m.GetByID(workoutID)
	if err != nil {
		return nil, err
	}
	return &domain.WorkoutWithUsageStats{Workout: *w, TimesUsed: 5}, nil
}

func (m *MockWorkoutRepository) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, creator string) ([]*domain.WorkoutWithCreator, int64, error) {
	if m.shouldError {
		return nil, 0, m.errorToReturn
	}
	return []*domain.WorkoutWithCreator{}, 0, nil
}

func (m *MockWorkoutRepository) Update(workout *domain.Workout) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockWorkoutRepository) Delete(id int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockWorkoutRepository) Search(query string, limit int) ([]*domain.Workout, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.workouts, nil
}

func (m *MockWorkoutRepository) CopyToStandard(id int64, newName string) (*domain.Workout, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	newWorkout := &domain.Workout{
		ID:        int64(len(m.workouts) + 1),
		Name:      newName,
		CreatedBy: nil, // nil CreatedBy indicates standard workout
	}
	m.workouts = append(m.workouts, newWorkout)
	return newWorkout, nil
}

// MockBackupService is a mock implementation of BackupService
type MockBackupService struct {
	backups       []domain.BackupMetadata
	shouldError   bool
	errorToReturn error
}

func NewMockBackupService() *MockBackupService {
	return &MockBackupService{
		backups: []domain.BackupMetadata{
			{
				Filename:       "backup_2024-01-15.json",
				FileSize:       1024,
				CreatedAt:      time.Now(),
				CreatedByEmail: "admin@example.com",
				Version:        "0.17.0",
				DatabaseDriver: "sqlite3",
				DatabaseName:   "test.db",
				TotalUsers:     5,
				TotalWorkouts:  100,
				TotalMovements: 50,
				TotalWODs:      20,
			},
		},
	}
}

func (m *MockBackupService) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockBackupService) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockBackupService) CreateBackup(createdByUserID int64) (string, error) {
	if m.shouldError {
		return "", m.errorToReturn
	}
	filename := "backup_test.json"
	backup := domain.BackupMetadata{
		Filename:       filename,
		FileSize:       2048,
		CreatedAt:      time.Now(),
		CreatedByEmail: "admin@example.com",
		Version:        "0.17.0",
		DatabaseDriver: "sqlite3",
		DatabaseName:   "test.db",
	}
	m.backups = append(m.backups, backup)
	return filename, nil
}

func (m *MockBackupService) ListBackups() ([]domain.BackupMetadata, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.backups, nil
}

func (m *MockBackupService) GetBackupMetadata(filename string) (*domain.BackupMetadata, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, b := range m.backups {
		if b.Filename == filename {
			return &b, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockBackupService) DownloadBackup(filename string) (string, error) {
	if m.shouldError {
		return "", m.errorToReturn
	}
	return "/tmp/backups/" + filename, nil
}

func (m *MockBackupService) UploadBackup(file interface{}, filename string, uploadedByUserID int64) (string, error) {
	if m.shouldError {
		return "", m.errorToReturn
	}
	// Add the uploaded backup to the list so GetBackupMetadata will find it
	backup := domain.BackupMetadata{
		Filename:       filename,
		FileSize:       1024,
		CreatedAt:      time.Now(),
		CreatedByEmail: "admin@example.com",
		Version:        "0.17.0",
		DatabaseDriver: "sqlite3",
		DatabaseName:   "test.db",
	}
	m.backups = append(m.backups, backup)
	return filename, nil
}

func (m *MockBackupService) DeleteBackup(filename string, deletedByUserID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockBackupService) RestoreBackup(filename string, restoredByUserID int64, mode domain.RestoreMode) (*domain.RestoreResult, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return &domain.RestoreResult{
		Mode:           mode,
		TablesRestored: 10,
		RecordsCreated: 500,
		RecordsUpdated: 0,
		RecordsSkipped: 0,
	}, nil
}

// createTestUserSettingsService creates a UserSettingsService with mock repositories
func createTestUserSettingsService() *service.UserSettingsService {
	settingsRepo := NewMockUserSettingsRepository()
	auditLogRepo := NewMockAuditLogRepository()
	return service.NewUserSettingsService(settingsRepo, auditLogRepo)
}

// MockNotificationLikeRepository is a mock implementation of NotificationLikeRepository
type MockNotificationLikeRepository struct {
	likes         map[int64][]*domain.NotificationLike
	userLikes     map[string]bool // key: "notifID-userID"
	shouldError   bool
	errorToReturn error
}

func NewMockNotificationLikeRepository() *MockNotificationLikeRepository {
	return &MockNotificationLikeRepository{
		likes:     make(map[int64][]*domain.NotificationLike),
		userLikes: make(map[string]bool),
	}
}

func (m *MockNotificationLikeRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockNotificationLikeRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockNotificationLikeRepository) Create(like *domain.NotificationLike) error {
	if m.shouldError {
		return m.errorToReturn
	}
	key := fmt.Sprintf("%d-%d", like.NotificationID, like.UserID)
	m.userLikes[key] = true
	m.likes[like.NotificationID] = append(m.likes[like.NotificationID], like)
	return nil
}

func (m *MockNotificationLikeRepository) Delete(notificationID, userID int64) error {
	if m.shouldError {
		return m.errorToReturn
	}
	key := fmt.Sprintf("%d-%d", notificationID, userID)
	delete(m.userLikes, key)
	return nil
}

func (m *MockNotificationLikeRepository) GetByNotificationID(notificationID int64) ([]*domain.NotificationLike, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	return m.likes[notificationID], nil
}

func (m *MockNotificationLikeRepository) GetLikeCount(notificationID int64) (int, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	return len(m.likes[notificationID]), nil
}

func (m *MockNotificationLikeRepository) HasUserLiked(notificationID, userID int64) (bool, error) {
	if m.shouldError {
		return false, m.errorToReturn
	}
	key := fmt.Sprintf("%d-%d", notificationID, userID)
	return m.userLikes[key], nil
}

// createTestNotificationLikeService creates a NotificationLikeService with mock repositories
func createTestNotificationLikeService() *service.NotificationLikeService {
	likeRepo := NewMockNotificationLikeRepository()
	notificationRepo := NewMockNotificationRepository()
	return service.NewNotificationLikeService(likeRepo, notificationRepo)
}

// MockEmailLogRepository is a mock implementation of EmailLogRepository
type MockEmailLogRepository struct {
	logs          []*domain.EmailLog
	shouldError   bool
	errorToReturn error
}

func NewMockEmailLogRepository() *MockEmailLogRepository {
	now := time.Now()
	userID := int64(1)
	userEmail := "admin@example.com"
	errMsg := "SMTP connection failed"
	return &MockEmailLogRepository{
		logs: []*domain.EmailLog{
			{
				ID:              1,
				RecipientEmail:  "test@example.com",
				EmailType:       domain.EmailTypeTest,
				Subject:         "ActaLog Test Email",
				Success:         true,
				SentByUserID:    &userID,
				CreatedAt:       now,
				SentByUserEmail: &userEmail,
			},
			{
				ID:              2,
				RecipientEmail:  "user@example.com",
				EmailType:       domain.EmailTypePasswordReset,
				Subject:         "Password Reset Request",
				Success:         true,
				CreatedAt:       now.Add(-time.Hour),
				SentByUserEmail: nil,
			},
			{
				ID:              3,
				RecipientEmail:  "failed@example.com",
				EmailType:       domain.EmailTypeTest,
				Subject:         "ActaLog Test Email",
				Success:         false,
				ErrorMessage:    &errMsg,
				SentByUserID:    &userID,
				CreatedAt:       now.Add(-2 * time.Hour),
				SentByUserEmail: &userEmail,
			},
		},
	}
}

func (m *MockEmailLogRepository) SetError(err error) {
	m.shouldError = true
	m.errorToReturn = err
}

func (m *MockEmailLogRepository) ClearError() {
	m.shouldError = false
	m.errorToReturn = nil
}

func (m *MockEmailLogRepository) Create(log *domain.EmailLog) error {
	if m.shouldError {
		return m.errorToReturn
	}
	log.ID = int64(len(m.logs) + 1)
	m.logs = append(m.logs, log)
	return nil
}

func (m *MockEmailLogRepository) GetByID(id int64) (*domain.EmailLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, ErrMockNotFound
}

func (m *MockEmailLogRepository) List(filters domain.EmailLogFilters, limit, offset int) ([]*domain.EmailLog, error) {
	if m.shouldError {
		return nil, m.errorToReturn
	}

	var result []*domain.EmailLog
	for _, log := range m.logs {
		// Apply filters
		if filters.EmailType != nil && log.EmailType != *filters.EmailType {
			continue
		}
		if filters.Success != nil && log.Success != *filters.Success {
			continue
		}
		if filters.RecipientEmail != nil && log.RecipientEmail != *filters.RecipientEmail {
			continue
		}
		result = append(result, log)
	}

	// Apply pagination
	if offset >= len(result) {
		return []*domain.EmailLog{}, nil
	}
	result = result[offset:]
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (m *MockEmailLogRepository) Count(filters domain.EmailLogFilters) (int, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}

	count := 0
	for _, log := range m.logs {
		// Apply filters
		if filters.EmailType != nil && log.EmailType != *filters.EmailType {
			continue
		}
		if filters.Success != nil && log.Success != *filters.Success {
			continue
		}
		if filters.RecipientEmail != nil && log.RecipientEmail != *filters.RecipientEmail {
			continue
		}
		count++
	}
	return count, nil
}

func (m *MockEmailLogRepository) DeleteOlderThan(before time.Time) (int64, error) {
	if m.shouldError {
		return 0, m.errorToReturn
	}
	deleted := int64(0)
	var remaining []*domain.EmailLog
	for _, log := range m.logs {
		if log.CreatedAt.Before(before) {
			deleted++
		} else {
			remaining = append(remaining, log)
		}
	}
	m.logs = remaining
	return deleted, nil
}

// createTestEmailLogService creates an EmailLogService with mock repository
func createTestEmailLogService() *service.EmailLogService {
	emailLogRepo := NewMockEmailLogRepository()
	return service.NewEmailLogService(emailLogRepo)
}
