package service

import (
	"database/sql"
	"errors"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// Mock UserWorkoutRepository
type mockUserWorkoutRepo struct {
	userWorkouts              map[int64]*domain.UserWorkout
	userWorkoutsDetails       map[int64]*domain.UserWorkoutWithDetails
	nextID                    int64
	getByIDError              error
	createError               error
	updateError               error
	deleteError               error
	listByUserAndDateRangeErr error
	listByUserWithDetailsErr  error
	getActiveUsersError       error
	activeUsersResult         []map[string]interface{}
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
	if m.listByUserWithDetailsErr != nil {
		return nil, m.listByUserWithDetailsErr
	}
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
	if m.listByUserAndDateRangeErr != nil {
		return nil, m.listByUserAndDateRangeErr
	}
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
		if uw.UserID == userID && uw.WorkoutID != nil && *uw.WorkoutID == workoutID && uw.WorkoutDate.Equal(date) {
			return uw, nil
		}
	}
	return nil, nil
}

func (m *mockUserWorkoutRepo) GetActiveUsersThisMonth(userID int64) ([]map[string]interface{}, error) {
	if m.getActiveUsersError != nil {
		return nil, m.getActiveUsersError
	}
	if m.activeUsersResult != nil {
		return m.activeUsersResult, nil
	}
	return []map[string]interface{}{}, nil
}

// Mock WorkoutRepository
type mockWorkoutRepo struct {
	workouts                        map[int64]*domain.Workout
	nextID                          int64
	getByIDError                    error
	createError                     error
	updateError                     error
	deleteError                     error
	listByUserError                 error
	listStandardError               error
	listAllUserCreatedError         error
	listUserCreatedWithInfoError    error
	listUserCreatedFilteredError    error
	countAllUserCreatedError        error
	getByIDWithDetailsError         error
}

func newMockWorkoutRepo() *mockWorkoutRepo {
	return &mockWorkoutRepo{
		workouts: make(map[int64]*domain.Workout),
		nextID:   1,
	}
}

func (m *mockWorkoutRepo) Create(workout *domain.Workout) error {
	if m.createError != nil {
		return m.createError
	}
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
	if m.listByUserError != nil {
		return nil, m.listByUserError
	}
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy != nil && *w.CreatedBy == userID {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *mockWorkoutRepo) Update(workout *domain.Workout) error {
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

func (m *mockWorkoutRepo) Delete(id int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
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
	if m.getByIDWithDetailsError != nil {
		return nil, m.getByIDWithDetailsError
	}
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
	if m.listStandardError != nil {
		return nil, m.listStandardError
	}
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
	if w == nil {
		return nil, nil
	}
	return &domain.WorkoutWithUsageStats{
		Workout:    *w,
		TimesUsed:  0,
		LastUsedAt: nil,
	}, nil
}

func (m *mockWorkoutRepo) ListAllUserCreated(limit, offset int) ([]*domain.Workout, error) {
	if m.listAllUserCreatedError != nil {
		return nil, m.listAllUserCreatedError
	}
	var result []*domain.Workout
	for _, w := range m.workouts {
		if w.CreatedBy != nil {
			result = append(result, w)
		}
	}
	return result, nil
}

func (m *mockWorkoutRepo) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WorkoutWithCreator, error) {
	if m.listUserCreatedWithInfoError != nil {
		return nil, m.listUserCreatedWithInfoError
	}
	return []*domain.WorkoutWithCreator{}, nil
}

func (m *mockWorkoutRepo) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, creator string) ([]*domain.WorkoutWithCreator, int64, error) {
	if m.listUserCreatedFilteredError != nil {
		return nil, 0, m.listUserCreatedFilteredError
	}
	return []*domain.WorkoutWithCreator{}, 0, nil
}

func (m *mockWorkoutRepo) CountAllUserCreated() (int64, error) {
	if m.countAllUserCreatedError != nil {
		return 0, m.countAllUserCreatedError
	}
	count := int64(0)
	for _, w := range m.workouts {
		if w.CreatedBy != nil {
			count++
		}
	}
	return count, nil
}

func (m *mockWorkoutRepo) CopyToStandard(id int64, newName string) (*domain.Workout, error) {
	w, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}
	if w == nil {
		return nil, sql.ErrNoRows
	}
	newWorkout := &domain.Workout{
		Name:      newName,
		Notes:     w.Notes,
		CreatedBy: nil, // Standard workout
	}
	if err := m.Create(newWorkout); err != nil {
		return nil, err
	}
	return newWorkout, nil
}

// Mock WorkoutMovementRepository
type mockWorkoutMovementRepo struct {
	movements             map[int64]*domain.WorkoutMovement
	nextID                int64
	maxWeights            map[int64]map[int64]*float64 // userID -> movementID -> maxWeight
	createError           error
	deleteByWorkoutIDError error
}

func newMockWorkoutMovementRepo() *mockWorkoutMovementRepo {
	return &mockWorkoutMovementRepo{
		movements:  make(map[int64]*domain.WorkoutMovement),
		nextID:     1,
		maxWeights: make(map[int64]map[int64]*float64),
	}
}

func (m *mockWorkoutMovementRepo) Create(workoutMovement *domain.WorkoutMovement) error {
	if m.createError != nil {
		return m.createError
	}
	workoutMovement.ID = m.nextID
	m.nextID++
	m.movements[workoutMovement.ID] = workoutMovement
	return nil
}

func (m *mockWorkoutMovementRepo) GetByID(id int64) (*domain.WorkoutMovement, error) {
	wm, ok := m.movements[id]
	if !ok {
		return nil, nil
	}
	return wm, nil
}

func (m *mockWorkoutMovementRepo) GetByWorkoutID(workoutID int64) ([]*domain.WorkoutMovement, error) {
	var result []*domain.WorkoutMovement
	for _, wm := range m.movements {
		if wm.WorkoutID == workoutID {
			result = append(result, wm)
		}
	}
	return result, nil
}

func (m *mockWorkoutMovementRepo) GetByUserIDAndMovementID(userID, movementID int64, limit int) ([]*domain.WorkoutMovement, error) {
	return []*domain.WorkoutMovement{}, nil
}

func (m *mockWorkoutMovementRepo) Update(wm *domain.WorkoutMovement) error {
	if _, ok := m.movements[wm.ID]; !ok {
		return sql.ErrNoRows
	}
	m.movements[wm.ID] = wm
	return nil
}

func (m *mockWorkoutMovementRepo) Delete(id int64) error {
	if _, ok := m.movements[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.movements, id)
	return nil
}

