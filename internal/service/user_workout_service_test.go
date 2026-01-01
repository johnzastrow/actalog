package service

import (
	"errors"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestUserWorkoutService_LogWorkout(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		workoutID     *int64
		workoutName   *string
		workoutDate   time.Time
		notes         *string
		totalTime     *int
		workoutType   *string
		setupMock     func(*mockWorkoutRepo)
		expectedError error
	}{
		{
			name:        "successful workout log",
			userID:      1,
			workoutID:   int64Ptr(1),
			workoutName: nil,
			workoutDate: time.Now(),
			notes:       stringPtr("Great workout!"),
			totalTime:   intPtr(1800),
			workoutType: stringPtr("metcon"),
			setupMock: func(m *mockWorkoutRepo) {
				userID := int64(1)
				m.workouts[1] = &domain.Workout{
					ID:        1,
					Name:      "Fran",
					CreatedBy: &userID,
				}
			},
			expectedError: nil,
		},
		{
			name:        "log standard workout (created_by is null)",
			userID:      1,
			workoutID:   int64Ptr(2),
			workoutName: nil,
			workoutDate: time.Now(),
			notes:       nil,
			totalTime:   nil,
			workoutType: nil,
			setupMock: func(m *mockWorkoutRepo) {
				m.workouts[2] = &domain.Workout{
					ID:        2,
					Name:      "Cindy",
					CreatedBy: nil, // Standard workout
				}
			},
			expectedError: nil,
		},
		{
			name:        "workout template not found",
			userID:      1,
			workoutID:   int64Ptr(999),
			workoutName: nil,
			workoutDate: time.Now(),
			notes:       nil,
			totalTime:   nil,
			workoutType: nil,
			setupMock: func(m *mockWorkoutRepo) {
				// Workout 999 doesn't exist
			},
			expectedError: ErrWorkoutNotFound,
		},
		{
			name:        "unauthorized access to another user's workout",
			userID:      1,
			workoutID:   int64Ptr(3),
			workoutName: nil,
			workoutDate: time.Now(),
			notes:       nil,
			totalTime:   nil,
			workoutType: nil,
			setupMock: func(m *mockWorkoutRepo) {
				otherUserID := int64(2)
				m.workouts[3] = &domain.Workout{
					ID:        3,
					Name:      "Custom Workout",
					CreatedBy: &otherUserID,
				}
			},
			expectedError: ErrUnauthorizedWorkoutAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userWorkoutRepo := newMockUserWorkoutRepo()
			workoutRepo := newMockWorkoutRepo()
			workoutMovementRepo := &mockWorkoutMovementRepo{}

			if tt.setupMock != nil {
				tt.setupMock(workoutRepo)
			}

			service := NewUserWorkoutService(
				userWorkoutRepo,
				workoutRepo,
				workoutMovementRepo,
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				newMockWODRepo(),
				nil,                 // auditLogRepo
				&mockMovementRepo{}, // movementRepo
				nil,                 // notificationService
				&mockUserRepo{},     // userRepo
				nil,                 // orgRepo
			)

			userWorkout, err := service.LogWorkout(
				tt.userID,
				"test@example.com",
				tt.workoutID,
				tt.workoutName,
				tt.workoutDate,
				tt.notes,
				tt.totalTime,
				tt.workoutType,
			)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if userWorkout == nil {
				t.Error("expected user workout, got nil")
				return
			}

			if userWorkout.UserID != tt.userID {
				t.Errorf("expected user ID %d, got %d", tt.userID, userWorkout.UserID)
			}

			if tt.workoutID != nil && (userWorkout.WorkoutID == nil || *userWorkout.WorkoutID != *tt.workoutID) {
				t.Errorf("expected workout ID %d, got %v", *tt.workoutID, userWorkout.WorkoutID)
			}
		})
	}
}

