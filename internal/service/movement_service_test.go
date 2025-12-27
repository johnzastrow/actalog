package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// testMovementRepo is a fully-featured mock for MovementRepository
type testMovementRepo struct {
	movements   map[int64]*domain.Movement
	nextID      int64
	createError error
	updateError error
	deleteError error
	getError    error
}

func newTestMovementRepo() *testMovementRepo {
	return &testMovementRepo{
		movements: make(map[int64]*domain.Movement),
		nextID:    1,
	}
}

func (m *testMovementRepo) Create(movement *domain.Movement) error {
	if m.createError != nil {
		return m.createError
	}
	movement.ID = m.nextID
	m.nextID++
	movement.CreatedAt = time.Now()
	movement.UpdatedAt = time.Now()
	m.movements[movement.ID] = movement
	return nil
}

func (m *testMovementRepo) GetByID(id int64) (*domain.Movement, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	mov, ok := m.movements[id]
	if !ok {
		return nil, nil // Return nil, nil for not found
	}
	return mov, nil
}

func (m *testMovementRepo) GetByName(name string) (*domain.Movement, error) {
	for _, mov := range m.movements {
		if strings.EqualFold(mov.Name, name) {
			return mov, nil
		}
	}
	return nil, nil
}

func (m *testMovementRepo) ListAll() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, mov := range m.movements {
		result = append(result, mov)
	}
	return result, nil
}

