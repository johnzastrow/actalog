package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// testWorkoutRepo is a mock for WorkoutRepository
type testWorkoutRepo struct {
	workouts    map[int64]*domain.Workout
	nextID      int64
	createError error
	updateError error
	deleteError error
	getError    error
}

func newTestWorkoutRepo() *testWorkoutRepo {
	return &testWorkoutRepo{
		workouts: make(map[int64]*domain.Workout),
		nextID:   1,
	}
}

func (m *testWorkoutRepo) Create(workout *domain.Workout) error {
	if m.createError != nil {
		return m.createError
	}
	workout.ID = m.nextID
	m.nextID++
	workout.CreatedAt = time.Now()
	workout.UpdatedAt = time.Now()
	m.workouts[workout.ID] = workout
	return nil
}

func (m *testWorkoutRepo) GetByID(id int64) (*domain.Workout, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	w, ok := m.workouts[id]
	if !ok {
		return nil, nil
	}
	return w, nil
}

func (m *testWorkoutRepo) GetByIDWithDetails(id int64) (*domain.Workout, error) {
	return m.GetByID(id)
}

func (m *testWorkoutRepo) ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy != nil && *w.CreatedBy == userID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *testWorkoutRepo) ListStandard(limit, offset int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy == nil {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *testWorkoutRepo) Update(workout *domain.Workout) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.workouts[workout.ID]; !ok {
		return sql.ErrNoRows
	}
	workout.UpdatedAt = time.Now()
	m.workouts[workout.ID] = workout
	return nil
}

func (m *testWorkoutRepo) Delete(id int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, ok := m.workouts[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.workouts, id)
	return nil
}

func (m *testWorkoutRepo) GetUsageCount(templateID int64) (int, error) {
	return 0, nil
}

func (m *testWorkoutRepo) Count(userID *int64) (int64, error) {
	count := int64(0)
	for _, w := range m.workouts {
		if userID == nil {
			if w.CreatedBy == nil {
				count++
			}
		} else if w.CreatedBy != nil && *w.CreatedBy == *userID {
			count++
		}
	}
	return count, nil
}

func (m *testWorkoutRepo) List(filters map[string]interface{}, limit, offset int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	for _, w := range m.workouts {
		result = append(result, w)
	}
	return result, nil
}

func (m *testWorkoutRepo) Search(query string, limit int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	queryLower := strings.ToLower(query)
	for _, w := range m.workouts {
		if strings.Contains(strings.ToLower(w.Name), queryLower) {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *testWorkoutRepo) GetUsageStats(workoutID int64) (*domain.WorkoutWithUsageStats, error) {
	w, err := m.GetByID(workoutID)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, nil
	}
	return &domain.WorkoutWithUsageStats{
		Workout:    *w,
		TimesUsed:  5,
		LastUsedAt: nil,
	}, nil
}

func (m *testWorkoutRepo) ListAllUserCreated(limit, offset int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy != nil {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *testWorkoutRepo) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WorkoutWithCreator, error) {
	return []*domain.WorkoutWithCreator{}, nil
}

func (m *testWorkoutRepo) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, creator string) ([]*domain.WorkoutWithCreator, int64, error) {
	return []*domain.WorkoutWithCreator{}, 0, nil
}

func (m *testWorkoutRepo) CountAllUserCreated() (int64, error) {
	count := int64(0)
	for _, w := range m.workouts {
		if w.CreatedBy != nil {
			count++
		}
	}
	return count, nil
}

func (m *testWorkoutRepo) CopyToStandard(id int64, newName string) (*domain.Workout, error) {
	w, _ := m.GetByID(id)
	if w == nil {
		return nil, sql.ErrNoRows
	}
	newWorkout := &domain.Workout{
		Name:      newName,
		Notes:     w.Notes,
		CreatedBy: nil,
	}
	m.Create(newWorkout)
	return newWorkout, nil
}

// testWorkoutMovementRepo is a mock for WorkoutMovementRepository
type testWorkoutMovementRepo struct {
	movements   map[int64]*domain.WorkoutMovement
	nextID      int64
	createError error
}