func TestUserWorkoutService_GetLoggedWorkout(t *testing.T) {
	tests := []struct {
		name          string
		userWorkoutID int64
		userID        int64
		setupMock     func(*mockUserWorkoutRepo)
		expectedError error
	}{
		{
			name:          "successful retrieval",
			userWorkoutID: 1,
			userID:        1,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Now(),
				}
			},
			expectedError: nil,
		},
		{
			name:          "workout not found",
			userWorkoutID: 999,
			userID:        1,
			setupMock: func(m *mockUserWorkoutRepo) {
				// Workout 999 doesn't exist
			},
			expectedError: ErrUserWorkoutNotFound,
		},
		{
			name:          "unauthorized access",
			userWorkoutID: 2,
			userID:        1,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[2] = &domain.UserWorkout{
					ID:          2,
					UserID:      2, // Different user
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Now(),
				}
			},
			expectedError: ErrUnauthorizedWorkoutAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userWorkoutRepo := newMockUserWorkoutRepo()
			workoutRepo := newMockWorkoutRepo()
			workoutMovementRepo := &mockWorkoutMovementRepo{}

			if tt.setupMock != nil {
				tt.setupMock(userWorkoutRepo)
			}

			service := NewUserWorkoutService(
				userWorkoutRepo,
				workoutRepo,
				workoutMovementRepo,
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				newMockWODRepo(),
				nil,                 // auditLogRepo
				&mockMovementRepo{}, // movementRepo
				nil,                 // notificationService
				&mockUserRepo{},     // userRepo
				nil,                 // orgRepo
			)

			userWorkout, err := service.GetLoggedWorkout(tt.userWorkoutID, tt.userID)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if userWorkout == nil {
				t.Error("expected user workout, got nil")
				return
			}

			if userWorkout.UserID != tt.userID {
				t.Errorf("expected user ID %d, got %d", tt.userID, userWorkout.UserID)
			}
		})
	}
}

func TestUserWorkoutService_UpdateLoggedWorkout(t *testing.T) {
	tests := []struct {
		name          string
		userWorkoutID int64
		userID        int64
		workoutName   *string
		notes         *string
		totalTime     *int
		workoutType   *string
		setupMock     func(*mockUserWorkoutRepo)
		expectedError error
	}{
		{
			name:          "successful update",
			userWorkoutID: 1,
			userID:        1,
			workoutName:   nil,
			notes:         stringPtr("Updated notes"),
			totalTime:     intPtr(2000),
			workoutType:   stringPtr("strength"),
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Now(),
				}
			},
			expectedError: nil,
		},
		{
			name:          "workout not found",
			userWorkoutID: 999,
			userID:        1,
			workoutName:   nil,
			notes:         stringPtr("Updated notes"),
			totalTime:     nil,
			workoutType:   nil,
			setupMock: func(m *mockUserWorkoutRepo) {
				// Workout 999 doesn't exist
			},
			expectedError: ErrUserWorkoutNotFound,
		},
		{
			name:          "unauthorized update",
			userWorkoutID: 2,
			userID:        1,
			workoutName:   nil,
			notes:         stringPtr("Trying to update"),
			totalTime:     nil,
			workoutType:   nil,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[2] = &domain.UserWorkout{
					ID:          2,
					UserID:      2, // Different user
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Now(),
				}
			},
			expectedError: ErrUnauthorizedWorkoutAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userWorkoutRepo := newMockUserWorkoutRepo()
			workoutRepo := newMockWorkoutRepo()
			workoutMovementRepo := &mockWorkoutMovementRepo{}

			if tt.setupMock != nil {
				tt.setupMock(userWorkoutRepo)
			}

			service := NewUserWorkoutService(
				userWorkoutRepo,
				workoutRepo,
				workoutMovementRepo,
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				newMockWODRepo(),
				nil,                 // auditLogRepo
				&mockMovementRepo{}, // movementRepo
				nil,                 // notificationService
				&mockUserRepo{},     // userRepo
				nil,                 // orgRepo
			)

			err := service.UpdateLoggedWorkout(
				tt.userWorkoutID,
				tt.userID,
				"test@example.com",
				tt.workoutName,
				tt.notes,
				tt.totalTime,
				tt.workoutType,
			)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify the update was applied
			updated := userWorkoutRepo.userWorkouts[tt.userWorkoutID]
			if tt.notes != nil && (updated.Notes == nil || *updated.Notes != *tt.notes) {
				t.Errorf("notes not updated correctly")
			}
			if tt.totalTime != nil && (updated.TotalTime == nil || *updated.TotalTime != *tt.totalTime) {
				t.Errorf("total time not updated correctly")
			}
		})
	}
}