func (m *testMovementRepo) ListStandard() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, mov := range m.movements {
		if mov.IsStandard {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *testMovementRepo) ListByUser(userID int64) ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, mov := range m.movements {
		if mov.CreatedBy != nil && *mov.CreatedBy == userID {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *testMovementRepo) ListAllUserCreated() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, mov := range m.movements {
		if !mov.IsStandard {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *testMovementRepo) ListAllUserCreatedWithUserInfo() ([]*domain.MovementWithCreator, error) {
	var result []*domain.MovementWithCreator
	for _, mov := range m.movements {
		if !mov.IsStandard {
			result = append(result, &domain.MovementWithCreator{
				Movement:     *mov,
				CreatorEmail: "test@example.com",
				CreatorName:  "Test User",
			})
		}
	}
	return result, nil
}

func (m *testMovementRepo) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, movementType, creator string) ([]*domain.MovementWithCreator, int64, error) {
	var result []*domain.MovementWithCreator
	for _, mov := range m.movements {
		if !mov.IsStandard {
			// Apply filters
			if search != "" && !strings.Contains(strings.ToLower(mov.Name), strings.ToLower(search)) {
				continue
			}
			if movementType != "" && string(mov.Type) != movementType {
				continue
			}
			result = append(result, &domain.MovementWithCreator{
				Movement:     *mov,
				CreatorEmail: "test@example.com",
				CreatorName:  "Test User",
			})
		}
	}
	return result, int64(len(result)), nil
}

func (m *testMovementRepo) CountAllUserCreated() (int64, error) {
	count := int64(0)
	for _, mov := range m.movements {
		if !mov.IsStandard {
			count++
		}
	}
	return count, nil
}

func (m *testMovementRepo) Update(movement *domain.Movement) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.movements[movement.ID]; !ok {
		return sql.ErrNoRows
	}
	movement.UpdatedAt = time.Now()
	m.movements[movement.ID] = movement
	return nil
}

func (m *testMovementRepo) Delete(id int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, ok := m.movements[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.movements, id)
	return nil
}

func (m *testMovementRepo) Search(query string, limit int) ([]*domain.Movement, error) {
	var result []*domain.Movement
	queryLower := strings.ToLower(query)
	for _, mov := range m.movements {
		if strings.Contains(strings.ToLower(mov.Name), queryLower) {
			result = append(result, mov)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *testMovementRepo) CopyToStandard(id int64, newName string) (*domain.Movement, error) {
	original, ok := m.movements[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	newMov := &domain.Movement{
		ID:          m.nextID,
		Name:        newName,
		Type:        original.Type,
		Description: original.Description,
		IsStandard:  true,
		CreatedBy:   nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.nextID++
	m.movements[newMov.ID] = newMov
	return newMov, nil
}

// Mock DataChangeLogRepository for tests
type testDataChangeLogRepo struct {
	logs   []*domain.DataChangeLog
	nextID int64
}

func newTestDataChangeLogRepo() *testDataChangeLogRepo {
	return &testDataChangeLogRepo{
		logs:   make([]*domain.DataChangeLog, 0),
		nextID: 1,
	}
}

func (m *testDataChangeLogRepo) Create(log *domain.DataChangeLog) error {
	log.ID = m.nextID
	m.nextID++
	m.logs = append(m.logs, log)
	return nil
}

func (m *testDataChangeLogRepo) GetByID(id int64) (*domain.DataChangeLog, error) {
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, nil
}

func (m *testDataChangeLogRepo) List(filters domain.DataChangeLogFilters, limit, offset int) ([]*domain.DataChangeLog, error) {
	return m.logs, nil
}

func (m *testDataChangeLogRepo) Count(filters domain.DataChangeLogFilters) (int, error) {
	return len(m.logs), nil
}

func (m *testDataChangeLogRepo) GetByEntityID(entityType string, entityID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	var result []*domain.DataChangeLog
	for _, log := range m.logs {
		if log.EntityType == entityType && log.EntityID == entityID {
			result = append(result, log)
		}
	}
	return result, nil
}

func (m *testDataChangeLogRepo) GetByUserID(userID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	var result []*domain.DataChangeLog
	for _, log := range m.logs {
		if log.UserID == userID {
			result = append(result, log)
		}
	}
	return result, nil
}

func (m *testDataChangeLogRepo) DeleteOlderThan(before time.Time) (int, error) {
	return 0, nil
}

// Helper to create test movement service
func newTestMovementService() (*MovementService, *testMovementRepo) {
	repo := newTestMovementRepo()
	auditRepo := newMockAuditLogRepo()
	dataChangeLogRepo := newTestDataChangeLogRepo()
	dataChangeLogService := NewDataChangeLogService(dataChangeLogRepo)
	return NewMovementService(repo, dataChangeLogService, auditRepo), repo
}

// TestMovementService_Create tests creating movements
func TestMovementService_Create(t *testing.T) {
	tests := []struct {
		name        string
		movement    *domain.Movement
		userID      int64
		userEmail   string
		setupRepo   func(*testMovementRepo)
		wantErr     bool
		errContains string
	}{
		{
			name: "valid weightlifting movement",
			movement: &domain.Movement{
				Name: "Back Squat",
				Type: "weightlifting",
			},
			userID:    1,
			userEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name: "valid gymnastics movement",
			movement: &domain.Movement{
				Name: "Pull-up",
				Type: "gymnastics",
			},
			userID:    1,
			userEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name: "valid cardio movement",
			movement: &domain.Movement{
				Name: "Running",
				Type: "cardio",
			},
			userID:    1,
			userEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name: "movement with description",
			movement: &domain.Movement{
				Name:        "Deadlift",
				Type:        "weightlifting",
				Description: "Standard conventional deadlift",
			},
			userID:    1,
			userEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name: "empty name fails validation",
			movement: &domain.Movement{
				Name: "",
				Type: "weightlifting",
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "whitespace-only name fails validation",
			movement: &domain.Movement{
				Name: "   ",
				Type: "weightlifting",
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "empty type fails validation",
			movement: &domain.Movement{
				Name: "Bench Press",
				Type: "",
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "type is required",
		},
		{
			name: "repository error propagates",
			movement: &domain.Movement{
				Name: "Clean",
				Type: "weightlifting",
			},
			userID:    1,
			userEmail: "test@example.com",
			setupRepo: func(r *testMovementRepo) {
				r.createError = errors.New("database connection failed")
			},
			wantErr:     true,
			errContains: "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			err := svc.Create(tt.movement, tt.userID, tt.userEmail)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify the movement was created correctly
				if tt.movement.ID == 0 {
					t.Error("expected movement ID to be set")
				}
				if tt.movement.IsStandard {
					t.Error("expected movement to not be standard")
				}
				if tt.movement.CreatedAt.IsZero() {
					t.Error("expected CreatedAt to be set")
				}
				if tt.movement.UpdatedAt.IsZero() {
					t.Error("expected UpdatedAt to be set")
				}
			}
		})
	}
}

// TestMovementService_GetByID tests retrieving movements by ID
func TestMovementService_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		setupRepo   func(*testMovementRepo)
		wantErr     bool
		errContains string
		wantNil     bool
	}{
		{
			name: "existing movement",
			id:   1,
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:   1,
					Name: "Back Squat",
					Type: "weightlifting",
				}
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:        "non-existent movement",
			id:          999,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "repository error",
			id:   1,
			setupRepo: func(r *testMovementRepo) {
				r.getError = errors.New("database error")
			},
			wantErr:     true,
			errContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			movement, err := svc.GetByID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.wantNil && movement != nil {
					t.Error("expected nil movement")
				}
				if !tt.wantNil && movement == nil {
					t.Error("expected non-nil movement")
				}
			}
		})
	}
}

// TestMovementService_GetByName tests retrieving movements by name
func TestMovementService_GetByName(t *testing.T) {
	tests := []struct {
		name        string
		searchName  string
		setupRepo   func(*testMovementRepo)
		wantErr     bool
		errContains string
		wantNil     bool
	}{
		{
			name:       "existing movement",
			searchName: "Back Squat",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:   1,
					Name: "Back Squat",
					Type: "weightlifting",
				}
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:       "non-existent movement",
			searchName: "Nonexistent",
			wantErr:    false,
			wantNil:    true,
		},
		{
			name:        "empty name",
			searchName:  "",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "whitespace-only name",
			searchName:  "   ",
			wantErr:     true,
			errContains: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			movement, err := svc.GetByName(tt.searchName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.wantNil && movement != nil {
					t.Error("expected nil movement")
				}
				if !tt.wantNil && movement == nil {
					t.Error("expected non-nil movement")
				}
			}
		})
	}
}