func newTestWorkoutMovementRepo() *testWorkoutMovementRepo {
	return &testWorkoutMovementRepo{
		movements: make(map[int64]*domain.WorkoutMovement),
		nextID:    1,
	}
}

func (m *testWorkoutMovementRepo) Create(wm *domain.WorkoutMovement) error {
	if m.createError != nil {
		return m.createError
	}
	wm.ID = m.nextID
	m.nextID++
	m.movements[wm.ID] = wm
	return nil
}

func (m *testWorkoutMovementRepo) GetByID(id int64) (*domain.WorkoutMovement, error) {
	wm, ok := m.movements[id]
	if !ok {
		return nil, nil
	}
	return wm, nil
}

func (m *testWorkoutMovementRepo) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutMovement, error) {
	var result []*domain.WorkoutMovement
	for _, wm := range m.movements {
		if wm.WorkoutID == workoutID {
			result = append(result, wm)
		}
	}
	return result, nil
}

func (m *testWorkoutMovementRepo) GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*domain.WorkoutMovement, error) {
	return []*domain.WorkoutMovement{}, nil
}

func (m *testWorkoutMovementRepo) Update(wm *domain.WorkoutMovement) error {
	if _, ok := m.movements[wm.ID]; !ok {
		return sql.ErrNoRows
	}
	m.movements[wm.ID] = wm
	return nil
}

func (m *testWorkoutMovementRepo) Delete(id int64) error {
	delete(m.movements, id)
	return nil
}

func (m *testWorkoutMovementRepo) DeleteByWorkoutID(workoutID int64) error {
	for id, wm := range m.movements {
		if wm.WorkoutID == workoutID {
			delete(m.movements, id)
		}
	}
	return nil
}

func (m *testWorkoutMovementRepo) GetPersonalRecords(userID int64) ([]*domain.PersonalRecord, error) {
	return []*domain.PersonalRecord{}, nil
}

func (m *testWorkoutMovementRepo) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	var maxWeight *float64
	for _, wm := range m.movements {
		if wm.MovementID == movementID && wm.Weight != nil {
			if maxWeight == nil || *wm.Weight > *maxWeight {
				maxWeight = wm.Weight
			}
		}
	}
	return maxWeight, nil
}

func (m *testWorkoutMovementRepo) GetPRMovements(userID int64, limit int) ([]*domain.WorkoutMovement, error) {
	var result []*domain.WorkoutMovement
	for _, wm := range m.movements {
		if wm.IsPR {
			result = append(result, wm)
		}
	}
	return result, nil
}

func (m *testWorkoutMovementRepo) ListByWorkout(workoutID int64) ([]*domain.WorkoutMovement, error) {
	return m.GetByWorkoutID(workoutID)
}

func (m *testWorkoutMovementRepo) DeleteByWorkout(workoutID int64) error {
	return m.DeleteByWorkoutID(workoutID)
}

// testWorkoutWODRepo is a mock for WorkoutWODRepository
type testWorkoutWODRepo struct {
	wods        map[int64]*domain.WorkoutWOD
	nextID      int64
	createError error
}

func newTestWorkoutWODRepo() *testWorkoutWODRepo {
	return &testWorkoutWODRepo{
		wods:   make(map[int64]*domain.WorkoutWOD),
		nextID: 1,
	}
}

func (m *testWorkoutWODRepo) Create(wod *domain.WorkoutWOD) error {
	if m.createError != nil {
		return m.createError
	}
	wod.ID = m.nextID
	m.nextID++
	m.wods[wod.ID] = wod
	return nil
}