func TestUserWorkoutService_DeleteLoggedWorkout(t *testing.T) {
	tests := []struct {
		name          string
		userWorkoutID int64
		userID        int64
		setupMock     func(*mockUserWorkoutRepo)
		expectedError error
	}{
		{
			name:          "successful deletion",
			userWorkoutID: 1,
			userID:        1,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Now(),
				}
			},
			expectedError: nil,
		},
		{
			name:          "workout not found",
			userWorkoutID: 999,
			userID:        1,
			setupMock: func(m *mockUserWorkoutRepo) {
				// Workout 999 doesn't exist
			},
			expectedError: ErrUserWorkoutNotFound,
		},
		{
			name:          "unauthorized deletion",
			userWorkoutID: 2,
			userID:        1,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[2] = &domain.UserWorkout{
					ID:          2,
					UserID:      2, // Different user
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Now(),
				}
			},
			expectedError: ErrUnauthorizedWorkoutAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userWorkoutRepo := newMockUserWorkoutRepo()
			workoutRepo := newMockWorkoutRepo()
			workoutMovementRepo := &mockWorkoutMovementRepo{}

			if tt.setupMock != nil {
				tt.setupMock(userWorkoutRepo)
			}

			service := NewUserWorkoutService(
				userWorkoutRepo,
				workoutRepo,
				workoutMovementRepo,
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				newMockWODRepo(),
				nil,                 // auditLogRepo
				&mockMovementRepo{}, // movementRepo
				nil,                 // notificationService
				&mockUserRepo{},     // userRepo
				nil,                 // orgRepo
			)

			err := service.DeleteLoggedWorkout(tt.userWorkoutID, tt.userID, "test@example.com")

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify deletion
			if _, exists := userWorkoutRepo.userWorkouts[tt.userWorkoutID]; exists {
				t.Error("workout should have been deleted")
			}
		})
	}
}

func TestUserWorkoutService_GetWorkoutStatsForMonth(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		year          int
		month         int
		setupMock     func(*mockUserWorkoutRepo)
		expectedCount int
	}{
		{
			name:   "workouts in month",
			userID: 1,
			year:   2025,
			month:  1,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				}
				m.userWorkouts[2] = &domain.UserWorkout{
					ID:          2,
					UserID:      1,
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
				}
				m.userWorkouts[3] = &domain.UserWorkout{
					ID:          3,
					UserID:      1,
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC), // Different month
				}
			},
			expectedCount: 2,
		},
		{
			name:   "no workouts in month",
			userID: 1,
			year:   2025,
			month:  3,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutID:   int64Ptr(1),
					WorkoutDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				}
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userWorkoutRepo := newMockUserWorkoutRepo()
			workoutRepo := newMockWorkoutRepo()
			workoutMovementRepo := &mockWorkoutMovementRepo{}

			if tt.setupMock != nil {
				tt.setupMock(userWorkoutRepo)
			}

			service := NewUserWorkoutService(
				userWorkoutRepo,
				workoutRepo,
				workoutMovementRepo,
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				newMockWODRepo(),
				nil,                 // auditLogRepo
				&mockMovementRepo{}, // movementRepo
				nil,                 // notificationService
				&mockUserRepo{},     // userRepo
				nil,                 // orgRepo
			)

			count, err := service.GetWorkoutStatsForMonth(tt.userID, tt.year, tt.month)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if count != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