func (m *mockWorkoutMovementRepo) DeleteByWorkoutID(workoutID int64) error {
	if m.deleteByWorkoutIDError != nil {
		return m.deleteByWorkoutIDError
	}
	for id, wm := range m.movements {
		if wm.WorkoutID == workoutID {
			delete(m.movements, id)
		}
	}
	return nil
}

func (m *mockWorkoutMovementRepo) GetPersonalRecords(userID int64) ([]*domain.PersonalRecord, error) {
	return []*domain.PersonalRecord{}, nil
}

func (m *mockWorkoutMovementRepo) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	if userWeights, ok := m.maxWeights[userID]; ok {
		if weight, ok := userWeights[movementID]; ok {
			return weight, nil
		}
	}
	return nil, nil
}

func (m *mockWorkoutMovementRepo) GetPRMovements(userID int64, limit int) ([]*domain.WorkoutMovement, error) {
	var result []*domain.WorkoutMovement
	for _, wm := range m.movements {
		if wm.IsPR {
			result = append(result, wm)
		}
	}
	return result, nil
}

func (m *mockWorkoutMovementRepo) ListByWorkout(workoutID int64) ([]*domain.WorkoutMovement, error) {
	return m.GetByWorkoutID(workoutID)
}

func (m *mockWorkoutMovementRepo) DeleteByWorkout(workoutID int64) error {
	return m.DeleteByWorkoutID(workoutID)
}

// Mock UserWorkoutMovementRepository
type mockUserWorkoutMovementRepo struct {
	movements                        map[int64]*domain.UserWorkoutMovement
	nextID                           int64
	createError                      error
	createBatchError                 error
	deleteByUserWorkoutIDError       error
	getPRMovementsError              error
	updatePRFlagError                error
	getAllPerformancesError          error
	getByUserWorkoutIDError          error
	previousPerformances             map[int64][]*domain.UserWorkoutMovement // movementID -> performances
}

func newMockUserWorkoutMovementRepo() *mockUserWorkoutMovementRepo {
	return &mockUserWorkoutMovementRepo{
		movements:            make(map[int64]*domain.UserWorkoutMovement),
		nextID:               1,
		previousPerformances: make(map[int64][]*domain.UserWorkoutMovement),
	}
}

func (m *mockUserWorkoutMovementRepo) Create(uwm *domain.UserWorkoutMovement) error {
	if m.createError != nil {
		return m.createError
	}
	uwm.ID = m.nextID
	m.nextID++
	m.movements[uwm.ID] = uwm
	return nil
}

func (m *mockUserWorkoutMovementRepo) CreateBatch(movements []*domain.UserWorkoutMovement) error {
	if m.createBatchError != nil {
		return m.createBatchError
	}
	for _, uwm := range movements {
		uwm.ID = m.nextID
		m.nextID++
		m.movements[uwm.ID] = uwm
	}
	return nil
}