func (m *testWorkoutWODRepo) GetByID(id int64) (*domain.WorkoutWOD, error) {
	wod, ok := m.wods[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return wod, nil
}

func (m *testWorkoutWODRepo) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutWOD, error) {
	var result []*domain.WorkoutWOD
	for _, wod := range m.wods {
		if wod.WorkoutID == workoutID {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *testWorkoutWODRepo) Update(wod *domain.WorkoutWOD) error {
	if _, ok := m.wods[wod.ID]; !ok {
		return sql.ErrNoRows
	}
	m.wods[wod.ID] = wod
	return nil
}

func (m *testWorkoutWODRepo) Delete(id int64) error {
	delete(m.wods, id)
	return nil
}

func (m *testWorkoutWODRepo) DeleteByWorkout(workoutID int64) error {
	for id, wod := range m.wods {
		if wod.WorkoutID == workoutID {
			delete(m.wods, id)
		}
	}
	return nil
}

func (m *testWorkoutWODRepo) ListByWorkout(workoutID int64) ([]*domain.WorkoutWOD, error) {
	return m.GetByWorkoutID(workoutID)
}

func (m *testWorkoutWODRepo) ListByWorkoutWithDetails(workoutID int64) ([]*domain.WorkoutWODWithDetails, error) {
	var result []*domain.WorkoutWODWithDetails
	for _, wod := range m.wods {
		if wod.WorkoutID == workoutID {
			result = append(result, &domain.WorkoutWODWithDetails{
				WorkoutWOD:   *wod,
				WODName:      "Test WOD",
				WODScoreType: "time",
			})
		}
	}
	return result, nil
}

func (m *testWorkoutWODRepo) TogglePR(id int64) error {
	wod, ok := m.wods[id]
	if !ok {
		return sql.ErrNoRows
	}
	wod.IsPR = !wod.IsPR
	return nil
}

// testMovementRepoForWorkout is a mock for MovementRepository used by WorkoutService
type testMovementRepoForWorkout struct {
	movements map[int64]*domain.Movement
	nextID    int64
}

func newTestMovementRepoForWorkout() *testMovementRepoForWorkout {
	return &testMovementRepoForWorkout{
		movements: make(map[int64]*domain.Movement),
		nextID:    1,
	}
}

func (m *testMovementRepoForWorkout) Create(movement *domain.Movement) error {
	movement.ID = m.nextID
	m.nextID++
	m.movements[movement.ID] = movement
	return nil
}

func (m *testMovementRepoForWorkout) GetByID(id int64) (*domain.Movement, error) {
	mov, ok := m.movements[id]
	if !ok {
		return nil, nil
	}
	return mov, nil
}

func (m *testMovementRepoForWorkout) GetByName(name string) (*domain.Movement, error) {
	for _, mov := range m.movements {
		if strings.EqualFold(mov.Name, name) {
			return mov, nil
		}
	}
	return nil, nil
}

func (m *testMovementRepoForWorkout) ListAll() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, mov := range m.movements {
		result = append(result, mov)
	}
	return result, nil
}

func (m *testMovementRepoForWorkout) ListStandard() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, mov := range m.movements {
		if mov.IsStandard {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *testMovementRepoForWorkout) ListByUser(userID int64) ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, mov := range m.movements {
		if mov.CreatedBy != nil && *mov.CreatedBy == userID {
			result = append(result, mov)
		}
	}
	return result, nil
}

func (m *testMovementRepoForWorkout) ListAllUserCreated() ([]*domain.Movement, error) {
	return []*domain.Movement{}, nil
}

func (m *testMovementRepoForWorkout) ListAllUserCreatedWithUserInfo() ([]*domain.MovementWithCreator, error) {
	return []*domain.MovementWithCreator{}, nil
}

func (m *testMovementRepoForWorkout) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, movementType, creator string) ([]*domain.MovementWithCreator, int64, error) {
	return []*domain.MovementWithCreator{}, 0, nil
}

func (m *testMovementRepoForWorkout) CountAllUserCreated() (int64, error) {
	return 0, nil
}

func (m *testMovementRepoForWorkout) Update(movement *domain.Movement) error {
	m.movements[movement.ID] = movement
	return nil
}

func (m *testMovementRepoForWorkout) Delete(id int64) error {
	delete(m.movements, id)
	return nil
}

func (m *testMovementRepoForWorkout) Search(query string, limit int) ([]*domain.Movement, error) {
	return []*domain.Movement{}, nil
}

func (m *testMovementRepoForWorkout) CopyToStandard(id int64, newName string) (*domain.Movement, error) {
	return nil, nil
}