func TestUserWorkoutService_ListLoggedWorkouts(t *testing.T) {
	userWorkoutRepo := newMockUserWorkoutRepo()
	workoutRepo := newMockWorkoutRepo()

	// Setup test data
	userWorkoutRepo.userWorkouts[1] = &domain.UserWorkout{
		ID:          1,
		UserID:      1,
		WorkoutID:   int64Ptr(1),
		WorkoutDate: time.Now(),
	}
	userWorkoutRepo.userWorkouts[2] = &domain.UserWorkout{
		ID:          2,
		UserID:      1,
		WorkoutID:   int64Ptr(2),
		WorkoutDate: time.Now().AddDate(0, 0, -1),
	}

	service := NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	workouts, err := service.ListLoggedWorkouts(1, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(workouts) != 2 {
		t.Errorf("expected 2 workouts, got %d", len(workouts))
	}
}

func TestUserWorkoutService_ListLoggedWorkoutsByDateRange(t *testing.T) {
	userWorkoutRepo := newMockUserWorkoutRepo()
	workoutRepo := newMockWorkoutRepo()

	// Setup test data with specific dates
	jan15 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	jan20 := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	feb5 := time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC)

	userWorkoutRepo.userWorkouts[1] = &domain.UserWorkout{
		ID:          1,
		UserID:      1,
		WorkoutID:   int64Ptr(1),
		WorkoutDate: jan15,
	}
	userWorkoutRepo.userWorkouts[2] = &domain.UserWorkout{
		ID:          2,
		UserID:      1,
		WorkoutID:   int64Ptr(2),
		WorkoutDate: jan20,
	}
	userWorkoutRepo.userWorkouts[3] = &domain.UserWorkout{
		ID:          3,
		UserID:      1,
		WorkoutID:   int64Ptr(3),
		WorkoutDate: feb5,
	}

	service := NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	// Query for January only
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 1, 31, 23, 59, 59, 0, time.UTC)

	workouts, err := service.ListLoggedWorkoutsByDateRange(1, startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(workouts) != 2 {
		t.Errorf("expected 2 workouts in January, got %d", len(workouts))
	}
}