// TestMovementService_ListAll tests listing all movements
func TestMovementService_ListAll(t *testing.T) {
	svc, repo := newTestMovementService()

	// Add some movements
	repo.movements[1] = &domain.Movement{ID: 1, Name: "Back Squat", Type: "weightlifting", IsStandard: true}
	repo.movements[2] = &domain.Movement{ID: 2, Name: "Pull-up", Type: "gymnastics", IsStandard: true}
	repo.movements[3] = &domain.Movement{ID: 3, Name: "Custom Move", Type: "other", IsStandard: false}

	movements, err := svc.ListAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(movements) != 3 {
		t.Errorf("expected 3 movements, got %d", len(movements))
	}
}

// TestMovementService_ListStandard tests listing standard movements only
func TestMovementService_ListStandard(t *testing.T) {
	svc, repo := newTestMovementService()

	// Add some movements
	repo.movements[1] = &domain.Movement{ID: 1, Name: "Back Squat", Type: "weightlifting", IsStandard: true}
	repo.movements[2] = &domain.Movement{ID: 2, Name: "Pull-up", Type: "gymnastics", IsStandard: true}
	repo.movements[3] = &domain.Movement{ID: 3, Name: "Custom Move", Type: "other", IsStandard: false}

	movements, err := svc.ListStandard()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(movements) != 2 {
		t.Errorf("expected 2 standard movements, got %d", len(movements))
	}

	for _, mov := range movements {
		if !mov.IsStandard {
			t.Errorf("expected only standard movements, got non-standard: %s", mov.Name)
		}
	}
}

// TestMovementService_Search tests searching movements
func TestMovementService_Search(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		limit     int
		setupRepo func(*testMovementRepo)
		wantCount int
	}{
		{
			name:  "search by name",
			query: "squat",
			limit: 10,
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{ID: 1, Name: "Back Squat", Type: "weightlifting"}
				r.movements[2] = &domain.Movement{ID: 2, Name: "Front Squat", Type: "weightlifting"}
				r.movements[3] = &domain.Movement{ID: 3, Name: "Pull-up", Type: "gymnastics"}
			},
			wantCount: 2,
		},
		{
			name:      "empty query returns empty",
			query:     "",
			limit:     10,
			wantCount: 0,
		},
		{
			name:      "whitespace query returns empty",
			query:     "   ",
			limit:     10,
			wantCount: 0,
		},
		{
			name:  "no matches",
			query: "zzznonexistent",
			limit: 10,
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{ID: 1, Name: "Back Squat", Type: "weightlifting"}
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			movements, err := svc.Search(tt.query, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(movements) != tt.wantCount {
				t.Errorf("expected %d movements, got %d", tt.wantCount, len(movements))
			}
		})
	}
}