// Helper to create test workout service
func newTestWorkoutService() (*WorkoutService, *testWorkoutRepo, *testWorkoutMovementRepo, *testWorkoutWODRepo, *testMovementRepoForWorkout) {
	workoutRepo := newTestWorkoutRepo()
	wmRepo := newTestWorkoutMovementRepo()
	wwRepo := newTestWorkoutWODRepo()
	movRepo := newTestMovementRepoForWorkout()
	return NewWorkoutService(workoutRepo, wmRepo, wwRepo, movRepo), workoutRepo, wmRepo, wwRepo, movRepo
}

// TestWorkoutService_CreateTemplate tests creating workout templates
func TestWorkoutService_CreateTemplate(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		workout     *domain.Workout
		setupRepo   func(*testWorkoutRepo, *testWorkoutMovementRepo, *testWorkoutWODRepo)
		wantErr     bool
		errContains string
	}{
		{
			name:   "create simple template",
			userID: 1,
			workout: &domain.Workout{
				Name:  "Morning Workout",
				Notes: stringPtr("Full body routine"),
			},
			wantErr: false,
		},
		{
			name:   "create template with movements",
			userID: 1,
			workout: &domain.Workout{
				Name: "Strength Day",
				Movements: []*domain.WorkoutMovement{
					{MovementID: 1, Sets: intPtr(3), Reps: intPtr(10)},
					{MovementID: 2, Sets: intPtr(4), Reps: intPtr(8)},
				},
			},
			wantErr: false,
		},
		{
			name:   "create template with WODs",
			userID: 1,
			workout: &domain.Workout{
				Name: "WOD Day",
				WODs: []*domain.WorkoutWODWithDetails{
					{WorkoutWOD: domain.WorkoutWOD{WODID: 1}},
					{WorkoutWOD: domain.WorkoutWOD{WODID: 2}},
				},
			},
			wantErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			workout: &domain.Workout{
				Name: "Test",
			},
			setupRepo: func(wr *testWorkoutRepo, wmr *testWorkoutMovementRepo, wwr *testWorkoutWODRepo) {
				wr.createError = errors.New("database error")
			},
			wantErr:     true,
			errContains: "database error",
		},
		{
			name:   "movement create error",
			userID: 1,
			workout: &domain.Workout{
				Name: "Test",
				Movements: []*domain.WorkoutMovement{
					{MovementID: 1},
				},
			},
			setupRepo: func(wr *testWorkoutRepo, wmr *testWorkoutMovementRepo, wwr *testWorkoutWODRepo) {
				wmr.createError = errors.New("movement error")
			},
			wantErr:     true,
			errContains: "movement error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, wmr, wwr, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr, wmr, wwr)
			}

			err := svc.CreateTemplate(tt.userID, tt.workout)

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
				if tt.workout.ID == 0 {
					t.Error("expected workout ID to be set")
				}
				if tt.workout.CreatedBy == nil || *tt.workout.CreatedBy != tt.userID {
					t.Error("expected CreatedBy to be set to userID")
				}
			}
		})
	}
}

// TestWorkoutService_GetTemplate tests retrieving workout templates
func TestWorkoutService_GetTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		setupRepo   func(*testWorkoutRepo)
		wantErr     bool
		errContains string
	}{
		{
			name:       "existing template",
			templateID: 1,
			setupRepo: func(r *testWorkoutRepo) {
				r.workouts[1] = &domain.Workout{
					ID:   1,
					Name: "Test Workout",
				}
			},
			wantErr: false,
		},
		{
			name:        "non-existent template",
			templateID:  999,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:       "repository error",
			templateID: 1,
			setupRepo: func(r *testWorkoutRepo) {
				r.getError = errors.New("database error")
			},
			wantErr:     true,
			errContains: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, _, _, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr)
			}

			workout, err := svc.GetTemplate(tt.templateID)

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
				if workout == nil {
					t.Error("expected non-nil workout")
				}
			}
		})
	}
}