func TestUserWorkoutService_UpdateWorkoutMovements(t *testing.T) {
	tests := []struct {
		name          string
		userWorkoutID int64
		userID        int64
		movements     []domain.UserWorkoutMovement
		setupMock     func(*mockUserWorkoutRepo)
		expectedError error
	}{
		{
			name:          "successful update",
			userWorkoutID: 1,
			userID:        1,
			movements: []domain.UserWorkoutMovement{
				{MovementID: 1, Weight: float64Ptr(100), Reps: intPtr(5)},
				{MovementID: 2, Weight: float64Ptr(150), Reps: intPtr(3)},
			},
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutDate: time.Now(),
				}
			},
			expectedError: nil,
		},
		{
			name:          "workout not found",
			userWorkoutID: 999,
			userID:        1,
			movements:     []domain.UserWorkoutMovement{},
			setupMock:     func(m *mockUserWorkoutRepo) {},
			expectedError: ErrUserWorkoutNotFound,
		},
		{
			name:          "unauthorized",
			userWorkoutID: 1,
			userID:        1,
			movements:     []domain.UserWorkoutMovement{},
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      2, // Different user
					WorkoutDate: time.Now(),
				}
			},
			expectedError: ErrUnauthorizedWorkoutAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userWorkoutRepo := newMockUserWorkoutRepo()
			if tt.setupMock != nil {
				tt.setupMock(userWorkoutRepo)
			}

			service := NewUserWorkoutService(
				userWorkoutRepo,
				newMockWorkoutRepo(),
				&mockWorkoutMovementRepo{},
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				newMockWODRepo(),
				nil,
				&mockMovementRepo{},
				nil,
				&mockUserRepo{},
				nil,
			)

			err := service.UpdateWorkoutMovements(tt.userWorkoutID, tt.userID, tt.movements)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUserWorkoutService_UpdateWorkoutWODs(t *testing.T) {
	tests := []struct {
		name          string
		userWorkoutID int64
		userID        int64
		wods          []domain.UserWorkoutWOD
		setupMock     func(*mockUserWorkoutRepo, *mockWODRepo)
		expectedError error
	}{
		{
			name:          "successful update with time-based WOD",
			userWorkoutID: 1,
			userID:        1,
			wods: []domain.UserWorkoutWOD{
				{WODID: 1, TimeSeconds: intPtr(300)},
			},
			setupMock: func(uwr *mockUserWorkoutRepo, wr *mockWODRepo) {
				uwr.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutDate: time.Now(),
				}
				wr.wods[1] = &domain.WOD{
					ID:        1,
					Name:      "Fran",
					ScoreType: "Time (HH:MM:SS)",
				}
			},
			expectedError: nil,
		},
		{
			name:          "workout not found",
			userWorkoutID: 999,
			userID:        1,
			wods:          []domain.UserWorkoutWOD{},
			setupMock:     func(uwr *mockUserWorkoutRepo, wr *mockWODRepo) {},
			expectedError: ErrUserWorkoutNotFound,
		},
		{
			name:          "unauthorized",
			userWorkoutID: 1,
			userID:        1,
			wods:          []domain.UserWorkoutWOD{},
			setupMock: func(uwr *mockUserWorkoutRepo, wr *mockWODRepo) {
				uwr.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      2, // Different user
					WorkoutDate: time.Now(),
				}
			},
			expectedError: ErrUnauthorizedWorkoutAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userWorkoutRepo := newMockUserWorkoutRepo()
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(userWorkoutRepo, wodRepo)
			}

			service := NewUserWorkoutService(
				userWorkoutRepo,
				newMockWorkoutRepo(),
				&mockWorkoutMovementRepo{},
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				wodRepo,
				nil,
				&mockMovementRepo{},
				nil,
				&mockUserRepo{},
				nil,
			)

			err := service.UpdateWorkoutWODs(tt.userWorkoutID, tt.userID, tt.wods)

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUserWorkoutService_ValidateWODScoreTypes(t *testing.T) {
	tests := []struct {
		name        string
		wods        []*domain.UserWorkoutWOD
		setupWODs   func(*mockWODRepo)
		expectError bool
	}{
		{
			name: "valid time-based WOD",
			wods: []*domain.UserWorkoutWOD{
				{WODID: 1, TimeSeconds: intPtr(300)},
			},
			setupWODs: func(wr *mockWODRepo) {
				wr.wods[1] = &domain.WOD{
					ID:        1,
					Name:      "Fran",
					ScoreType: "Time (HH:MM:SS)",
				}
			},
			expectError: false,
		},
		{
			name: "valid rounds+reps WOD",
			wods: []*domain.UserWorkoutWOD{
				{WODID: 2, Rounds: intPtr(10), Reps: intPtr(5)},
			},
			setupWODs: func(wr *mockWODRepo) {
				wr.wods[2] = &domain.WOD{
					ID:        2,
					Name:      "Cindy",
					ScoreType: "Rounds+Reps",
				}
			},
			expectError: false,
		},
		{
			name: "valid max weight WOD",
			wods: []*domain.UserWorkoutWOD{
				{WODID: 3, Weight: float64Ptr(225)},
			},
			setupWODs: func(wr *mockWODRepo) {
				wr.wods[3] = &domain.WOD{
					ID:        3,
					Name:      "1RM Deadlift",
					ScoreType: "Max Weight",
				}
			},
			expectError: false,
		},
		{
			name: "invalid - time WOD missing time_seconds",
			wods: []*domain.UserWorkoutWOD{
				{WODID: 1, Rounds: intPtr(5)}, // Wrong field
			},
			setupWODs: func(wr *mockWODRepo) {
				wr.wods[1] = &domain.WOD{
					ID:        1,
					Name:      "Fran",
					ScoreType: "Time (HH:MM:SS)",
				}
			},
			expectError: true,
		},
		{
			name: "invalid - rounds WOD missing rounds",
			wods: []*domain.UserWorkoutWOD{
				{WODID: 2, TimeSeconds: intPtr(600)}, // Wrong field
			},
			setupWODs: func(wr *mockWODRepo) {
				wr.wods[2] = &domain.WOD{
					ID:        2,
					Name:      "Cindy",
					ScoreType: "Rounds+Reps",
				}
			},
			expectError: true,
		},
		{
			name: "invalid - max weight WOD missing weight",
			wods: []*domain.UserWorkoutWOD{
				{WODID: 3, Rounds: intPtr(1)}, // Wrong field
			},
			setupWODs: func(wr *mockWODRepo) {
				wr.wods[3] = &domain.WOD{
					ID:        3,
					Name:      "1RM Deadlift",
					ScoreType: "Max Weight",
				}
			},
			expectError: true,
		},
		{
			name: "invalid - WOD not found",
			wods: []*domain.UserWorkoutWOD{
				{WODID: 999, TimeSeconds: intPtr(300)},
			},
			setupWODs:   func(wr *mockWODRepo) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupWODs != nil {
				tt.setupWODs(wodRepo)
			}

			service := NewUserWorkoutService(
				newMockUserWorkoutRepo(),
				newMockWorkoutRepo(),
				&mockWorkoutMovementRepo{},
				&mockUserWorkoutMovementRepo{},
				&mockUserWorkoutWODRepo{},
				wodRepo,
				nil,
				&mockMovementRepo{},
				nil,
				&mockUserRepo{},
				nil,
			)

			err := service.ValidateWODScoreTypes(tt.wods)

			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUserWorkoutService_DetectAndFlagMovementPRs(t *testing.T) {
	service := NewUserWorkoutService(
		newMockUserWorkoutRepo(),
		newMockWorkoutRepo(),
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	// Test with no previous performances (first time doing movement = PR)
	movements := []*domain.UserWorkoutMovement{
		{UserWorkoutID: 1, MovementID: 1, Weight: float64Ptr(100), Reps: intPtr(5)},
	}

	err := service.DetectAndFlagMovementPRs(1, movements)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First time doing a movement should be a PR
	if !movements[0].IsPR {
		t.Error("first time doing movement should be flagged as PR")
	}

	// Test movement without weight/reps (should not be flagged)
	movements2 := []*domain.UserWorkoutMovement{
		{UserWorkoutID: 1, MovementID: 2, Weight: nil, Reps: nil},
	}

	err = service.DetectAndFlagMovementPRs(1, movements2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if movements2[0].IsPR {
		t.Error("movement without weight/reps should not be flagged as PR")
	}
}

func TestUserWorkoutService_DetectAndFlagWODPRs(t *testing.T) {
	service := NewUserWorkoutService(
		newMockUserWorkoutRepo(),
		newMockWorkoutRepo(),
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	// Test with time-based WOD (first time = PR)
	wods := []*domain.UserWorkoutWOD{
		{WODID: 1, TimeSeconds: intPtr(300)},
	}

	err := service.DetectAndFlagWODPRs(1, wods)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !wods[0].IsPR {
		t.Error("first time doing WOD should be flagged as PR")
	}

	// Test with rounds+reps WOD (first time = PR)
	wods2 := []*domain.UserWorkoutWOD{
		{WODID: 2, Rounds: intPtr(10), Reps: intPtr(5)},
	}

	err = service.DetectAndFlagWODPRs(1, wods2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !wods2[0].IsPR {
		t.Error("first time doing rounds+reps WOD should be flagged as PR")
	}
}

func TestUserWorkoutService_GetPRMovements(t *testing.T) {
	service := NewUserWorkoutService(
		newMockUserWorkoutRepo(),
		newMockWorkoutRepo(),
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	movements, err := service.GetPRMovements(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With empty mock, should return empty list
	if movements == nil {
		t.Error("expected non-nil slice")
	}
}

func TestUserWorkoutService_GetPRWODs(t *testing.T) {
	service := NewUserWorkoutService(
		newMockUserWorkoutRepo(),
		newMockWorkoutRepo(),
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	wods, err := service.GetPRWODs(1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// With empty mock, should return empty list
	if wods == nil {
		t.Error("expected non-nil slice")
	}
}

func TestUserWorkoutService_LogWorkout_AdHoc(t *testing.T) {
	// Test logging an ad-hoc workout (no template)
	userWorkoutRepo := newMockUserWorkoutRepo()
	workoutRepo := newMockWorkoutRepo()

	service := NewUserWorkoutService(
		userWorkoutRepo,
		workoutRepo,
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	workoutName := "My Custom Workout"
	userWorkout, err := service.LogWorkout(
		1,
		"test@example.com",
		nil, // No template ID
		&workoutName,
		time.Now(),
		stringPtr("Ad-hoc workout notes"),
		intPtr(1800),
		stringPtr("strength"),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if userWorkout.WorkoutID != nil {
		t.Error("ad-hoc workout should have nil WorkoutID")
	}

	if userWorkout.WorkoutName == nil || *userWorkout.WorkoutName != workoutName {
		t.Errorf("expected workout name '%s', got '%v'", workoutName, userWorkout.WorkoutName)
	}
}

func TestUserWorkoutService_UpdateLoggedWorkout_AdHocName(t *testing.T) {
	// Test that ad-hoc workouts can have their name updated
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutRepo.userWorkouts[1] = &domain.UserWorkout{
		ID:          1,
		UserID:      1,
		WorkoutID:   nil, // Ad-hoc workout
		WorkoutName: stringPtr("Original Name"),
		WorkoutDate: time.Now(),
	}

	service := NewUserWorkoutService(
		userWorkoutRepo,
		newMockWorkoutRepo(),
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	newName := "Updated Name"
	err := service.UpdateLoggedWorkout(1, 1, "test@example.com", &newName, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := userWorkoutRepo.userWorkouts[1]
	if updated.WorkoutName == nil || *updated.WorkoutName != newName {
		t.Errorf("expected updated name '%s', got '%v'", newName, updated.WorkoutName)
	}
}

func TestUserWorkoutService_UpdateLoggedWorkout_TemplateNameIgnored(t *testing.T) {
	// Test that template-based workouts cannot have their name updated
	userWorkoutRepo := newMockUserWorkoutRepo()
	userWorkoutRepo.userWorkouts[1] = &domain.UserWorkout{
		ID:          1,
		UserID:      1,
		WorkoutID:   int64Ptr(1), // Template-based workout
		WorkoutName: nil,
		WorkoutDate: time.Now(),
	}

	service := NewUserWorkoutService(
		userWorkoutRepo,
		newMockWorkoutRepo(),
		&mockWorkoutMovementRepo{},
		&mockUserWorkoutMovementRepo{},
		&mockUserWorkoutWODRepo{},
		newMockWODRepo(),
		nil,
		&mockMovementRepo{},
		nil,
		&mockUserRepo{},
		nil,
	)

	newName := "Trying to change name"
	err := service.UpdateLoggedWorkout(1, 1, "test@example.com", &newName, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Name should NOT be updated for template-based workout
	updated := userWorkoutRepo.userWorkouts[1]
	if updated.WorkoutName != nil {
		t.Error("template-based workout name should not be updated")
	}
}

// Helper function
func float64Ptr(v float64) *float64 {
	return &v
}
