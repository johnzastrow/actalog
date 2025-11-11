package service

import (
	"database/sql"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// Mock UserWorkoutRepository
type mockUserWorkoutRepo struct {
	userWorkouts        map[int64]*domain.UserWorkout
	userWorkoutsDetails map[int64]*domain.UserWorkoutWithDetails
	nextID              int64
	getByIDError        error
	createError         error
	updateError         error
	deleteError         error
}

func newMockUserWorkoutRepo() *mockUserWorkoutRepo {
	return &mockUserWorkoutRepo{
		userWorkouts:        make(map[int64]*domain.UserWorkout),
		userWorkoutsDetails: make(map[int64]*domain.UserWorkoutWithDetails),
		nextID:              1,
	}
}

func (m *mockUserWorkoutRepo) Create(userWorkout *domain.UserWorkout) error {
	if m.createError != nil {
		return m.createError
	}
	m.nextID++
	userWorkout.ID = m.nextID
	userWorkout.CreatedAt = time.Now()
	userWorkout.UpdatedAt = time.Now()
	m.userWorkouts[userWorkout.ID] = userWorkout
	return nil
}

func (m *mockUserWorkoutRepo) GetByID(id int64) (*domain.UserWorkout, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	uw, ok := m.userWorkouts[id]
	if !ok {
		return nil, nil
	}
	return uw, nil
}

func (m *mockUserWorkoutRepo) GetByIDWithDetails(id int64, userID int64) (*domain.UserWorkoutWithDetails, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	uw, ok := m.userWorkoutsDetails[id]
	if !ok {
		basicUW, exists := m.userWorkouts[id]
		if !exists {
			return nil, nil
		}
		if basicUW.UserID != userID {
			return nil, nil
		}
		uw = &domain.UserWorkoutWithDetails{
			UserWorkout:        *basicUW,
			WorkoutName:        "Test Workout",
			WorkoutDescription: nil,
			Movements:          []*domain.WorkoutMovement{},
			WODs:               []*domain.WorkoutWODWithDetails{},
		}
		m.userWorkoutsDetails[id] = uw
	}
	if uw.UserID != userID {
		return nil, nil
	}
	return uw, nil
}

func (m *mockUserWorkoutRepo) ListByUser(userID int64, limit, offset int) ([]*domain.UserWorkout, error) {
	var result []*domain.UserWorkout
	for _, uw := range m.userWorkouts {
		if uw.UserID == userID {
			result = append(result, uw)
		}
	}
	return result, nil
}

func (m *mockUserWorkoutRepo) ListByUserWithDetails(userID int64, limit, offset int) ([]*domain.UserWorkoutWithDetails, error) {
	var result []*domain.UserWorkoutWithDetails
	for _, uw := range m.userWorkouts {
		if uw.UserID == userID {
			details := &domain.UserWorkoutWithDetails{
				UserWorkout:        *uw,
				WorkoutName:        "Test Workout",
				WorkoutDescription: nil,
				Movements:          []*domain.WorkoutMovement{},
				WODs:               []*domain.WorkoutWODWithDetails{},
			}
			result = append(result, details)
		}
	}
	return result, nil
}

func (m *mockUserWorkoutRepo) ListByUserAndDateRange(userID int64, startDate, endDate time.Time) ([]*domain.UserWorkout, error) {
	var result []*domain.UserWorkout
	for _, uw := range m.userWorkouts {
		if uw.UserID == userID && !uw.WorkoutDate.Before(startDate) && !uw.WorkoutDate.After(endDate) {
			result = append(result, uw)
		}
	}
	return result, nil
}

func (m *mockUserWorkoutRepo) Update(userWorkout *domain.UserWorkout) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.userWorkouts[userWorkout.ID]; !ok {
		return sql.ErrNoRows
	}
	userWorkout.UpdatedAt = time.Now()
	m.userWorkouts[userWorkout.ID] = userWorkout
	return nil
}

func (m *mockUserWorkoutRepo) Delete(id int64, userID int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	uw, ok := m.userWorkouts[id]
	if !ok {
		return sql.ErrNoRows
	}
	if uw.UserID != userID {
		return sql.ErrNoRows
	}
	delete(m.userWorkouts, id)
	delete(m.userWorkoutsDetails, id)
	return nil
}

func (m *mockUserWorkoutRepo) GetByUserWorkoutDate(userID, workoutID int64, date time.Time) (*domain.UserWorkout, error) {
	for _, uw := range m.userWorkouts {
		if uw.UserID == userID && uw.WorkoutID == workoutID && uw.WorkoutDate.Equal(date) {
			return uw, nil
		}
	}
	return nil, nil
}

// Mock WorkoutRepository
type mockWorkoutRepo struct {
	workouts     map[int64]*domain.Workout
	nextID       int64
	getByIDError error
}

func newMockWorkoutRepo() *mockWorkoutRepo {
	return &mockWorkoutRepo{
		workouts: make(map[int64]*domain.Workout),
		nextID:   1,
	}
}

func (m *mockWorkoutRepo) Create(workout *domain.Workout) error {
	m.nextID++
	workout.ID = m.nextID
	workout.CreatedAt = time.Now()
	workout.UpdatedAt = time.Now()
	m.workouts[workout.ID] = workout
	return nil
}

func (m *mockWorkoutRepo) GetByID(id int64) (*domain.Workout, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	w, ok := m.workouts[id]
	if !ok {
		return nil, nil
	}
	return w, nil
}