// TestWorkoutService_ListTemplates tests listing workout templates
func TestWorkoutService_ListTemplates(t *testing.T) {
	tests := []struct {
		name      string
		userID    *int64
		setupRepo func(*testWorkoutRepo)
		wantCount int
	}{
		{
			name:   "list standard only (no user)",
			userID: nil,
			setupRepo: func(r *testWorkoutRepo) {
				r.workouts[1] = &domain.Workout{ID: 1, Name: "Standard 1", CreatedBy: nil}
				r.workouts[2] = &domain.Workout{ID: 2, Name: "Standard 2", CreatedBy: nil}
				userID := int64(1)
				r.workouts[3] = &domain.Workout{ID: 3, Name: "User", CreatedBy: &userID}
			},
			wantCount: 2,
		},
		{
			name:   "list standard plus user templates",
			userID: int64Ptr(1),
			setupRepo: func(r *testWorkoutRepo) {
				r.workouts[1] = &domain.Workout{ID: 1, Name: "Standard", CreatedBy: nil}
				userID := int64(1)
				r.workouts[2] = &domain.Workout{ID: 2, Name: "User", CreatedBy: &userID}
				otherUser := int64(2)
				r.workouts[3] = &domain.Workout{ID: 3, Name: "Other", CreatedBy: &otherUser}
			},
			wantCount: 2, // 1 standard + 1 user's own
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, _, _, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr)
			}

			templates, err := svc.ListTemplates(tt.userID, 100, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(templates) != tt.wantCount {
				t.Errorf("expected %d templates, got %d", tt.wantCount, len(templates))
			}
		})
	}
}

// TestWorkoutService_UpdateTemplate tests updating workout templates
func TestWorkoutService_UpdateTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		userID      int64
		updates     *domain.Workout
		setupRepo   func(*testWorkoutRepo)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful update",
			templateID: 1,
			userID:     1,
			updates: &domain.Workout{
				Name:  "Updated Name",
				Notes: stringPtr("Updated notes"),
			},
			setupRepo: func(r *testWorkoutRepo) {
				userID := int64(1)
				r.workouts[1] = &domain.Workout{
					ID:        1,
					Name:      "Original",
					CreatedBy: &userID,
				}
			},
			wantErr: false,
		},
		{
			name:       "unauthorized - different user",
			templateID: 1,
			userID:     2,
			updates:    &domain.Workout{Name: "Updated"},
			setupRepo: func(r *testWorkoutRepo) {
				userID := int64(1)
				r.workouts[1] = &domain.Workout{
					ID:        1,
					Name:      "Original",
					CreatedBy: &userID,
				}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},
		{
			name:       "unauthorized - standard template",
			templateID: 1,
			userID:     1,
			updates:    &domain.Workout{Name: "Updated"},
			setupRepo: func(r *testWorkoutRepo) {
				r.workouts[1] = &domain.Workout{
					ID:        1,
					Name:      "Standard",
					CreatedBy: nil, // Standard template
				}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},
		{
			name:        "template not found",
			templateID:  999,
			userID:      1,
			updates:     &domain.Workout{Name: "Updated"},
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, _, _, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr)
			}

			err := svc.UpdateTemplate(tt.templateID, tt.userID, tt.updates)

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
				// Verify update applied
				updated := wr.workouts[tt.templateID]
				if updated.Name != tt.updates.Name {
					t.Errorf("expected name %q, got %q", tt.updates.Name, updated.Name)
				}
			}
		})
	}
}

// TestWorkoutService_DeleteTemplate tests deleting workout templates
func TestWorkoutService_DeleteTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		userID      int64
		setupRepo   func(*testWorkoutRepo)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful delete",
			templateID: 1,
			userID:     1,
			setupRepo: func(r *testWorkoutRepo) {
				userID := int64(1)
				r.workouts[1] = &domain.Workout{
					ID:        1,
					Name:      "To Delete",
					CreatedBy: &userID,
				}
			},
			wantErr: false,
		},
		{
			name:       "unauthorized - different user",
			templateID: 1,
			userID:     2,
			setupRepo: func(r *testWorkoutRepo) {
				userID := int64(1)
				r.workouts[1] = &domain.Workout{
					ID:        1,
					Name:      "Original",
					CreatedBy: &userID,
				}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},
		{
			name:        "template not found",
			templateID:  999,
			userID:      1,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, _, _, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr)
			}

			err := svc.DeleteTemplate(tt.templateID, tt.userID)

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
				// Verify deleted
				if _, exists := wr.workouts[tt.templateID]; exists {
					t.Error("expected template to be deleted")
				}
			}
		})
	}
}

