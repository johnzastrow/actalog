package service

import (
	"database/sql"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// Extended mock movement repository for testing
type testMovementRepo struct {
	movements    map[int64]*domain.Movement
	nextID       int64
	createError  error
	getByIDError error
	updateError  error
	deleteError  error
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
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	movement, ok := m.movements[id]
	if !ok {
		return nil, nil
	}
	return movement, nil
}

func (m *testMovementRepo) GetByName(name string) (*domain.Movement, error) {
	for _, movement := range m.movements {
		if movement.Name == name {
			return movement, nil
		}
	}
	return nil, nil
}

func (m *testMovementRepo) ListAll() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		result = append(result, movement)
	}
	return result, nil
}

func (m *testMovementRepo) ListStandard() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		if movement.IsStandard {
			result = append(result, movement)
		}
	}
	return result, nil
}

func (m *testMovementRepo) ListByUser(userID int64) ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		if movement.CreatedBy != nil && *movement.CreatedBy == userID {
			result = append(result, movement)
		}
	}
	return result, nil
}

func (m *testMovementRepo) ListAllUserCreated() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		if !movement.IsStandard {
			result = append(result, movement)
		}
	}
	return result, nil
}

func (m *testMovementRepo) ListAllUserCreatedWithUserInfo() ([]*domain.MovementWithCreator, error) {
	return []*domain.MovementWithCreator{}, nil
}

func (m *testMovementRepo) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, movementType, creator string) ([]*domain.MovementWithCreator, int64, error) {
	return []*domain.MovementWithCreator{}, 0, nil
}