// TestMovementService_Update tests updating movements
func TestMovementService_Update(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(*testMovementRepo)
		movement    *domain.Movement
		userID      int64
		userEmail   string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful update of custom movement",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Old Name",
					Type:       "weightlifting",
					IsStandard: false,
					CreatedAt:  time.Now().Add(-time.Hour),
				}
			},
			movement: &domain.Movement{
				ID:   1,
				Name: "New Name",
				Type: "weightlifting",
			},
			userID:    1,
			userEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name: "cannot update standard movement",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Standard Move",
					Type:       "weightlifting",
					IsStandard: true,
				}
			},
			movement: &domain.Movement{
				ID:   1,
				Name: "New Name",
				Type: "weightlifting",
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "unauthorized",
		},
		{
			name: "movement not found",
			movement: &domain.Movement{
				ID:   999,
				Name: "New Name",
				Type: "weightlifting",
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "empty name fails validation",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Old Name",
					Type:       "weightlifting",
					IsStandard: false,
				}
			},
			movement: &domain.Movement{
				ID:   1,
				Name: "",
				Type: "weightlifting",
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name: "empty type fails validation",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Old Name",
					Type:       "weightlifting",
					IsStandard: false,
				}
			},
			movement: &domain.Movement{
				ID:   1,
				Name: "New Name",
				Type: "",
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			err := svc.Update(tt.movement, tt.userID, tt.userEmail)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify the movement was updated
				updated := repo.movements[tt.movement.ID]
				if updated == nil {
					t.Error("expected movement to exist after update")
				} else if updated.Name != tt.movement.Name {
					t.Errorf("expected name %q, got %q", tt.movement.Name, updated.Name)
				}
			}
		})
	}
}

// TestMovementService_UpdateAsAdmin tests admin updates
func TestMovementService_UpdateAsAdmin(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(*testMovementRepo)
		movement    *domain.Movement
		userID      int64
		userEmail   string
		wantErr     bool
		errContains string
	}{
		{
			name: "admin can update standard movement",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Standard Move",
					Type:       "weightlifting",
					IsStandard: true,
					CreatedAt:  time.Now().Add(-time.Hour),
				}
			},
			movement: &domain.Movement{
				ID:   1,
				Name: "Updated Standard Move",
				Type: "weightlifting",
			},
			userID:    1,
			userEmail: "admin@example.com",
			wantErr:   false,
		},
		{
			name: "admin can update custom movement",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Custom Move",
					Type:       "gymnastics",
					IsStandard: false,
					CreatedAt:  time.Now().Add(-time.Hour),
				}
			},
			movement: &domain.Movement{
				ID:   1,
				Name: "Updated Custom Move",
				Type: "gymnastics",
			},
			userID:    1,
			userEmail: "admin@example.com",
			wantErr:   false,
		},
		{
			name: "movement not found",
			movement: &domain.Movement{
				ID:   999,
				Name: "New Name",
				Type: "weightlifting",
			},
			userID:      1,
			userEmail:   "admin@example.com",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "validation still applies for admin",
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Standard Move",
					Type:       "weightlifting",
					IsStandard: true,
				}
			},
			movement: &domain.Movement{
				ID:   1,
				Name: "",
				Type: "weightlifting",
			},
			userID:      1,
			userEmail:   "admin@example.com",
			wantErr:     true,
			errContains: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			err := svc.UpdateAsAdmin(tt.movement, tt.userID, tt.userEmail)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestMovementService_Delete tests deleting movements