func (m *mockWorkoutRepo) ListByUser(userID int64, limit, offset int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy != nil && *w.CreatedBy == userID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *mockWorkoutRepo) Update(workout *domain.Workout) error {
	if _, ok := m.workouts[workout.ID]; !ok {
		return sql.ErrNoRows
	}
	workout.UpdatedAt = time.Now()
	m.workouts[workout.ID] = workout
	return nil
}

func (m *mockWorkoutRepo) Delete(id int64) error {
	if _, ok := m.workouts[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.workouts, id)
	return nil
}

func (m *mockWorkoutRepo) GetUsageCount(templateID int64) (int, error) {
	return 0, nil
}

func (m *mockWorkoutRepo) Count(userID *int64) (int64, error) {
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

func (m *mockWorkoutRepo) GetByIDWithDetails(id int64) (*domain.Workout, error) {
	return m.GetByID(id)
}

func (m *mockWorkoutRepo) List(filters map[string]interface{}, limit, offset int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	for _, w := range m.workouts {
		result = append(result, w)
	}
	return result, nil
}

func (m *mockWorkoutRepo) ListStandard(limit, offset int) ([]*domain.Workout, error) {
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy == nil {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *mockWorkoutRepo) Search(query string, limit int) ([]*domain.Workout, error) {
	return []*domain.Workout{}, nil
}

func (m *mockWorkoutRepo) GetUsageStats(workoutID int64) (*domain.WorkoutWithUsageStats, error) {
	w, err := m.GetByID(workoutID)
	if err != nil {
		return nil, err
	}
	return &domain.WorkoutWithUsageStats{
		Workout:    *w,
		TimesUsed:  0,
		LastUsedAt: nil,
	}, nil
}

// Mock WorkoutMovementRepository
type mockWorkoutMovementRepo struct{}

func (m *mockWorkoutMovementRepo) Create(workoutMovement *domain.WorkoutMovement) error {
	return nil
}

func (m *mockWorkoutMovementRepo) GetByID(id int64) (*domain.WorkoutMovement, error) {
	return nil, sql.ErrNoRows
}

func (m *mockWorkoutMovementRepo) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutMovement, error) {
	return []*domain.WorkoutMovement{}, nil
}

func (m *mockWorkoutMovementRepo) GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*domain.WorkoutMovement, error) {
	return []*domain.WorkoutMovement{}, nil
}

func (m *mockWorkoutMovementRepo) Update(wm *domain.WorkoutMovement) error {
	return nil
}

func (m *mockWorkoutMovementRepo) Delete(id int64) error {
	return nil
}

func (m *mockWorkoutMovementRepo) DeleteByWorkoutID(workoutID int64) error {
	return nil
}

func (m *mockWorkoutMovementRepo) GetPersonalRecords(userID int64) ([]*domain.PersonalRecord, error) {
	return []*domain.PersonalRecord{}, nil
}

func (m *mockWorkoutMovementRepo) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	return nil, nil
}

func (m *mockWorkoutMovementRepo) GetPRMovements(userID int64, limit int) ([]*domain.WorkoutMovement, error) {
	return []*domain.WorkoutMovement{}, nil
}

func (m *mockWorkoutMovementRepo) ListByWorkout(workoutID int64) ([]*domain.WorkoutMovement, error) {
	return []*domain.WorkoutMovement{}, nil
}

func (m *mockWorkoutMovementRepo) DeleteByWorkout(workoutID int64) error {
	return nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

// Mock WODRepository
type mockWODRepo struct {
	wods         map[int64]*domain.WOD
	nextID       int64
	getByIDError error
	createError  error
	updateError  error
	deleteError  error
}

func newMockWODRepo() *mockWODRepo {
	return &mockWODRepo{
		wods:   make(map[int64]*domain.WOD),
		nextID: 1,
	}
}

func (m *mockWODRepo) Create(wod *domain.WOD) error {
	if m.createError != nil {
		return m.createError
	}
	m.nextID++
	wod.ID = m.nextID
	m.wods[wod.ID] = wod
	return nil
}

func (m *mockWODRepo) GetByID(id int64) (*domain.WOD, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	wod, ok := m.wods[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return wod, nil
}

func (m *mockWODRepo) GetByName(name string) (*domain.WOD, error) {
	for _, wod := range m.wods {
		if wod.Name == name {
			return wod, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockWODRepo) List(filters map[string]interface{}) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		result = append(result, wod)
	}
	return result, nil
}

func (m *mockWODRepo) ListStandard() ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		if wod.IsStandard {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *mockWODRepo) ListByUser(userID int64) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		if wod.CreatedBy != nil && *wod.CreatedBy == userID {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *mockWODRepo) Update(wod *domain.WOD) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.wods[wod.ID]; !ok {
		return sql.ErrNoRows
	}
	m.wods[wod.ID] = wod
	return nil
}

func (m *mockWODRepo) Delete(id int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, ok := m.wods[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.wods, id)
	return nil
}

func (m *mockWODRepo) Search(query string) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		// Simple case-insensitive substring match
		if len(query) == 0 || matchString(wod.Name, query) {
			result = append(result, wod)
		}
	}
	return result, nil
}

// Helper function for simple case-insensitive string matching
func matchString(s, substr string) bool {
	// Convert to lowercase for case-insensitive matching
	sLower := toLower(s)
	substrLower := toLower(substr)

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}