func (m *testMovementRepo) CountAllUserCreated() (int64, error) {
	count := int64(0)
	for _, movement := range m.movements {
		if !movement.IsStandard {
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
	for _, movement := range m.movements {
		if matchString(movement.Name, query) {
			result = append(result, movement)
			if len(result) >= limit {
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
	newMovement := &domain.Movement{
		ID:          m.nextID,
		Name:        newName,
		Description: original.Description,
		Type:        original.Type,
		IsStandard:  true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.nextID++
	m.movements[newMovement.ID] = newMovement
	return newMovement, nil
}

// =============================================================================
// Movement Service Tests
// =============================================================================

func TestMovementService_Create(t *testing.T) {
	repo := newTestMovementRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewMovementService(repo, nil, auditRepo)

	tests := []struct {
		name      string
		movement  *domain.Movement
		userID    int64
		userEmail string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid movement",
			movement: &domain.Movement{
				Name: "Back Squat",
				Type: "strength",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   false,
		},
		{
			name: "movement with description",
			movement: &domain.Movement{
				Name:        "Front Squat",
				Type:        "strength",
				Description: "Barbell front squat",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   false,
		},
		{
			name: "empty name",
			movement: &domain.Movement{
				Name: "",
				Type: "strength",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errMsg:    "movement name is required",
		},
		{
			name: "whitespace name",
			movement: &domain.Movement{
				Name: "   ",
				Type: "strength",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errMsg:    "movement name is required",
		},
		{
			name: "empty type",
			movement: &domain.Movement{
				Name: "Deadlift",
				Type: "",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errMsg:    "movement type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Create(tt.movement, tt.userID, tt.userEmail)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Create() error = %v, want %v", err, tt.errMsg)
				}
			}
			if !tt.wantErr {
				if tt.movement.ID == 0 {
					t.Error("Create() did not set movement ID")
				}
				if tt.movement.IsStandard {
					t.Error("Create() should set IsStandard to false")
				}
			}
		})
	}
}

func TestMovementService_GetByID(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create a movement
	movement := &domain.Movement{
		Name: "Test Movement",
		Type: "strength",
	}
	repo.Create(movement)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
		errType error
	}{
		{
			name:    "existing movement",
			id:      movement.ID,
			wantErr: false,
		},
		{
			name:    "non-existent movement",
			id:      999,
			wantErr: true,
			errType: ErrMovementNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && err != tt.errType {
				t.Errorf("GetByID() error = %v, want %v", err, tt.errType)
			}
			if !tt.wantErr && result == nil {
				t.Error("GetByID() returned nil for existing movement")
			}
		})
	}
}

func TestMovementService_GetByName(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create a movement
	movement := &domain.Movement{
		Name: "Bench Press",
		Type: "strength",
	}
	repo.Create(movement)

	tests := []struct {
		name    string
		movName string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "existing movement",
			movName: "Bench Press",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "non-existent movement",
			movName: "Does Not Exist",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "empty name",
			movName: "",
			wantErr: true,
		},
		{
			name:    "whitespace name",
			movName: "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetByName(tt.movName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (result == nil) != tt.wantNil {
				t.Errorf("GetByName() result nil = %v, wantNil %v", result == nil, tt.wantNil)
			}
		})
	}
}

func TestMovementService_Update(t *testing.T) {
	repo := newTestMovementRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewMovementService(repo, nil, auditRepo)

	// Create a custom movement
	customMovement := &domain.Movement{
		Name:       "Custom Movement",
		Type:       "strength",
		IsStandard: false,
	}
	repo.Create(customMovement)

	// Create a standard movement
	standardMovement := &domain.Movement{
		Name:       "Standard Movement",
		Type:       "cardio",
		IsStandard: true,
	}
	repo.Create(standardMovement)

	tests := []struct {
		name      string
		movement  *domain.Movement
		userID    int64
		userEmail string
		wantErr   bool
		errType   error
	}{
		{
			name: "update custom movement",
			movement: &domain.Movement{
				ID:   customMovement.ID,
				Name: "Updated Custom Movement",
				Type: "strength",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   false,
		},
		{
			name: "update standard movement (unauthorized)",
			movement: &domain.Movement{
				ID:   standardMovement.ID,
				Name: "Updated Standard Movement",
				Type: "cardio",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errType:   ErrMovementUnauthorized,
		},
		{
			name: "update non-existent movement",
			movement: &domain.Movement{
				ID:   999,
				Name: "Does Not Exist",
				Type: "strength",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errType:   ErrMovementNotFound,
		},
		{
			name: "update with empty name",
			movement: &domain.Movement{
				ID:   customMovement.ID,
				Name: "",
				Type: "strength",
			},
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errType:   ErrMovementNameRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Update(tt.movement, tt.userID, tt.userEmail)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && err != tt.errType {
				t.Errorf("Update() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestMovementService_UpdateAsAdmin(t *testing.T) {
	repo := newTestMovementRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewMovementService(repo, nil, auditRepo)

	// Create a standard movement
	standardMovement := &domain.Movement{
		Name:       "Standard Movement",
		Type:       "cardio",
		IsStandard: true,
	}
	repo.Create(standardMovement)

	// Admin should be able to update standard movements
	updated := &domain.Movement{
		ID:   standardMovement.ID,
		Name: "Updated Standard Movement",
		Type: "cardio",
	}

	err := service.UpdateAsAdmin(updated, 1, "admin@example.com")
	if err != nil {
		t.Errorf("UpdateAsAdmin() error = %v, want nil", err)
	}

	// Verify the update
	result, _ := service.GetByID(standardMovement.ID)
	if result.Name != "Updated Standard Movement" {
		t.Errorf("UpdateAsAdmin() name = %v, want %v", result.Name, "Updated Standard Movement")
	}
}

func TestMovementService_Delete(t *testing.T) {
	repo := newTestMovementRepo()
	auditRepo := newMockAuditLogRepo()
	service := NewMovementService(repo, nil, auditRepo)

	// Create a custom movement
	customMovement := &domain.Movement{
		Name:       "Custom Movement",
		Type:       "strength",
		IsStandard: false,
	}
	repo.Create(customMovement)

	// Create a standard movement
	standardMovement := &domain.Movement{
		Name:       "Standard Movement",
		Type:       "cardio",
		IsStandard: true,
	}
	repo.Create(standardMovement)

	tests := []struct {
		name      string
		id        int64
		userID    int64
		userEmail string
		wantErr   bool
		errType   error
	}{
		{
			name:      "delete custom movement",
			id:        customMovement.ID,
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   false,
		},
		{
			name:      "delete standard movement (unauthorized)",
			id:        standardMovement.ID,
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errType:   ErrMovementUnauthorized,
		},
		{
			name:      "delete non-existent movement",
			id:        999,
			userID:    1,
			userEmail: "user@example.com",
			wantErr:   true,
			errType:   ErrMovementNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Delete(tt.id, tt.userID, tt.userEmail)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && err != tt.errType {
				t.Errorf("Delete() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestMovementService_Search(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create some movements
	movements := []*domain.Movement{
		{Name: "Back Squat", Type: "strength"},
		{Name: "Front Squat", Type: "strength"},
		{Name: "Overhead Squat", Type: "strength"},
		{Name: "Deadlift", Type: "strength"},
		{Name: "Pull-up", Type: "gymnastics"},
	}

	for _, m := range movements {
		repo.Create(m)
	}

	tests := []struct {
		name      string
		query     string
		limit     int
		wantCount int
	}{
		{
			name:      "search squat",
			query:     "Squat",
			limit:     10,
			wantCount: 3,
		},
		{
			name:      "search deadlift",
			query:     "Deadlift",
			limit:     10,
			wantCount: 1,
		},
		{
			name:      "search with limit",
			query:     "Squat",
			limit:     2,
			wantCount: 2,
		},
		{
			name:      "empty query",
			query:     "",
			limit:     10,
			wantCount: 0,
		},
		{
			name:      "no matches",
			query:     "ZZZZZ",
			limit:     10,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.Search(tt.query, tt.limit)
			if err != nil {
				t.Errorf("Search() error = %v", err)
				return
			}
			if len(results) != tt.wantCount {
				t.Errorf("Search() returned %d results, want %d", len(results), tt.wantCount)
			}
		})
	}
}

func TestMovementService_ListAll(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create some movements
	movements := []*domain.Movement{
		{Name: "Movement 1", Type: "strength", IsStandard: true},
		{Name: "Movement 2", Type: "cardio", IsStandard: false},
		{Name: "Movement 3", Type: "gymnastics", IsStandard: true},
	}

	for _, m := range movements {
		repo.Create(m)
	}

	results, err := service.ListAll()
	if err != nil {
		t.Errorf("ListAll() error = %v", err)
		return
	}

	if len(results) != 3 {
		t.Errorf("ListAll() returned %d results, want 3", len(results))
	}
}

func TestMovementService_ListStandard(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create some movements
	movements := []*domain.Movement{
		{Name: "Movement 1", Type: "strength", IsStandard: true},
		{Name: "Movement 2", Type: "cardio", IsStandard: false},
		{Name: "Movement 3", Type: "gymnastics", IsStandard: true},
	}

	for _, m := range movements {
		repo.Create(m)
	}

	results, err := service.ListStandard()
	if err != nil {
		t.Errorf("ListStandard() error = %v", err)
		return
	}

	if len(results) != 2 {
		t.Errorf("ListStandard() returned %d results, want 2", len(results))
	}

	for _, r := range results {
		if !r.IsStandard {
			t.Errorf("ListStandard() returned non-standard movement: %s", r.Name)
		}
	}
}

func TestMovementService_CopyToStandard(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create a custom movement
	customMovement := &domain.Movement{
		Name:        "Custom Back Squat",
		Type:        "strength",
		Description: "Custom squat variation",
		IsStandard:  false,
	}
	repo.Create(customMovement)

	tests := []struct {
		name    string
		id      int64
		newName string
		wantErr bool
	}{
		{
			name:    "copy to standard",
			id:      customMovement.ID,
			newName: "Standard Back Squat",
			wantErr: false,
		},
		{
			name:    "empty new name",
			id:      customMovement.ID,
			newName: "",
			wantErr: true,
		},
		{
			name:    "whitespace new name",
			id:      customMovement.ID,
			newName: "   ",
			wantErr: true,
		},
		{
			name:    "non-existent movement",
			id:      999,
			newName: "New Standard Movement",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CopyToStandard(tt.id, tt.newName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CopyToStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result == nil {
					t.Error("CopyToStandard() returned nil")
					return
				}
				if result.Name != tt.newName {
					t.Errorf("CopyToStandard() name = %v, want %v", result.Name, tt.newName)
				}
				if !result.IsStandard {
					t.Error("CopyToStandard() did not set IsStandard to true")
				}
			}
		})
	}
}

func TestMovementService_CopyToStandard_DuplicateName(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create a standard movement with the target name
	existingStandard := &domain.Movement{
		Name:       "Existing Standard",
		Type:       "strength",
		IsStandard: true,
	}
	repo.Create(existingStandard)

	// Create a custom movement to copy
	customMovement := &domain.Movement{
		Name:       "Custom Movement",
		Type:       "strength",
		IsStandard: false,
	}
	repo.Create(customMovement)

	// Try to copy with the same name as existing standard
	_, err := service.CopyToStandard(customMovement.ID, "Existing Standard")
	if err == nil {
		t.Error("CopyToStandard() should fail when standard movement with same name exists")
	}
}

func TestMovementService_ListAllUserCreated(t *testing.T) {
	repo := newTestMovementRepo()
	service := NewMovementService(repo, nil, nil)

	// Create some movements
	userID := int64(1)
	movements := []*domain.Movement{
		{Name: "Standard 1", Type: "strength", IsStandard: true},
		{Name: "Custom 1", Type: "strength", IsStandard: false, CreatedBy: &userID},
		{Name: "Custom 2", Type: "cardio", IsStandard: false, CreatedBy: &userID},
		{Name: "Standard 2", Type: "gymnastics", IsStandard: true},
	}

	for _, m := range movements {
		repo.Create(m)
	}

	results, count, err := service.ListAllUserCreated()
	if err != nil {
		t.Errorf("ListAllUserCreated() error = %v", err)
		return
	}

	if len(results) != 2 {
		t.Errorf("ListAllUserCreated() returned %d results, want 2", len(results))
	}

	if count != 2 {
		t.Errorf("ListAllUserCreated() count = %d, want 2", count)
	}

	for _, r := range results {
		if r.IsStandard {
			t.Errorf("ListAllUserCreated() returned standard movement: %s", r.Name)
		}
	}
}