func TestMovementService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		setupRepo   func(*testMovementRepo)
		userID      int64
		userEmail   string
		wantErr     bool
		errContains string
	}{
		{
			name: "successful delete of custom movement",
			id:   1,
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Custom Move",
					Type:       "other",
					IsStandard: false,
				}
			},
			userID:    1,
			userEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name: "cannot delete standard movement",
			id:   1,
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Standard Move",
					Type:       "weightlifting",
					IsStandard: true,
				}
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "unauthorized",
		},
		{
			name:        "movement not found",
			id:          999,
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "repository error propagates",
			id:   1,
			setupRepo: func(r *testMovementRepo) {
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Custom Move",
					Type:       "other",
					IsStandard: false,
				}
				r.deleteError = errors.New("foreign key constraint")
			},
			userID:      1,
			userEmail:   "test@example.com",
			wantErr:     true,
			errContains: "foreign key constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			err := svc.Delete(tt.id, tt.userID, tt.userEmail)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// Verify the movement was deleted
				if _, exists := repo.movements[tt.id]; exists {
					t.Error("expected movement to be deleted")
				}
			}
		})
	}
}

// TestMovementService_ListAllUserCreated tests listing user-created movements
func TestMovementService_ListAllUserCreated(t *testing.T) {
	svc, repo := newTestMovementService()

	userID := int64Ptr(1)
	repo.movements[1] = &domain.Movement{ID: 1, Name: "Standard Move", Type: "weightlifting", IsStandard: true}
	repo.movements[2] = &domain.Movement{ID: 2, Name: "Custom Move 1", Type: "gymnastics", IsStandard: false, CreatedBy: userID}
	repo.movements[3] = &domain.Movement{ID: 3, Name: "Custom Move 2", Type: "other", IsStandard: false, CreatedBy: userID}

	movements, count, err := svc.ListAllUserCreated()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(movements) != 2 {
		t.Errorf("expected 2 user-created movements, got %d", len(movements))
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

// TestMovementService_ListAllUserCreatedWithUserInfo tests listing with user info
func TestMovementService_ListAllUserCreatedWithUserInfo(t *testing.T) {
	svc, repo := newTestMovementService()

	userID := int64Ptr(1)
	repo.movements[1] = &domain.Movement{ID: 1, Name: "Standard Move", Type: "weightlifting", IsStandard: true}
	repo.movements[2] = &domain.Movement{ID: 2, Name: "Custom Move", Type: "gymnastics", IsStandard: false, CreatedBy: userID}

	movements, count, err := svc.ListAllUserCreatedWithUserInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(movements) != 1 {
		t.Errorf("expected 1 user-created movement, got %d", len(movements))
	}

	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	if movements[0].CreatorEmail == "" {
		t.Error("expected CreatorEmail to be set")
	}
}

// TestMovementService_ListAllUserCreatedWithUserInfoFiltered tests filtered listing
func TestMovementService_ListAllUserCreatedWithUserInfoFiltered(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		offset       int
		search       string
		movementType string
		creator      string
		setupRepo    func(*testMovementRepo)
		wantCount    int
	}{
		{
			name:         "filter by search term",
			limit:        10,
			offset:       0,
			search:       "squat",
			movementType: "",
			creator:      "",
			setupRepo: func(r *testMovementRepo) {
				userID := int64Ptr(1)
				r.movements[1] = &domain.Movement{ID: 1, Name: "Back Squat", Type: "weightlifting", IsStandard: false, CreatedBy: userID}
				r.movements[2] = &domain.Movement{ID: 2, Name: "Front Squat", Type: "weightlifting", IsStandard: false, CreatedBy: userID}
				r.movements[3] = &domain.Movement{ID: 3, Name: "Pull-up", Type: "gymnastics", IsStandard: false, CreatedBy: userID}
			},
			wantCount: 2,
		},
		{
			name:         "filter by type",
			limit:        10,
			offset:       0,
			search:       "",
			movementType: "weightlifting",
			creator:      "",
			setupRepo: func(r *testMovementRepo) {
				userID := int64Ptr(1)
				r.movements[1] = &domain.Movement{ID: 1, Name: "Back Squat", Type: "weightlifting", IsStandard: false, CreatedBy: userID}
				r.movements[2] = &domain.Movement{ID: 2, Name: "Pull-up", Type: "gymnastics", IsStandard: false, CreatedBy: userID}
			},
			wantCount: 1,
		},
		{
			name:         "no filters",
			limit:        10,
			offset:       0,
			search:       "",
			movementType: "",
			creator:      "",
			setupRepo: func(r *testMovementRepo) {
				userID := int64Ptr(1)
				r.movements[1] = &domain.Movement{ID: 1, Name: "Move 1", Type: "weightlifting", IsStandard: false, CreatedBy: userID}
				r.movements[2] = &domain.Movement{ID: 2, Name: "Move 2", Type: "gymnastics", IsStandard: false, CreatedBy: userID}
				r.movements[3] = &domain.Movement{ID: 3, Name: "Standard", Type: "other", IsStandard: true} // Should be excluded
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			movements, count, err := svc.ListAllUserCreatedWithUserInfoFiltered(tt.limit, tt.offset, tt.search, tt.movementType, tt.creator)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(movements) != tt.wantCount {
				t.Errorf("expected %d movements, got %d", tt.wantCount, len(movements))
			}

			if int(count) != tt.wantCount {
				t.Errorf("expected count %d, got %d", tt.wantCount, count)
			}
		})
	}
}

// TestMovementService_CopyToStandard tests copying a movement to standard
func TestMovementService_CopyToStandard(t *testing.T) {
	tests := []struct {
		name        string
		id          int64
		newName     string
		setupRepo   func(*testMovementRepo)
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful copy",
			id:      1,
			newName: "New Standard Movement",
			setupRepo: func(r *testMovementRepo) {
				userID := int64Ptr(1)
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Custom Move",
					Type:       "weightlifting",
					IsStandard: false,
					CreatedBy:  userID,
				}
			},
			wantErr: false,
		},
		{
			name:        "empty new name",
			id:          1,
			newName:     "",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "whitespace new name",
			id:          1,
			newName:     "   ",
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:    "duplicate standard name",
			id:      1,
			newName: "Existing Standard",
			setupRepo: func(r *testMovementRepo) {
				userID := int64Ptr(1)
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Custom Move",
					Type:       "weightlifting",
					IsStandard: false,
					CreatedBy:  userID,
				}
				r.movements[2] = &domain.Movement{
					ID:         2,
					Name:       "Existing Standard",
					Type:       "weightlifting",
					IsStandard: true,
				}
			},
			wantErr:     true,
			errContains: "already exists",
		},
		{
			name:    "allow duplicate with non-standard",
			id:      1,
			newName: "Same Name",
			setupRepo: func(r *testMovementRepo) {
				userID := int64Ptr(1)
				r.movements[1] = &domain.Movement{
					ID:         1,
					Name:       "Custom Move",
					Type:       "weightlifting",
					IsStandard: false,
					CreatedBy:  userID,
				}
				r.movements[2] = &domain.Movement{
					ID:         2,
					Name:       "Same Name",
					Type:       "weightlifting",
					IsStandard: false, // Not standard, so should be allowed
					CreatedBy:  userID,
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := newTestMovementService()
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			movement, err := svc.CopyToStandard(tt.id, tt.newName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if movement == nil {
					t.Error("expected non-nil movement")
				} else {
					if movement.Name != tt.newName {
						t.Errorf("expected name %q, got %q", tt.newName, movement.Name)
					}
					if !movement.IsStandard {
						t.Error("expected copied movement to be standard")
					}
					if movement.CreatedBy != nil {
						t.Error("expected CreatedBy to be nil for standard movement")
					}
				}
			}
		})
	}
}

// TestMovementService_ValidateMovement tests validation logic
func TestMovementService_ValidateMovement(t *testing.T) {
	tests := []struct {
		name        string
		movement    *domain.Movement
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid movement",
			movement: &domain.Movement{Name: "Valid", Type: "weightlifting"},
			wantErr:  false,
		},
		{
			name:        "missing name",
			movement:    &domain.Movement{Name: "", Type: "weightlifting"},
			wantErr:     true,
			errContains: "name is required",
		},
		{
			name:        "missing type",
			movement:    &domain.Movement{Name: "Valid", Type: ""},
			wantErr:     true,
			errContains: "type is required",
		},
		{
			name:        "both missing",
			movement:    &domain.Movement{Name: "", Type: ""},
			wantErr:     true,
			errContains: "name is required", // Name is checked first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := newTestMovementService()
			err := svc.validateMovement(tt.movement)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