func (m *mockUserWorkoutMovementRepo) GetByID(id int64) (*domain.UserWorkoutMovement, error) {
	uwm, ok := m.movements[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return uwm, nil
}

func (m *mockUserWorkoutMovementRepo) GetByUserWorkoutID(userWorkoutID int64) ([]*domain.UserWorkoutMovement, error) {
	if m.getByUserWorkoutIDError != nil {
		return nil, m.getByUserWorkoutIDError
	}
	var result []*domain.UserWorkoutMovement
	for _, uwm := range m.movements {
		if uwm.UserWorkoutID == userWorkoutID {
			result = append(result, uwm)
		}
	}
	return result, nil
}

func (m *mockUserWorkoutMovementRepo) Update(uwm *domain.UserWorkoutMovement) error {
	if _, ok := m.movements[uwm.ID]; !ok {
		return sql.ErrNoRows
	}
	m.movements[uwm.ID] = uwm
	return nil
}

func (m *mockUserWorkoutMovementRepo) Delete(id int64) error {
	delete(m.movements, id)
	return nil
}

func (m *mockUserWorkoutMovementRepo) DeleteByUserWorkoutID(userWorkoutID int64) error {
	if m.deleteByUserWorkoutIDError != nil {
		return m.deleteByUserWorkoutIDError
	}
	for id, uwm := range m.movements {
		if uwm.UserWorkoutID == userWorkoutID {
			delete(m.movements, id)
		}
	}
	return nil
}

func (m *mockUserWorkoutMovementRepo) GetMaxWeightForMovement(userID, movementID int64) (*float64, error) {
	return nil, nil
}

func (m *mockUserWorkoutMovementRepo) GetPRMovements(userID int64, limit int) ([]*domain.UserWorkoutMovement, error) {
	if m.getPRMovementsError != nil {
		return nil, m.getPRMovementsError
	}
	result := []*domain.UserWorkoutMovement{}
	for _, uwm := range m.movements {
		if uwm.IsPR {
			result = append(result, uwm)
		}
	}
	return result, nil
}

func (m *mockUserWorkoutMovementRepo) UpdatePRFlag(id int64, isPR bool) error {
	if m.updatePRFlagError != nil {
		return m.updatePRFlagError
	}
	if uwm, ok := m.movements[id]; ok {
		uwm.IsPR = isPR
	}
	return nil
}

func (m *mockUserWorkoutMovementRepo) GetAllPerformancesForMovement(userID, movementID int64, excludeWorkoutID *int64) ([]*domain.UserWorkoutMovement, error) {
	if m.getAllPerformancesError != nil {
		return nil, m.getAllPerformancesError
	}
	if perfs, ok := m.previousPerformances[movementID]; ok {
		return perfs, nil
	}
	return []*domain.UserWorkoutMovement{}, nil
}

// Mock UserWorkoutWODRepository
type mockUserWorkoutWODRepo struct {
	wods                       map[int64]*domain.UserWorkoutWOD
	nextID                     int64
	createError                error
	createBatchError           error
	deleteByUserWorkoutIDError error
	getPRWODsError             error
	updatePRFlagError          error
	getBestTimeError           error
	getBestRoundsRepsError     error
	getByUserWorkoutIDError    error
	bestTimes                  map[int64]*int            // wodID -> bestTime
	bestRoundsReps             map[int64]struct{ rounds, reps *int }
}

func newMockUserWorkoutWODRepo() *mockUserWorkoutWODRepo {
	return &mockUserWorkoutWODRepo{
		wods:           make(map[int64]*domain.UserWorkoutWOD),
		nextID:         1,
		bestTimes:      make(map[int64]*int),
		bestRoundsReps: make(map[int64]struct{ rounds, reps *int }),
	}
}

func (m *mockUserWorkoutWODRepo) Create(uww *domain.UserWorkoutWOD) error {
	if m.createError != nil {
		return m.createError
	}
	uww.ID = m.nextID
	m.nextID++
	m.wods[uww.ID] = uww
	return nil
}

func (m *mockUserWorkoutWODRepo) CreateBatch(wods []*domain.UserWorkoutWOD) error {
	if m.createBatchError != nil {
		return m.createBatchError
	}
	for _, uww := range wods {
		uww.ID = m.nextID
		m.nextID++
		m.wods[uww.ID] = uww
	}
	return nil
}

func (m *mockUserWorkoutWODRepo) GetByID(id int64) (*domain.UserWorkoutWOD, error) {
	uww, ok := m.wods[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return uww, nil
}

func (m *mockUserWorkoutWODRepo) GetByUserWorkoutID(userWorkoutID int64) ([]*domain.UserWorkoutWOD, error) {
	if m.getByUserWorkoutIDError != nil {
		return nil, m.getByUserWorkoutIDError
	}
	var result []*domain.UserWorkoutWOD
	for _, uww := range m.wods {
		if uww.UserWorkoutID == userWorkoutID {
			result = append(result, uww)
		}
	}
	return result, nil
}

func (m *mockUserWorkoutWODRepo) Update(uww *domain.UserWorkoutWOD) error {
	if _, ok := m.wods[uww.ID]; !ok {
		return sql.ErrNoRows
	}
	m.wods[uww.ID] = uww
	return nil
}

func (m *mockUserWorkoutWODRepo) Delete(id int64) error {
	delete(m.wods, id)
	return nil
}

func (m *mockUserWorkoutWODRepo) DeleteByUserWorkoutID(userWorkoutID int64) error {
	if m.deleteByUserWorkoutIDError != nil {
		return m.deleteByUserWorkoutIDError
	}
	for id, uww := range m.wods {
		if uww.UserWorkoutID == userWorkoutID {
			delete(m.wods, id)
		}
	}
	return nil
}

func (m *mockUserWorkoutWODRepo) GetBestTimeForWOD(userID, wodID int64) (*int, error) {
	if m.getBestTimeError != nil {
		return nil, m.getBestTimeError
	}
	if t, ok := m.bestTimes[wodID]; ok {
		return t, nil
	}
	return nil, nil
}

func (m *mockUserWorkoutWODRepo) GetBestRoundsRepsForWOD(userID, wodID int64) (rounds *int, reps *int, err error) {
	if m.getBestRoundsRepsError != nil {
		return nil, nil, m.getBestRoundsRepsError
	}
	if rr, ok := m.bestRoundsReps[wodID]; ok {
		return rr.rounds, rr.reps, nil
	}
	return nil, nil, nil
}

func (m *mockUserWorkoutWODRepo) GetPRWODs(userID int64, limit int) ([]*domain.UserWorkoutWOD, error) {
	if m.getPRWODsError != nil {
		return nil, m.getPRWODsError
	}
	result := []*domain.UserWorkoutWOD{}
	for _, uww := range m.wods {
		if uww.IsPR {
			result = append(result, uww)
		}
	}
	return result, nil
}

func (m *mockUserWorkoutWODRepo) UpdatePRFlag(id int64, isPR bool) error {
	if m.updatePRFlagError != nil {
		return m.updatePRFlagError
	}
	if uww, ok := m.wods[id]; ok {
		uww.IsPR = isPR
	}
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
		return nil, nil // Service expects nil, nil for not found
	}
	return wod, nil
}

func (m *mockWODRepo) GetByName(name string) (*domain.WOD, error) {
	for _, wod := range m.wods {
		if wod.Name == name {
			return wod, nil
		}
	}
	// Return nil, nil when not found (not sql.ErrNoRows) to match repository behavior
	return nil, nil
}

func (m *mockWODRepo) List(filters map[string]interface{}, limit, offset int) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		// Apply is_standard filter if present
		if isStandard, ok := filters["is_standard"]; ok {
			if isStandardBool, ok := isStandard.(bool); ok {
				if wod.IsStandard != isStandardBool {
					continue
				}
			}
		}
		result = append(result, wod)
	}
	return result, nil
}

func (m *mockWODRepo) ListStandard(limit, offset int) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		if wod.IsStandard {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *mockWODRepo) ListByUser(userID int64, limit, offset int) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		if wod.CreatedBy != nil && *wod.CreatedBy == userID {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *mockWODRepo) ListAllUserCreated(limit, offset int) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		if wod.CreatedBy != nil {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *mockWODRepo) ListAllUserCreatedWithUserInfo(limit, offset int) ([]*domain.WODWithCreator, error) {
	return []*domain.WODWithCreator{}, nil
}

func (m *mockWODRepo) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, scoreType, creator string) ([]*domain.WODWithCreator, int64, error) {
	return []*domain.WODWithCreator{}, 0, nil
}

func (m *mockWODRepo) CountAllUserCreated() (int64, error) {
	count := int64(0)
	for _, wod := range m.wods {
		if wod.CreatedBy != nil {
			count++
		}
	}
	return count, nil
}

func (m *mockWODRepo) UpdateStandard(wod *domain.WOD) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.wods[wod.ID]; !ok {
		return sql.ErrNoRows
	}
	m.wods[wod.ID] = wod
	return nil
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

func (m *mockWODRepo) Search(query string, limit int) ([]*domain.WOD, error) {
	var result []*domain.WOD
	for _, wod := range m.wods {
		// Simple case-insensitive substring match
		if len(query) == 0 || matchString(wod.Name, query) {
			result = append(result, wod)
		}
	}
	return result, nil
}

