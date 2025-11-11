package service

import (
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestUserWorkoutService_LogWorkout(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		workoutID     int64
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
			workoutID:   1,
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
			workoutID:   2,
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
			workoutID:   999,
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
			workoutID:   3,
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

			service := NewUserWorkoutService(userWorkoutRepo, workoutRepo, workoutMovementRepo)

			userWorkout, err := service.LogWorkout(
				tt.userID,
				tt.workoutID,
				tt.workoutDate,
				tt.notes,
				tt.totalTime,
				tt.workoutType,
			)

			if tt.expectedError != nil {
				if err != tt.expectedError {
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

			if userWorkout.WorkoutID != tt.workoutID {
				t.Errorf("expected workout ID %d, got %d", tt.workoutID, userWorkout.WorkoutID)
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
					WorkoutID:   1,
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
					WorkoutID:   1,
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

			service := NewUserWorkoutService(userWorkoutRepo, workoutRepo, workoutMovementRepo)

			userWorkout, err := service.GetLoggedWorkout(tt.userWorkoutID, tt.userID)

			if tt.expectedError != nil {
				if err != tt.expectedError {
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
			notes:         stringPtr("Updated notes"),
			totalTime:     intPtr(2000),
			workoutType:   stringPtr("strength"),
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[1] = &domain.UserWorkout{
					ID:          1,
					UserID:      1,
					WorkoutID:   1,
					WorkoutDate: time.Now(),
				}
			},
			expectedError: nil,
		},
		{
			name:          "workout not found",
			userWorkoutID: 999,
			userID:        1,
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
			notes:         stringPtr("Trying to update"),
			totalTime:     nil,
			workoutType:   nil,
			setupMock: func(m *mockUserWorkoutRepo) {
				m.userWorkouts[2] = &domain.UserWorkout{
					ID:          2,
					UserID:      2, // Different user
					WorkoutID:   1,
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

			service := NewUserWorkoutService(userWorkoutRepo, workoutRepo, workoutMovementRepo)

			err := service.UpdateLoggedWorkout(
				tt.userWorkoutID,
				tt.userID,
				tt.notes,
				tt.totalTime,
				tt.workoutType,
			)

			if tt.expectedError != nil {
				if err != tt.expectedError {
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
					WorkoutID:   1,
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
					WorkoutID:   1,
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

			service := NewUserWorkoutService(userWorkoutRepo, workoutRepo, workoutMovementRepo)

			err := service.DeleteLoggedWorkout(tt.userWorkoutID, tt.userID)

			if tt.expectedError != nil {
				if err != tt.expectedError {
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
					WorkoutID:   1,
					WorkoutDate: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				}
				m.userWorkouts[2] = &domain.UserWorkout{
					ID:          2,
					UserID:      1,
					WorkoutID:   1,
					WorkoutDate: time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC),
				}
				m.userWorkouts[3] = &domain.UserWorkout{
					ID:          3,
					UserID:      1,
					WorkoutID:   1,
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
					WorkoutID:   1,
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

			service := NewUserWorkoutService(userWorkoutRepo, workoutRepo, workoutMovementRepo)

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