// TestWorkoutService_GetTemplateUsageStats tests getting usage statistics
func TestWorkoutService_GetTemplateUsageStats(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		setupRepo   func(*testWorkoutRepo)
		wantErr     bool
		errContains string
	}{
		{
			name:       "existing template",
			templateID: 1,
			setupRepo: func(r *testWorkoutRepo) {
				r.workouts[1] = &domain.Workout{ID: 1, Name: "Test"}
			},
			wantErr: false,
		},
		{
			name:        "non-existent template",
			templateID:  999,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, _, _, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr)
			}

			stats, err := svc.GetTemplateUsageStats(tt.templateID)

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
				if stats == nil {
					t.Error("expected non-nil stats")
				}
			}
		})
	}
}

// TestWorkoutService_AddMovementToTemplate tests adding movements to templates
func TestWorkoutService_AddMovementToTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		movementID  int64
		userID      int64
		setupRepo   func(*testWorkoutRepo, *testMovementRepoForWorkout)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful add",
			templateID: 1,
			movementID: 1,
			userID:     1,
			setupRepo: func(wr *testWorkoutRepo, mr *testMovementRepoForWorkout) {
				userID := int64(1)
				wr.workouts[1] = &domain.Workout{ID: 1, CreatedBy: &userID}
				mr.movements[1] = &domain.Movement{ID: 1, Name: "Squat"}
			},
			wantErr: false,
		},
		{
			name:       "template not found",
			templateID: 999,
			movementID: 1,
			userID:     1,
			setupRepo: func(wr *testWorkoutRepo, mr *testMovementRepoForWorkout) {
				mr.movements[1] = &domain.Movement{ID: 1, Name: "Squat"}
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:       "unauthorized",
			templateID: 1,
			movementID: 1,
			userID:     2,
			setupRepo: func(wr *testWorkoutRepo, mr *testMovementRepoForWorkout) {
				userID := int64(1)
				wr.workouts[1] = &domain.Workout{ID: 1, CreatedBy: &userID}
				mr.movements[1] = &domain.Movement{ID: 1, Name: "Squat"}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},
		{
			name:       "movement not found",
			templateID: 1,
			movementID: 999,
			userID:     1,
			setupRepo: func(wr *testWorkoutRepo, mr *testMovementRepoForWorkout) {
				userID := int64(1)
				wr.workouts[1] = &domain.Workout{ID: 1, CreatedBy: &userID}
			},
			wantErr:     true,
			errContains: "movement not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, _, _, mr := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr, mr)
			}

			wm := &domain.WorkoutMovement{}
			err := svc.AddMovementToTemplate(tt.templateID, tt.movementID, tt.userID, wm)

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

// TestWorkoutService_AddWODToTemplate tests adding WODs to templates
func TestWorkoutService_AddWODToTemplate(t *testing.T) {
	tests := []struct {
		name        string
		templateID  int64
		wodID       int64
		userID      int64
		setupRepo   func(*testWorkoutRepo)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful add",
			templateID: 1,
			wodID:      1,
			userID:     1,
			setupRepo: func(wr *testWorkoutRepo) {
				userID := int64(1)
				wr.workouts[1] = &domain.Workout{ID: 1, CreatedBy: &userID}
			},
			wantErr: false,
		},
		{
			name:        "template not found",
			templateID:  999,
			wodID:       1,
			userID:      1,
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:       "unauthorized",
			templateID: 1,
			wodID:      1,
			userID:     2,
			setupRepo: func(wr *testWorkoutRepo) {
				userID := int64(1)
				wr.workouts[1] = &domain.Workout{ID: 1, CreatedBy: &userID}
			},
			wantErr:     true,
			errContains: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, wr, _, _, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wr)
			}

			wod := &domain.WorkoutWOD{}
			err := svc.AddWODToTemplate(tt.templateID, tt.wodID, tt.userID, wod)

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

// TestWorkoutService_ListMovements tests listing movements
func TestWorkoutService_ListMovements(t *testing.T) {
	svc, _, _, _, mr := newTestWorkoutService()

	// Add some movements
	userID := int64(1)
	mr.movements[1] = &domain.Movement{ID: 1, Name: "Standard Move", IsStandard: true}
	mr.movements[2] = &domain.Movement{ID: 2, Name: "Custom Move", IsStandard: false, CreatedBy: &userID}
	otherUser := int64(2)
	mr.movements[3] = &domain.Movement{ID: 3, Name: "Other User Move", IsStandard: false, CreatedBy: &otherUser}

	movements, err := svc.ListMovements(userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should get 1 standard + 1 user's custom = 2
	if len(movements) != 2 {
		t.Errorf("expected 2 movements, got %d", len(movements))
	}
}

// TestWorkoutService_DetectAndFlagPRs tests PR detection
func TestWorkoutService_DetectAndFlagPRs(t *testing.T) {
	svc, _, wmr, _, _ := newTestWorkoutService()

	// Add existing max weight
	weight100 := float64(100)
	wmr.movements[1] = &domain.WorkoutMovement{
		ID:         1,
		MovementID: 1,
		Weight:     &weight100,
	}

	// New movements to check
	weight150 := float64(150) // PR (exceeds 100)
	weight80 := float64(80)   // Not PR (less than 100)
	movements := []*domain.WorkoutMovement{
		{MovementID: 1, Weight: &weight150},
		{MovementID: 1, Weight: &weight80},
		{MovementID: 2, Weight: &weight100}, // New movement, should be PR
	}

	err := svc.DetectAndFlagPRs(1, movements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !movements[0].IsPR {
		t.Error("expected 150 to be flagged as PR")
	}
	if movements[1].IsPR {
		t.Error("expected 80 to NOT be flagged as PR")
	}
	if !movements[2].IsPR {
		t.Error("expected new movement to be flagged as PR")
	}
}

// TestWorkoutService_GetPersonalRecords tests getting personal records
func TestWorkoutService_GetPersonalRecords(t *testing.T) {
	svc, _, _, _, _ := newTestWorkoutService()

	records, err := svc.GetPersonalRecords(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mock returns empty slice
	if records == nil {
		t.Error("expected non-nil records")
	}
}

// TestWorkoutService_GetPRMovements tests getting PR movements
func TestWorkoutService_GetPRMovements(t *testing.T) {
	svc, _, wmr, _, _ := newTestWorkoutService()

	// Add some PR movements
	wmr.movements[1] = &domain.WorkoutMovement{ID: 1, IsPR: true}
	wmr.movements[2] = &domain.WorkoutMovement{ID: 2, IsPR: false}
	wmr.movements[3] = &domain.WorkoutMovement{ID: 3, IsPR: true}

	movements, err := svc.GetPRMovements(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(movements) != 2 {
		t.Errorf("expected 2 PR movements, got %d", len(movements))
	}
}

// TestWorkoutService_TogglePRFlag tests toggling PR flag
func TestWorkoutService_TogglePRFlag(t *testing.T) {
	tests := []struct {
		name         string
		movementID   int64
		setupRepo    func(*testWorkoutMovementRepo)
		wantErr      bool
		errContains  string
		wantPRAfter  bool
	}{
		{
			name:       "toggle false to true",
			movementID: 1,
			setupRepo: func(r *testWorkoutMovementRepo) {
				r.movements[1] = &domain.WorkoutMovement{ID: 1, IsPR: false}
			},
			wantErr:     false,
			wantPRAfter: true,
		},
		{
			name:       "toggle true to false",
			movementID: 1,
			setupRepo: func(r *testWorkoutMovementRepo) {
				r.movements[1] = &domain.WorkoutMovement{ID: 1, IsPR: true}
			},
			wantErr:     false,
			wantPRAfter: false,
		},
		{
			name:        "movement not found",
			movementID:  999,
			wantErr:     true,
			errContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, wmr, _, _ := newTestWorkoutService()
			if tt.setupRepo != nil {
				tt.setupRepo(wmr)
			}

			err := svc.TogglePRFlag(tt.movementID, 1)

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
				// Verify toggle
				wm := wmr.movements[tt.movementID]
				if wm.IsPR != tt.wantPRAfter {
					t.Errorf("expected IsPR=%v after toggle, got %v", tt.wantPRAfter, wm.IsPR)
				}
			}
		})
	}
}