func (m *mockWODRepo) CopyToStandard(id int64, newName string) (*domain.WOD, error) {
	wod, err := m.GetByID(id)
	if err != nil {
		return nil, err
	}
	newWOD := &domain.WOD{
		Name:       newName,
		ScoreType:  wod.ScoreType,
		IsStandard: true,
		CreatedBy:  nil,
	}
	if err := m.Create(newWOD); err != nil {
		return nil, err
	}
	return newWOD, nil
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

// Mock UserSubscriptionRepository
type mockUserSubscriptionRepo struct {
	subscriptions map[int64]*domain.UserSubscription
	nextID        int64
	createError   error
	getByIDError  error
	updateError   error
}

func newMockUserSubscriptionRepo() *mockUserSubscriptionRepo {
	return &mockUserSubscriptionRepo{
		subscriptions: make(map[int64]*domain.UserSubscription),
		nextID:        1,
	}
}

func (m *mockUserSubscriptionRepo) Create(sub *domain.UserSubscription) error {
	if m.createError != nil {
		return m.createError
	}
	sub.ID = m.nextID
	m.nextID++
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *mockUserSubscriptionRepo) GetByID(id int64) (*domain.UserSubscription, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	sub, ok := m.subscriptions[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return sub, nil
}

func (m *mockUserSubscriptionRepo) GetActiveByUserID(userID int64) (*domain.UserSubscription, error) {
	for _, sub := range m.subscriptions {
		if sub.UserID == userID && sub.Status == domain.SubscriptionStatusActive {
			return sub, nil
		}
	}
	return nil, nil
}

func (m *mockUserSubscriptionRepo) GetByUserID(userID int64) ([]*domain.UserSubscription, error) {
	var result []*domain.UserSubscription
	for _, sub := range m.subscriptions {
		if sub.UserID == userID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *mockUserSubscriptionRepo) Update(sub *domain.UserSubscription) error {
	if m.updateError != nil {
		return m.updateError
	}
	sub.UpdatedAt = time.Now()
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *mockUserSubscriptionRepo) Delete(id int64) error {
	delete(m.subscriptions, id)
	return nil
}

func (m *mockUserSubscriptionRepo) ListAll() ([]*domain.UserSubscription, error) {
	var result []*domain.UserSubscription
	for _, sub := range m.subscriptions {
		result = append(result, sub)
	}
	return result, nil
}

func (m *mockUserSubscriptionRepo) MarkAsPaid(id int64, paymentDate time.Time, adminUserID int64, durationDays *int) error {
	sub, ok := m.subscriptions[id]
	if !ok {
		return sql.ErrNoRows
	}
	sub.LastPaymentDate = &paymentDate

	// Calculate new end date
	duration := 30 * 24 * time.Hour // Default
	if sub.SubscriptionType == domain.SubscriptionTypeAnnual {
		duration = 365 * 24 * time.Hour
	}
	if durationDays != nil {
		duration = time.Duration(*durationDays) * 24 * time.Hour
	}

	newEndDate := paymentDate.Add(duration)
	sub.EndDate = &newEndDate
	nextBilling := newEndDate
	sub.NextBillingDate = &nextBilling
	sub.UpdatedAt = time.Now()

	m.subscriptions[id] = sub
	return nil
}

func (m *mockUserSubscriptionRepo) MarkAsExpired(id int64) error {
	sub, ok := m.subscriptions[id]
	if !ok {
		return sql.ErrNoRows
	}
	sub.Status = domain.SubscriptionStatusExpired
	sub.UpdatedAt = time.Now()
	m.subscriptions[id] = sub
	return nil
}

func (m *mockUserSubscriptionRepo) Cancel(id int64, reason string, adminUserID int64) error {
	sub, ok := m.subscriptions[id]
	if !ok {
		return sql.ErrNoRows
	}
	sub.Status = domain.SubscriptionStatusCancelled
	now := time.Now()
	sub.CancelledAt = &now
	sub.CancelledReason = &reason
	sub.UpdatedAt = time.Now()
	m.subscriptions[id] = sub
	return nil
}

func (m *mockUserSubscriptionRepo) ListExpiring(days int) ([]*domain.UserSubscription, error) {
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

func (m *mockUserSubscriptionRepo) ListExpired() ([]*domain.UserSubscription, error) {
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

// Mock OrganizationSubscriptionRepository
type mockOrganizationSubscriptionRepo struct {
	subscriptions map[int64]*domain.OrganizationSubscription
	nextID        int64
	createError   error
	getByIDError  error
	updateError   error
}

func newMockOrganizationSubscriptionRepo() *mockOrganizationSubscriptionRepo {
	return &mockOrganizationSubscriptionRepo{
		subscriptions: make(map[int64]*domain.OrganizationSubscription),
		nextID:        1,
	}
}

func (m *mockOrganizationSubscriptionRepo) Create(sub *domain.OrganizationSubscription) error {
	if m.createError != nil {
		return m.createError
	}
	sub.ID = m.nextID
	m.nextID++
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *mockOrganizationSubscriptionRepo) GetByID(id int64) (*domain.OrganizationSubscription, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	sub, ok := m.subscriptions[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return sub, nil
}

func (m *mockOrganizationSubscriptionRepo) GetActiveByOrganizationID(orgID int64) (*domain.OrganizationSubscription, error) {
	for _, sub := range m.subscriptions {
		if sub.OrganizationID == orgID && sub.Status == domain.SubscriptionStatusActive {
			return sub, nil
		}
	}
	return nil, nil
}

func (m *mockOrganizationSubscriptionRepo) GetByOrganizationID(orgID int64) ([]*domain.OrganizationSubscription, error) {
	var result []*domain.OrganizationSubscription
	for _, sub := range m.subscriptions {
		if sub.OrganizationID == orgID {
			result = append(result, sub)
		}
	}
	return result, nil
}

func (m *mockOrganizationSubscriptionRepo) Update(sub *domain.OrganizationSubscription) error {
	if m.updateError != nil {
		return m.updateError
	}
	sub.UpdatedAt = time.Now()
	m.subscriptions[sub.ID] = sub
	return nil
}

func (m *mockOrganizationSubscriptionRepo) Delete(id int64) error {
	delete(m.subscriptions, id)
	return nil
}

func (m *mockOrganizationSubscriptionRepo) ListAll() ([]*domain.OrganizationSubscription, error) {
	var result []*domain.OrganizationSubscription
	for _, sub := range m.subscriptions {
		result = append(result, sub)
	}
	return result, nil
}

func (m *mockOrganizationSubscriptionRepo) MarkAsPaid(id int64, paymentDate time.Time, adminUserID int64, durationDays *int) error {
	sub, ok := m.subscriptions[id]
	if !ok {
		return sql.ErrNoRows
	}
	sub.LastPaymentDate = &paymentDate

	duration := 30 * 24 * time.Hour
	if sub.SubscriptionType == domain.SubscriptionTypeAnnual {
		duration = 365 * 24 * time.Hour
	}
	if durationDays != nil {
		duration = time.Duration(*durationDays) * 24 * time.Hour
	}

	newEndDate := paymentDate.Add(duration)
	sub.EndDate = &newEndDate
	nextBilling := newEndDate
	sub.NextBillingDate = &nextBilling
	sub.UpdatedAt = time.Now()

	m.subscriptions[id] = sub
	return nil
}

func (m *mockOrganizationSubscriptionRepo) MarkAsExpired(id int64) error {
	sub, ok := m.subscriptions[id]
	if !ok {
		return sql.ErrNoRows
	}
	sub.Status = domain.SubscriptionStatusExpired
	sub.UpdatedAt = time.Now()
	m.subscriptions[id] = sub
	return nil
}

func (m *mockOrganizationSubscriptionRepo) Cancel(id int64, reason string, adminUserID int64) error {
	sub, ok := m.subscriptions[id]
	if !ok {
		return sql.ErrNoRows
	}
	sub.Status = domain.SubscriptionStatusCancelled
	now := time.Now()
	sub.CancelledAt = &now
	sub.CancelledReason = &reason
	sub.UpdatedAt = time.Now()
	m.subscriptions[id] = sub
	return nil
}

func (m *mockOrganizationSubscriptionRepo) ListExpiring(days int) ([]*domain.OrganizationSubscription, error) {
	now := time.Now()
	futureDate := now.AddDate(0, 0, days)
	var result []*domain.OrganizationSubscription
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

func (m *mockOrganizationSubscriptionRepo) ListExpired() ([]*domain.OrganizationSubscription, error) {
	now := time.Now()
	var result []*domain.OrganizationSubscription
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

// Mock SubscriptionAccessRepository
type mockSubscriptionAccessRepo struct {
	accessResults map[int64]*domain.SubscriptionAccessResult
}

func newMockSubscriptionAccessRepo() *mockSubscriptionAccessRepo {
	return &mockSubscriptionAccessRepo{
		accessResults: make(map[int64]*domain.SubscriptionAccessResult),
	}
}

func (m *mockSubscriptionAccessRepo) CheckUserAccess(userID int64) (*domain.SubscriptionAccessResult, error) {
	result, ok := m.accessResults[userID]
	if !ok {
		return &domain.SubscriptionAccessResult{
			HasAccess:        false,
			Source:           "none",
			UserSubscription: nil,
			OrgSubscriptions: nil,
		}, nil
	}
	return result, nil
}

// Mock OrganizationRepository
type mockOrganizationRepo struct {
	organizations     map[int64]*domain.Organization
	userOrganizations map[int64]map[int64]string // userID -> orgID -> role
	orgUsers          map[int64][]*domain.User   // orgID -> users
	nextID            int64
}

func newMockOrganizationRepo() *mockOrganizationRepo {
	return &mockOrganizationRepo{
		organizations:     make(map[int64]*domain.Organization),
		userOrganizations: make(map[int64]map[int64]string),
		orgUsers:          make(map[int64][]*domain.User),
		nextID:            1,
	}
}

func (m *mockOrganizationRepo) GetByID(id int64) (*domain.Organization, error) {
	org, ok := m.organizations[id]
	if !ok {
		return nil, nil // Service expects nil, nil for not found
	}
	return org, nil
}

// Mock AuditLogRepository
type mockAuditLogRepo struct {
	logs   []*domain.AuditLog
	nextID int64
}

func newMockAuditLogRepo() *mockAuditLogRepo {
	return &mockAuditLogRepo{
		logs:   make([]*domain.AuditLog, 0),
		nextID: 1,
	}
}

func (m *mockAuditLogRepo) Create(log *domain.AuditLog) error {
	log.ID = m.nextID
	m.nextID++
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockAuditLogRepo) GetByID(id int64) (*domain.AuditLog, error) {
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockAuditLogRepo) List(filters domain.AuditLogFilters, limit, offset int) ([]*domain.AuditLog, error) {
	return m.logs, nil
}

func (m *mockAuditLogRepo) GetByTargetUserID(targetUserID int64, limit, offset int) ([]*domain.AuditLog, error) {
	var filtered []*domain.AuditLog
	for _, log := range m.logs {
		if log.TargetUserID != nil && *log.TargetUserID == targetUserID {
			filtered = append(filtered, log)
		}
	}
	return filtered, nil
}

func (m *mockAuditLogRepo) GetByUserID(userID int64, limit, offset int) ([]*domain.AuditLog, error) {
	var filtered []*domain.AuditLog
	for _, log := range m.logs {
		if log.UserID != nil && *log.UserID == userID {
			filtered = append(filtered, log)
		}
	}
	return filtered, nil
}

func (m *mockAuditLogRepo) Delete(id int64) error {
	for i, log := range m.logs {
		if log.ID == id {
			m.logs = append(m.logs[:i], m.logs[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

// Mock MovementRepository
type mockMovementRepo struct {
	movements map[int64]*domain.Movement
	nextID    int64
}

func newMockMovementRepo() *mockMovementRepo {
	return &mockMovementRepo{
		movements: make(map[int64]*domain.Movement),
		nextID:    1,
	}
}

func (m *mockMovementRepo) Create(movement *domain.Movement) error {
	movement.ID = m.nextID
	m.nextID++
	m.movements[movement.ID] = movement
	return nil
}

func (m *mockMovementRepo) GetByID(id int64) (*domain.Movement, error) {
	movement, ok := m.movements[id]
	if !ok {
		return nil, nil
	}
	return movement, nil
}

func (m *mockMovementRepo) GetByName(name string) (*domain.Movement, error) {
	for _, movement := range m.movements {
		if movement.Name == name {
			return movement, nil
		}
	}
	return nil, nil
}

func (m *mockMovementRepo) ListAll() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, m := range m.movements {
		result = append(result, m)
	}
	return result, nil
}

func (m *mockMovementRepo) ListStandard() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		if movement.IsStandard {
			result = append(result, movement)
		}
	}
	return result, nil
}

func (m *mockMovementRepo) ListByUser(userID int64) ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		if movement.CreatedBy != nil && *movement.CreatedBy == userID {
			result = append(result, movement)
		}
	}
	return result, nil
}

func (m *mockMovementRepo) ListAllUserCreated() ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		if movement.CreatedBy != nil {
			result = append(result, movement)
		}
	}
	return result, nil
}

func (m *mockMovementRepo) ListAllUserCreatedWithUserInfo() ([]*domain.MovementWithCreator, error) {
	return []*domain.MovementWithCreator{}, nil
}

func (m *mockMovementRepo) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, movementType, creator string) ([]*domain.MovementWithCreator, int64, error) {
	return []*domain.MovementWithCreator{}, 0, nil
}

func (m *mockMovementRepo) CountAllUserCreated() (int64, error) {
	count := int64(0)
	for _, movement := range m.movements {
		if movement.CreatedBy != nil {
			count++
		}
	}
	return count, nil
}

func (m *mockMovementRepo) Update(movement *domain.Movement) error {
	if _, ok := m.movements[movement.ID]; !ok {
		return sql.ErrNoRows
	}
	m.movements[movement.ID] = movement
	return nil
}

func (m *mockMovementRepo) Delete(id int64) error {
	if _, ok := m.movements[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.movements, id)
	return nil
}

func (m *mockMovementRepo) Search(query string, limit int) ([]*domain.Movement, error) {
	var result []*domain.Movement
	for _, movement := range m.movements {
		if matchString(movement.Name, query) {
			result = append(result, movement)
		}
	}
	return result, nil
}

func (m *mockMovementRepo) CopyToStandard(id int64, newName string) (*domain.Movement, error) {
	movement, ok := m.movements[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	newMovement := &domain.Movement{
		Name:       newName,
		Type:       movement.Type,
		IsStandard: true,
		CreatedBy:  nil,
	}
	_ = m.Create(newMovement)
	return newMovement, nil
}

// Mock UserRepository

// Mock UserRepository (complete implementation)
type mockUserRepo struct {
	users        map[int64]*domain.User
	nextID       int64
	getByIDError error // If set, GetByID returns this error for non-existent users
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[int64]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(user *domain.User) error {
	m.nextID++
	user.ID = m.nextID
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(id int64) (*domain.User, error) {
	user, ok := m.users[id]
	if !ok {
		if m.getByIDError != nil {
			return nil, m.getByIDError
		}
		return nil, nil
	}
	return user, nil
}

func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByResetToken(token string) (*domain.User, error) {
	for _, user := range m.users {
		if user.ResetToken != nil && *user.ResetToken == token {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) GetByVerificationToken(token string) (*domain.User, error) {
	for _, user := range m.users {
		if user.VerificationToken != nil && *user.VerificationToken == token {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) Update(user *domain.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return sql.ErrNoRows
	}
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Delete(id int64) error {
	if _, ok := m.users[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) List(limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockUserRepo) CountTotal() (int, error) {
	return len(m.users), nil
}

func (m *mockUserRepo) UpdatePassword(userID int64, hashedPassword string) error {
	if user, ok := m.users[userID]; ok {
		user.PasswordHash = hashedPassword
		user.UpdatedAt = time.Now()
		return nil
	}
	return sql.ErrNoRows
}

func (m *mockUserRepo) SetRole(id int64, role string) error {
	user, ok := m.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	user.Role = role
	user.UpdatedAt = time.Now()
	m.users[id] = user
	return nil
}

func (m *mockUserRepo) Disable(id int64) error {
	user, ok := m.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	disabled := true
	user.AccountDisabled = disabled
	user.UpdatedAt = time.Now()
	m.users[id] = user
	return nil
}

func (m *mockUserRepo) Enable(id int64) error {
	user, ok := m.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	disabled := false
	user.AccountDisabled = disabled
	user.UpdatedAt = time.Now()
	m.users[id] = user
	return nil
}

func (m *mockUserRepo) GetLoginAttempts(email string) (int, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user.FailedLoginAttempts, nil
		}
	}
	return 0, nil
}

func (m *mockUserRepo) IncrementLoginAttempts(email string) error {
	for _, user := range m.users {
		if user.Email == email {
			user.FailedLoginAttempts++
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockUserRepo) ResetLoginAttempts(email string) error {
	for _, user := range m.users {
		if user.Email == email {
			user.FailedLoginAttempts = 0
			return nil
		}
	}
	return nil
}

func (m *mockUserRepo) LockAccount(userID int64, duration time.Duration) error {
	user, ok := m.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	until := time.Now().Add(duration)
	user.LockedUntil = &until
	m.users[userID] = user
	return nil
}

func (m *mockUserRepo) IsAccountLocked(userID int64) (bool, *time.Time, error) {
	user, ok := m.users[userID]
	if !ok {
		return false, nil, sql.ErrNoRows
	}
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return true, user.LockedUntil, nil
	}
	return false, nil, nil
}

func (m *mockUserRepo) UpdateProfileImage(id int64, imagePath string) error {
	user, ok := m.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	user.ProfileImage = &imagePath
	user.UpdatedAt = time.Now()
	m.users[id] = user
	return nil
}

func (m *mockUserRepo) DeleteProfileImage(id int64) error {
	user, ok := m.users[id]
	if !ok {
		return sql.ErrNoRows
	}
	user.ProfileImage = nil
	user.UpdatedAt = time.Now()
	m.users[id] = user
	return nil
}

// Add Count method to mockUserRepo
func (m *mockUserRepo) Count() (int64, error) {
	return int64(len(m.users)), nil
}

// Add Count method to mockAuditLogRepo
func (m *mockAuditLogRepo) Count(filters domain.AuditLogFilters) (int, error) {
	return len(m.logs), nil
}

// Add missing methods to mockOrganizationRepo
func (m *mockOrganizationRepo) Create(org *domain.Organization) error {
	org.ID = m.nextID
	m.nextID++
	m.organizations[org.ID] = org
	return nil
}

func (m *mockOrganizationRepo) Update(org *domain.Organization) error {
	if _, ok := m.organizations[org.ID]; !ok {
		return sql.ErrNoRows
	}
	m.organizations[org.ID] = org
	return nil
}

func (m *mockOrganizationRepo) Delete(id int64) error {
	// Check if any users are assigned
	for _, orgMap := range m.userOrganizations {
		if _, ok := orgMap[id]; ok {
			return errors.New("cannot delete organization: users are still assigned")
		}
	}
	delete(m.organizations, id)
	return nil
}

func (m *mockOrganizationRepo) List(limit, offset int) ([]*domain.Organization, int64, error) {
	var result []*domain.Organization
	for _, org := range m.organizations {
		result = append(result, org)
	}
	return result, int64(len(result)), nil
}

func (m *mockOrganizationRepo) AddUserToOrganization(userID, orgID int64, role string) error {
	if m.userOrganizations[userID] == nil {
		m.userOrganizations[userID] = make(map[int64]string)
	}
	m.userOrganizations[userID][orgID] = role
	return nil
}

func (m *mockOrganizationRepo) RemoveUserFromOrganization(userID, orgID int64) error {
	if m.userOrganizations[userID] != nil {
		delete(m.userOrganizations[userID], orgID)
	}
	return nil
}

func (m *mockOrganizationRepo) GetUserOrganizations(userID int64) ([]*domain.Organization, error) {
	var orgs []*domain.Organization
	if orgMap, ok := m.userOrganizations[userID]; ok {
		for orgID := range orgMap {
			if org, ok := m.organizations[orgID]; ok {
				orgs = append(orgs, org)
			}
		}
	}
	return orgs, nil
}

func (m *mockOrganizationRepo) GetOrganizationUsers(orgID int64) ([]*domain.User, error) {
	if users, ok := m.orgUsers[orgID]; ok {
		return users, nil
	}
	return []*domain.User{}, nil
}

func (m *mockOrganizationRepo) IsUserInOrganization(userID, orgID int64) (bool, error) {
	if orgMap, ok := m.userOrganizations[userID]; ok {
		_, inOrg := orgMap[orgID]
		return inOrg, nil
	}
	return false, nil
}

func (m *mockOrganizationRepo) SetPermanentFree(id int64, isPermanent bool) error {
	org, ok := m.organizations[id]
	if !ok {
		return sql.ErrNoRows
	}
	m.organizations[id] = org
	return nil
}

func (m *mockOrganizationRepo) Count() (int64, error) {
	return int64(len(m.organizations)), nil
}

// Add DisableAccount and EnableAccount methods to mockUserRepo
func (m *mockUserRepo) DisableAccount(userID int64, disabledBy int64, reason string) error {
	user, ok := m.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	user.AccountDisabled = true
	user.DisableReason = &reason
	user.DisabledByUserID = &disabledBy
	now := time.Now()
	user.DisabledAt = &now
	user.UpdatedAt = time.Now()
	m.users[userID] = user
	return nil
}

func (m *mockUserRepo) EnableAccount(userID int64) error {
	user, ok := m.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	user.AccountDisabled = false
	user.DisableReason = nil
	user.DisabledByUserID = nil
	user.DisabledAt = nil
	user.UpdatedAt = time.Now()
	m.users[userID] = user
	return nil
}

// Add missing methods to mockAuditLogRepo
func (m *mockAuditLogRepo) DeleteOlderThan(before time.Time) (int, error) {
	return 0, nil
}

// Add missing IncrementFailedAttempts to mockUserRepo (different signature than existing)
func (m *mockUserRepo) IncrementFailedAttempts(userID int64) error {
	user, ok := m.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	user.FailedLoginAttempts++
	m.users[userID] = user
	return nil
}

// Add missing methods to mockOrganizationRepo
func (m *mockOrganizationRepo) GetByName(name string) (*domain.Organization, error) {
	for _, org := range m.organizations {
		if org.Name == name {
			return org, nil
		}
	}
	return nil, nil // Service expects nil, nil for not found
}

func (m *mockOrganizationRepo) GetUserOrganizationIDs(userID int64) ([]int64, error) {
	var ids []int64
	if orgMap, ok := m.userOrganizations[userID]; ok {
		for orgID := range orgMap {
			ids = append(ids, orgID)
		}
	}
	return ids, nil
}

// Add missing ResetFailedAttempts to mockUserRepo (different signature)
func (m *mockUserRepo) ResetFailedAttempts(userID int64) error {
	user, ok := m.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	user.FailedLoginAttempts = 0
	m.users[userID] = user
	return nil
}

// Add missing UnlockAccount to mockUserRepo
func (m *mockUserRepo) UnlockAccount(userID int64) error {
	user, ok := m.users[userID]
	if !ok {
		return sql.ErrNoRows
	}
	user.LockedUntil = nil
	user.FailedLoginAttempts = 0
	m.users[userID] = user
	return nil
}

// Add missing ListAdmins to mockUserRepo
func (m *mockUserRepo) ListAdmins() ([]*domain.User, error) {
	var admins []*domain.User
	for _, user := range m.users {
		if user.Role == "admin" {
			admins = append(admins, user)
		}
	}
	return admins, nil
}

// Mock DataChangeLogRepository
type mockDataChangeLogRepo struct {
	logs   []*domain.DataChangeLog
	nextID int64
}

func newMockDataChangeLogRepo() *mockDataChangeLogRepo {
	return &mockDataChangeLogRepo{
		logs:   make([]*domain.DataChangeLog, 0),
		nextID: 1,
	}
}

func (m *mockDataChangeLogRepo) Create(log *domain.DataChangeLog) error {
	log.ID = m.nextID
	m.nextID++
	log.CreatedAt = time.Now()
	m.logs = append(m.logs, log)
	return nil
}

func (m *mockDataChangeLogRepo) GetByID(id int64) (*domain.DataChangeLog, error) {
	for _, log := range m.logs {
		if log.ID == id {
			return log, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockDataChangeLogRepo) List(filters domain.DataChangeLogFilters, limit, offset int) ([]*domain.DataChangeLog, error) {
	return m.logs, nil
}

func (m *mockDataChangeLogRepo) Count(filters domain.DataChangeLogFilters) (int, error) {
	return len(m.logs), nil
}

func (m *mockDataChangeLogRepo) GetByEntityID(entityType string, entityID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	var result []*domain.DataChangeLog
	for _, log := range m.logs {
		if log.EntityType == entityType && log.EntityID == entityID {
			result = append(result, log)
		}
	}
	return result, nil
}

func (m *mockDataChangeLogRepo) GetByUserID(userID int64, limit, offset int) ([]*domain.DataChangeLog, error) {
	var result []*domain.DataChangeLog
	for _, log := range m.logs {
		if log.UserID == userID {
			result = append(result, log)
		}
	}
	return result, nil
}

func (m *mockDataChangeLogRepo) DeleteOlderThan(before time.Time) (int, error) {
	count := 0
	var remaining []*domain.DataChangeLog
	for _, log := range m.logs {
		if log.CreatedAt.Before(before) {
			count++
		} else {
			remaining = append(remaining, log)
		}
	}
	m.logs = remaining
	return count, nil
}

// Mock NotificationLikeRepository
type mockNotificationLikeRepo struct {
	likes  []*domain.NotificationLike
	nextID int64
}

func newMockNotificationLikeRepo() *mockNotificationLikeRepo {
	return &mockNotificationLikeRepo{
		likes:  make([]*domain.NotificationLike, 0),
		nextID: 1,
	}
}

func (m *mockNotificationLikeRepo) Create(like *domain.NotificationLike) error {
	like.ID = m.nextID
	m.nextID++
	like.CreatedAt = time.Now()
	m.likes = append(m.likes, like)
	return nil
}

func (m *mockNotificationLikeRepo) Delete(notificationID, userID int64) error {
	for i, like := range m.likes {
		if like.NotificationID == notificationID && like.UserID == userID {
			m.likes = append(m.likes[:i], m.likes[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

func (m *mockNotificationLikeRepo) GetByNotificationID(notificationID int64) ([]*domain.NotificationLike, error) {
	var result []*domain.NotificationLike
	for _, like := range m.likes {
		if like.NotificationID == notificationID {
			result = append(result, like)
		}
	}
	return result, nil
}

func (m *mockNotificationLikeRepo) GetLikeCount(notificationID int64) (int, error) {
	count := 0
	for _, like := range m.likes {
		if like.NotificationID == notificationID {
			count++
		}
	}
	return count, nil
}

func (m *mockNotificationLikeRepo) HasUserLiked(notificationID, userID int64) (bool, error) {
	for _, like := range m.likes {
		if like.NotificationID == notificationID && like.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

// Mock NotificationRepository (for notification like service tests)
type mockNotificationRepo struct {
	notifications map[int64]*domain.Notification
	nextID        int64
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{
		notifications: make(map[int64]*domain.Notification),
		nextID:        1,
	}
}

func (m *mockNotificationRepo) Create(notification *domain.Notification) error {
	notification.ID = m.nextID
	m.nextID++
	notification.CreatedAt = time.Now()
	m.notifications[notification.ID] = notification
	return nil
}

func (m *mockNotificationRepo) GetByID(id int64) (*domain.Notification, error) {
	notification, ok := m.notifications[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return notification, nil
}

func (m *mockNotificationRepo) GetByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *mockNotificationRepo) GetUnreadByUserID(userID int64, limit, offset int) ([]*domain.Notification, error) {
	var result []*domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *mockNotificationRepo) CountUnreadByUserID(userID int64) (int64, error) {
	count := int64(0)
	for _, n := range m.notifications {
		if n.UserID == userID && n.ReadAt == nil {
			count++
		}
	}
	return count, nil
}

func (m *mockNotificationRepo) MarkAsRead(id int64) error {
	if n, ok := m.notifications[id]; ok {
		now := time.Now()
		n.ReadAt = &now
		return nil
	}
	return sql.ErrNoRows
}

func (m *mockNotificationRepo) MarkAsUnread(id int64) error {
	if n, ok := m.notifications[id]; ok {
		n.ReadAt = nil
		return nil
	}
	return sql.ErrNoRows
}

func (m *mockNotificationRepo) MarkAllAsReadForUser(userID int64) error {
	now := time.Now()
	for _, n := range m.notifications {
		if n.UserID == userID {
			n.ReadAt = &now
		}
	}
	return nil
}

func (m *mockNotificationRepo) Delete(id int64) error {
	delete(m.notifications, id)
	return nil
}

func (m *mockNotificationRepo) DeleteOlderThan(before time.Time) (int64, error) {
	return 0, nil
}

// Mock UserSettingsRepository
type mockUserSettingsRepo struct {
	settings    map[int64]*domain.UserSettings
	createError error
	updateError error
}

func newMockUserSettingsRepo() *mockUserSettingsRepo {
	return &mockUserSettingsRepo{
		settings: make(map[int64]*domain.UserSettings),
	}
}

func (m *mockUserSettingsRepo) GetByUserID(userID int64) (*domain.UserSettings, error) {
	settings, ok := m.settings[userID]
	if !ok {
		return nil, nil // Not found returns nil, nil
	}
	return settings, nil
}

func (m *mockUserSettingsRepo) Create(settings *domain.UserSettings) error {
	if m.createError != nil {
		return m.createError
	}
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()
	m.settings[settings.UserID] = settings
	return nil
}

func (m *mockUserSettingsRepo) Update(settings *domain.UserSettings) error {
	if m.updateError != nil {
		return m.updateError
	}
	settings.UpdatedAt = time.Now()
	m.settings[settings.UserID] = settings
	return nil
}

func (m *mockUserSettingsRepo) Delete(userID int64) error {
	delete(m.settings, userID)
	return nil
}

// Mock WorkoutWODRepository
type mockWorkoutWODRepo struct {
	workoutWODs                   map[int64]*domain.WorkoutWOD
	nextID                        int64
	createError                   error
	deleteError                   error
	getByIDError                  error
	updateError                   error
	togglePRError                 error
	listByWorkoutWithDetailsError error
}

func newMockWorkoutWODRepo() *mockWorkoutWODRepo {
	return &mockWorkoutWODRepo{
		workoutWODs: make(map[int64]*domain.WorkoutWOD),
		nextID:      1,
	}
}

func (m *mockWorkoutWODRepo) Create(workoutWOD *domain.WorkoutWOD) error {
	if m.createError != nil {
		return m.createError
	}
	workoutWOD.ID = m.nextID
	m.nextID++
	workoutWOD.CreatedAt = time.Now()
	workoutWOD.UpdatedAt = time.Now()
	m.workoutWODs[workoutWOD.ID] = workoutWOD
	return nil
}

func (m *mockWorkoutWODRepo) GetByID(id int64) (*domain.WorkoutWOD, error) {
	if m.getByIDError != nil {
		return nil, m.getByIDError
	}
	ww, ok := m.workoutWODs[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return ww, nil
}

func (m *mockWorkoutWODRepo) ListByWorkout(workoutID int64) ([]*domain.WorkoutWOD, error) {
	var result []*domain.WorkoutWOD
	for _, ww := range m.workoutWODs {
		if ww.WorkoutID == workoutID {
			result = append(result, ww)
		}
	}
	return result, nil
}

func (m *mockWorkoutWODRepo) ListByWorkoutWithDetails(workoutID int64) ([]*domain.WorkoutWODWithDetails, error) {
	if m.listByWorkoutWithDetailsError != nil {
		return nil, m.listByWorkoutWithDetailsError
	}
	var result []*domain.WorkoutWODWithDetails
	for _, ww := range m.workoutWODs {
		if ww.WorkoutID == workoutID {
			result = append(result, &domain.WorkoutWODWithDetails{
				WorkoutWOD:     *ww,
				WODName:        "Test WOD",
				WODType:        "AMRAP",
				WODRegime:      "CrossFit",
				WODScoreType:   "rounds_reps",
				WODDescription: "Test description",
			})
		}
	}
	return result, nil
}

func (m *mockWorkoutWODRepo) Update(workoutWOD *domain.WorkoutWOD) error {
	if m.updateError != nil {
		return m.updateError
	}
	if _, ok := m.workoutWODs[workoutWOD.ID]; !ok {
		return sql.ErrNoRows
	}
	workoutWOD.UpdatedAt = time.Now()
	m.workoutWODs[workoutWOD.ID] = workoutWOD
	return nil
}

func (m *mockWorkoutWODRepo) Delete(id int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, ok := m.workoutWODs[id]; !ok {
		return sql.ErrNoRows
	}
	delete(m.workoutWODs, id)
	return nil
}

func (m *mockWorkoutWODRepo) DeleteByWorkout(workoutID int64) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	for id, ww := range m.workoutWODs {
		if ww.WorkoutID == workoutID {
			delete(m.workoutWODs, id)
		}
	}
	return nil
}

func (m *mockWorkoutWODRepo) TogglePR(id int64) error {
	if m.togglePRError != nil {
		return m.togglePRError
	}
	ww, ok := m.workoutWODs[id]
	if !ok {
		return sql.ErrNoRows
	}
	ww.IsPR = !ww.IsPR
	return nil
}
