package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestWODService_CreateWOD(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		wod           *domain.WOD
		setupMock     func(*mockWODRepo)
		expectedError error
	}{
		{
			name:   "successful custom WOD creation",
			userID: 1,
			wod: &domain.WOD{
				Name:        "My Custom WOD",
				Source:      "Self-recorded",
				Type:        "Self-created",
				Regime:      "AMRAP",
				ScoreType:   "Rounds+Reps",
				Description: "10 min AMRAP: 10 burpees, 10 pull-ups",
			},
			setupMock:     func(m *mockWODRepo) {},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			err := service.Create(tt.wod, tt.userID, "test@example.com")

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

			// Verify WOD was created with correct attributes
			if tt.wod.ID == 0 {
				t.Error("WOD ID should be set after creation")
			}
			if tt.wod.CreatedBy == nil || *tt.wod.CreatedBy != tt.userID {
				t.Errorf("CreatedBy should be set to user ID %d", tt.userID)
			}
			if tt.wod.IsStandard {
				t.Error("Custom WOD should not be marked as standard")
			}
			if tt.wod.CreatedAt.IsZero() {
				t.Error("CreatedAt should be set")
			}
			if tt.wod.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should be set")
			}
		})
	}
}

func TestWODService_GetWOD(t *testing.T) {
	tests := []struct {
		name          string
		wodID         int64
		setupMock     func(*mockWODRepo)
		expectedError bool
	}{
		{
			name:  "successful retrieval",
			wodID: 1,
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:          1,
					Name:        "Fran",
					IsStandard:  true,
					CreatedBy:   &userID,
					Description: "21-15-9: Thrusters, Pull-ups",
				}
			},
			expectedError: false,
		},
		{
			name:  "WOD not found",
			wodID: 999,
			setupMock: func(m *mockWODRepo) {
				// WOD 999 doesn't exist
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			wod, err := service.GetByID(tt.wodID)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if wod == nil {
				t.Error("expected WOD, got nil")
				return
			}

			if wod.ID != tt.wodID {
				t.Errorf("expected WOD ID %d, got %d", tt.wodID, wod.ID)
			}
		})
	}
}

func TestWODService_GetWODByName(t *testing.T) {
	tests := []struct {
		name          string
		wodName       string
		setupMock     func(*mockWODRepo)
		expectedError bool
	}{
		{
			name:    "successful retrieval by name",
			wodName: "Fran",
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{
					ID:          1,
					Name:        "Fran",
					IsStandard:  true,
					Description: "21-15-9: Thrusters, Pull-ups",
				}
			},
			expectedError: false,
		},
		{
			name:    "WOD not found",
			wodName: "NonExistent",
			setupMock: func(m *mockWODRepo) {
				// WOD doesn't exist
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			wod, err := service.GetByName(tt.wodName)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if wod == nil {
				t.Error("expected WOD, got nil")
				return
			}

			if wod.Name != tt.wodName {
				t.Errorf("expected WOD name %s, got %s", tt.wodName, wod.Name)
			}
		})
	}
}

func TestWODService_ListStandardWODs(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*mockWODRepo)
		expectedCount int
	}{
		{
			name: "list standard WODs",
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					IsStandard: true,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "Murph",
					IsStandard: true,
				}
				userID := int64(1)
				m.wods[3] = &domain.WOD{
					ID:         3,
					Name:       "Custom",
					IsStandard: false,
					CreatedBy:  &userID,
				}
			},
			expectedCount: 2,
		},
		{
			name: "no standard WODs",
			setupMock: func(m *mockWODRepo) {
				// No WODs
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			wods, err := service.ListStandard(100, 0)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wods) != tt.expectedCount {
				t.Errorf("expected %d WODs, got %d", tt.expectedCount, len(wods))
			}

			// Verify all returned WODs are standard
			for _, wod := range wods {
				if !wod.IsStandard {
					t.Error("non-standard WOD returned in standard list")
				}
			}
		})
	}
}

func TestWODService_ListUserWODs(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		setupMock     func(*mockWODRepo)
		expectedCount int
	}{
		{
			name:   "list user's custom WODs",
			userID: 1,
			setupMock: func(m *mockWODRepo) {
				userID1 := int64(1)
				userID2 := int64(2)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Custom 1",
					IsStandard: false,
					CreatedBy:  &userID1,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "Custom 2",
					IsStandard: false,
					CreatedBy:  &userID1,
				}
				m.wods[3] = &domain.WOD{
					ID:         3,
					Name:       "Other User WOD",
					IsStandard: false,
					CreatedBy:  &userID2,
				}
				m.wods[4] = &domain.WOD{
					ID:         4,
					Name:       "Fran",
					IsStandard: true,
				}
			},
			expectedCount: 2,
		},
		{
			name:   "no custom WODs for user",
			userID: 1,
			setupMock: func(m *mockWODRepo) {
				// No WODs for this user
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			wods, err := service.ListByUser(tt.userID, 100, 0)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wods) != tt.expectedCount {
				t.Errorf("expected %d WODs, got %d", tt.expectedCount, len(wods))
			}

			// Verify all returned WODs belong to the user
			for _, wod := range wods {
				if wod.CreatedBy == nil || *wod.CreatedBy != tt.userID {
					t.Errorf("WOD does not belong to user %d", tt.userID)
				}
			}
		})
	}
}

func TestWODService_ListAllWODs(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		setupMock     func(*mockWODRepo)
		expectedCount int
	}{
		{
			name:   "list all WODs (standard + user's custom)",
			userID: 1,
			setupMock: func(m *mockWODRepo) {
				userID1 := int64(1)
				userID2 := int64(2)
				// Standard WODs
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					IsStandard: true,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "Murph",
					IsStandard: true,
				}
				// User 1 custom WODs
				m.wods[3] = &domain.WOD{
					ID:         3,
					Name:       "Custom 1",
					IsStandard: false,
					CreatedBy:  &userID1,
				}
				// User 2 custom WOD (should not be included)
				m.wods[4] = &domain.WOD{
					ID:         4,
					Name:       "Other Custom",
					IsStandard: false,
					CreatedBy:  &userID2,
				}
			},
			expectedCount: 3, // 2 standard + 1 user custom
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			wods, err := service.ListAll(&tt.userID, 100, 0)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wods) != tt.expectedCount {
				t.Errorf("expected %d WODs, got %d", tt.expectedCount, len(wods))
			}
		})
	}
}

func TestWODService_SearchWODs(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		setupMock     func(*mockWODRepo)
		expectedCount int
	}{
		{
			name:  "search by partial name",
			query: "Fran",
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					IsStandard: true,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "Murph",
					IsStandard: true,
				}
				m.wods[3] = &domain.WOD{
					ID:         3,
					Name:       "Francesca", // Contains "Fran"
					IsStandard: false,
				}
			},
			expectedCount: 2, // "Fran" and "Francesca"
		},
		{
			name:  "empty query returns empty (by design)",
			query: "",
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					IsStandard: true,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "Murph",
					IsStandard: true,
				}
			},
			expectedCount: 0, // Service returns empty for empty query
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			wods, err := service.Search(tt.query, 100)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wods) != tt.expectedCount {
				t.Errorf("expected %d WODs, got %d", tt.expectedCount, len(wods))
			}
		})
	}
}

func TestWODService_UpdateWOD(t *testing.T) {
	tests := []struct {
		name          string
		wodID         int64
		userID        int64
		updates       *domain.WOD
		setupMock     func(*mockWODRepo)
		expectedError error
	}{
		{
			name:   "successful update",
			wodID:  1,
			userID: 1,
			updates: &domain.WOD{
				Name:        "Updated WOD",
				Source:      "Self-recorded",
				Type:        "Self-created",
				Regime:      "AMRAP",
				ScoreType:   "Rounds+Reps",
				Description: "Updated description",
			},
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:          1,
					Name:        "Original WOD",
					Source:      "Self-recorded",
					Type:        "Self-created",
					IsStandard:  false,
					CreatedBy:   &userID,
					Description: "Original description",
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-24 * time.Hour),
				}
			},
			expectedError: nil,
		},
		{
			name:   "unauthorized update (different user)",
			wodID:  1,
			userID: 2,
			updates: &domain.WOD{
				Name:      "Updated WOD",
				Source:    "Self-recorded",
				Type:      "Self-created",
				Regime:    "AMRAP",
				ScoreType: "Rounds+Reps",
			},
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Original WOD",
					Source:     "Self-recorded",
					Type:       "Self-created",
					IsStandard: false,
					CreatedBy:  &userID,
				}
			},
			expectedError: ErrWODOwnership,
		},
		{
			name:   "cannot update standard WOD",
			wodID:  1,
			userID: 1,
			updates: &domain.WOD{
				Name:      "Updated WOD",
				Source:    "Self-recorded",
				Type:      "Self-created",
				Regime:    "AMRAP",
				ScoreType: "Rounds+Reps",
			},
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					Source:     "CrossFit",
					Type:       "Benchmark",
					IsStandard: true,
					CreatedBy:  &userID,
				}
			},
			expectedError: ErrWODUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			tt.updates.ID = tt.wodID
			err := service.Update(tt.updates, tt.userID, "test@test.com")

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
			updated := wodRepo.wods[tt.wodID]
			if updated.Name != tt.updates.Name {
				t.Errorf("name not updated correctly")
			}
			if updated.Description != tt.updates.Description {
				t.Errorf("description not updated correctly")
			}
		})
	}
}

func TestWODService_DeleteWOD(t *testing.T) {
	tests := []struct {
		name          string
		wodID         int64
		userID        int64
		setupMock     func(*mockWODRepo)
		expectedError error
	}{
		{
			name:   "successful deletion",
			wodID:  1,
			userID: 1,
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Custom WOD",
					IsStandard: false,
					CreatedBy:  &userID,
				}
			},
			expectedError: nil,
		},
		{
			name:   "unauthorized deletion (different user)",
			wodID:  1,
			userID: 2,
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Custom WOD",
					IsStandard: false,
					CreatedBy:  &userID,
				}
			},
			expectedError: ErrWODOwnership,
		},
		{
			name:   "cannot delete standard WOD",
			wodID:  1,
			userID: 1,
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					IsStandard: true,
					CreatedBy:  &userID,
				}
			},
			expectedError: ErrWODUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			err := service.Delete(tt.wodID, tt.userID, "test@test.com")

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
			if _, exists := wodRepo.wods[tt.wodID]; exists {
				t.Error("WOD should have been deleted")
			}
		})
	}
}

// TestWODService_List tests the List function with filters
func TestWODService_List(t *testing.T) {
	tests := []struct {
		name          string
		filters       map[string]interface{}
		limit         int
		offset        int
		setupMock     func(*mockWODRepo)
		expectedCount int
		expectError   bool
	}{
		{
			name:    "list all WODs",
			filters: nil,
			limit:   100,
			offset:  0,
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{ID: 1, Name: "Fran", IsStandard: true}
				m.wods[2] = &domain.WOD{ID: 2, Name: "Murph", IsStandard: true}
				userID := int64(1)
				m.wods[3] = &domain.WOD{ID: 3, Name: "Custom", IsStandard: false, CreatedBy: &userID}
			},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:    "list with standard filter true",
			filters: map[string]interface{}{"is_standard": true},
			limit:   100,
			offset:  0,
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{ID: 1, Name: "Fran", IsStandard: true}
				m.wods[2] = &domain.WOD{ID: 2, Name: "Murph", IsStandard: true}
				userID := int64(1)
				m.wods[3] = &domain.WOD{ID: 3, Name: "Custom", IsStandard: false, CreatedBy: &userID}
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "list empty",
			filters:       nil,
			limit:         100,
			offset:        0,
			setupMock:     func(m *mockWODRepo) {},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)
			wods, err := service.List(tt.filters, tt.limit, tt.offset)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wods) != tt.expectedCount {
				t.Errorf("expected %d WODs, got %d", tt.expectedCount, len(wods))
			}
		})
	}
}

// TestWODService_Count tests the Count function
func TestWODService_Count(t *testing.T) {
	tests := []struct {
		name          string
		isStandard    *bool
		setupMock     func(*mockWODRepo)
		expectedCount int64
		expectError   bool
	}{
		{
			name:       "count all WODs",
			isStandard: nil,
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{ID: 1, Name: "Fran", IsStandard: true}
				m.wods[2] = &domain.WOD{ID: 2, Name: "Murph", IsStandard: true}
				userID := int64(1)
				m.wods[3] = &domain.WOD{ID: 3, Name: "Custom", IsStandard: false, CreatedBy: &userID}
			},
			expectedCount: 3,
			expectError:   false,
		},
		{
			name:       "count standard WODs",
			isStandard: boolPtr(true),
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{ID: 1, Name: "Fran", IsStandard: true}
				m.wods[2] = &domain.WOD{ID: 2, Name: "Murph", IsStandard: true}
				userID := int64(1)
				m.wods[3] = &domain.WOD{ID: 3, Name: "Custom", IsStandard: false, CreatedBy: &userID}
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:       "count custom WODs",
			isStandard: boolPtr(false),
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{ID: 1, Name: "Fran", IsStandard: true}
				m.wods[2] = &domain.WOD{ID: 2, Name: "Murph", IsStandard: true}
				userID := int64(1)
				m.wods[3] = &domain.WOD{ID: 3, Name: "Custom", IsStandard: false, CreatedBy: &userID}
			},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "count empty",
			isStandard:    nil,
			setupMock:     func(m *mockWODRepo) {},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)
			count, err := service.Count(tt.isStandard)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

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

// TestWODService_UpdateAsAdmin tests admin update functionality
func TestWODService_UpdateAsAdmin(t *testing.T) {
	tests := []struct {
		name          string
		wodID         int64
		userID        int64
		updates       *domain.WOD
		setupMock     func(*mockWODRepo)
		expectedError error
	}{
		{
			name:   "admin update standard WOD",
			wodID:  1,
			userID: 1,
			updates: &domain.WOD{
				Name:        "Updated Fran",
				Source:      "CrossFit",
				Type:        "Benchmark",
				Description: "Updated description",
			},
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{
					ID:          1,
					Name:        "Fran",
					Source:      "CrossFit",
					Type:        "Benchmark",
					IsStandard:  true,
					Description: "21-15-9: Thrusters, Pull-ups",
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-24 * time.Hour),
				}
			},
			expectedError: nil,
		},
		{
			name:   "admin update custom WOD",
			wodID:  2,
			userID: 1,
			updates: &domain.WOD{
				Name:        "Updated Custom",
				Source:      "Self-recorded",
				Type:        "Self-created",
				Description: "Updated custom description",
			},
			setupMock: func(m *mockWODRepo) {
				userID := int64(2) // Different user created it
				m.wods[2] = &domain.WOD{
					ID:          2,
					Name:        "Custom WOD",
					Source:      "Self-recorded",
					Type:        "Self-created",
					IsStandard:  false,
					CreatedBy:   &userID,
					Description: "Original custom description",
					CreatedAt:   time.Now().Add(-24 * time.Hour),
					UpdatedAt:   time.Now().Add(-24 * time.Hour),
				}
			},
			expectedError: nil,
		},
		{
			name:   "admin update WOD not found",
			wodID:  999,
			userID: 1,
			updates: &domain.WOD{
				Name:   "Updated",
				Source: "CrossFit",
				Type:   "Benchmark",
			},
			setupMock:     func(m *mockWODRepo) {},
			expectedError: ErrWODNotFound,
		},
		{
			name:   "admin update duplicate name",
			wodID:  1,
			userID: 1,
			updates: &domain.WOD{
				Name:   "Murph", // Already exists
				Source: "CrossFit",
				Type:   "Benchmark",
			},
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					Source:     "CrossFit",
					Type:       "Benchmark",
					IsStandard: true,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "Murph",
					Source:     "CrossFit",
					Type:       "Hero",
					IsStandard: true,
				}
			},
			expectedError: ErrWODDuplicateName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)

			tt.updates.ID = tt.wodID
			err := service.UpdateAsAdmin(tt.updates, tt.userID, "admin@test.com")

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
			updated := wodRepo.wods[tt.wodID]
			if updated.Name != tt.updates.Name {
				t.Errorf("name not updated correctly")
			}
		})
	}
}

// TestWODService_ListAllUserCreated tests listing all user-created WODs
func TestWODService_ListAllUserCreated(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		offset        int
		setupMock     func(*mockWODRepo)
		expectedCount int
		expectError   bool
	}{
		{
			name:   "list all user-created WODs",
			limit:  100,
			offset: 0,
			setupMock: func(m *mockWODRepo) {
				m.wods[1] = &domain.WOD{ID: 1, Name: "Fran", IsStandard: true}
				userID1 := int64(1)
				userID2 := int64(2)
				m.wods[2] = &domain.WOD{ID: 2, Name: "Custom1", IsStandard: false, CreatedBy: &userID1}
				m.wods[3] = &domain.WOD{ID: 3, Name: "Custom2", IsStandard: false, CreatedBy: &userID2}
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "no user-created WODs",
			limit:         100,
			offset:        0,
			setupMock:     func(m *mockWODRepo) {},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)
			wods, count, err := service.ListAllUserCreated(tt.limit, tt.offset)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wods) != tt.expectedCount {
				t.Errorf("expected %d WODs, got %d", tt.expectedCount, len(wods))
			}

			if count != int64(tt.expectedCount) {
				t.Errorf("expected count %d, got %d", tt.expectedCount, count)
			}
		})
	}
}

// TestWODService_ListAllUserCreatedWithUserInfo tests listing user-created WODs with creator info
func TestWODService_ListAllUserCreatedWithUserInfo(t *testing.T) {
	wodRepo := newMockWODRepo()
	userID1 := int64(1)
	wodRepo.wods[1] = &domain.WOD{ID: 1, Name: "Custom1", IsStandard: false, CreatedBy: &userID1}

	service := NewWODService(wodRepo, nil, nil)
	wods, count, err := service.ListAllUserCreatedWithUserInfo(100, 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// The mock returns empty but count comes from CountAllUserCreated
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// The mock returns empty list
	if wods == nil {
		t.Error("expected non-nil list")
	}
}

// TestWODService_ListAllUserCreatedWithUserInfoFiltered tests filtered listing
func TestWODService_ListAllUserCreatedWithUserInfoFiltered(t *testing.T) {
	wodRepo := newMockWODRepo()
	userID1 := int64(1)
	wodRepo.wods[1] = &domain.WOD{ID: 1, Name: "Custom1", IsStandard: false, CreatedBy: &userID1, ScoreType: "Time (HH:MM:SS)"}

	service := NewWODService(wodRepo, nil, nil)
	wods, count, err := service.ListAllUserCreatedWithUserInfoFiltered(100, 0, "Custom", "", "")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// The mock returns 0 count
	if count != 0 {
		t.Errorf("expected count 0 (mock returns 0), got %d", count)
	}

	// The mock returns empty list
	if wods == nil {
		t.Error("expected non-nil list")
	}
}

// TestWODService_CopyToStandard tests copying a user WOD to standard
func TestWODService_CopyToStandard(t *testing.T) {
	tests := []struct {
		name        string
		wodID       int64
		newName     string
		setupMock   func(*mockWODRepo)
		expectError error
	}{
		{
			name:    "successful copy to standard",
			wodID:   1,
			newName: "New Standard WOD",
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:          1,
					Name:        "Custom WOD",
					Source:      "Self-recorded",
					Type:        "Self-created",
					IsStandard:  false,
					CreatedBy:   &userID,
					Description: "Test description",
				}
			},
			expectError: nil,
		},
		{
			name:        "empty name",
			wodID:       1,
			newName:     "",
			setupMock:   func(m *mockWODRepo) {},
			expectError: ErrWODNameRequired,
		},
		{
			name:    "duplicate standard WOD name",
			wodID:   1,
			newName: "Fran",
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Custom WOD",
					IsStandard: false,
					CreatedBy:  &userID,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "Fran",
					IsStandard: true, // Standard WOD with same name
				}
			},
			expectError: ErrWODDuplicateName,
		},
		{
			name:    "same name as user WOD is OK",
			wodID:   1,
			newName: "User Custom",
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Original Custom",
					IsStandard: false,
					CreatedBy:  &userID,
				}
				m.wods[2] = &domain.WOD{
					ID:         2,
					Name:       "User Custom",
					IsStandard: false, // Non-standard WOD with same name - should be OK
					CreatedBy:  &userID,
				}
			},
			expectError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil, nil)
			wod, err := service.CopyToStandard(tt.wodID, tt.newName)

			if tt.expectError != nil {
				if !errors.Is(err, tt.expectError) {
					t.Errorf("expected error %v, got %v", tt.expectError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if wod == nil {
				t.Error("expected WOD, got nil")
				return
			}

			if wod.Name != tt.newName {
				t.Errorf("expected name %s, got %s", tt.newName, wod.Name)
			}

			if !wod.IsStandard {
				t.Error("copied WOD should be standard")
			}
		})
	}
}

// TestWODService_ValidateWOD tests WOD validation
func TestWODService_ValidateWOD(t *testing.T) {
	tests := []struct {
		name        string
		wod         *domain.WOD
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid WOD",
			wod: &domain.WOD{
				Name:   "Test WOD",
				Source: "CrossFit",
				Type:   "Benchmark",
			},
			expectError: false,
		},
		{
			name: "missing name",
			wod: &domain.WOD{
				Name:   "",
				Source: "CrossFit",
				Type:   "Benchmark",
			},
			expectError: true,
			errorMsg:    "wod name is required",
		},
		{
			name: "missing source",
			wod: &domain.WOD{
				Name:   "Test WOD",
				Source: "",
				Type:   "Benchmark",
			},
			expectError: true,
			errorMsg:    "wod source is required",
		},
		{
			name: "missing type",
			wod: &domain.WOD{
				Name:   "Test WOD",
				Source: "CrossFit",
				Type:   "",
			},
			expectError: true,
			errorMsg:    "wod type is required",
		},
		{
			name: "invalid source",
			wod: &domain.WOD{
				Name:   "Test WOD",
				Source: "Invalid Source",
				Type:   "Benchmark",
			},
			expectError: true,
			errorMsg:    "invalid source",
		},
		{
			name: "invalid type",
			wod: &domain.WOD{
				Name:   "Test WOD",
				Source: "CrossFit",
				Type:   "Invalid Type",
			},
			expectError: true,
			errorMsg:    "invalid type",
		},
		{
			name: "invalid regime",
			wod: &domain.WOD{
				Name:   "Test WOD",
				Source: "CrossFit",
				Type:   "Benchmark",
				Regime: "Invalid Regime",
			},
			expectError: true,
			errorMsg:    "invalid regime",
		},
		{
			name: "invalid score type",
			wod: &domain.WOD{
				Name:      "Test WOD",
				Source:    "CrossFit",
				Type:      "Benchmark",
				ScoreType: "Invalid Score Type",
			},
			expectError: true,
			errorMsg:    "invalid score type",
		},
		{
			name: "valid with all optional fields",
			wod: &domain.WOD{
				Name:        "Complete WOD",
				Source:      "CrossFit",
				Type:        "Hero",
				Regime:      "AMRAP",
				ScoreType:   "Rounds+Reps",
				Description: "Full workout description",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			service := NewWODService(wodRepo, nil, nil)

			err := service.Create(tt.wod, 1, "test@test.com")

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorMsg != "" && !errors.Is(err, ErrWODNameRequired) &&
					!errors.Is(err, ErrWODSourceRequired) &&
					!errors.Is(err, ErrWODTypeRequired) {
					// For custom validation errors, check the message
					if err.Error() != tt.errorMsg && !strings.Contains(err.Error(), tt.errorMsg) {
						t.Errorf("expected error containing '%s', got '%v'", tt.errorMsg, err)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestWODService_GetByName_EmptyName tests GetByName with empty name
func TestWODService_GetByName_EmptyName(t *testing.T) {
	wodRepo := newMockWODRepo()
	service := NewWODService(wodRepo, nil, nil)

	_, err := service.GetByName("")
	if !errors.Is(err, ErrWODNameRequired) {
		t.Errorf("expected ErrWODNameRequired, got %v", err)
	}

	_, err = service.GetByName("   ") // Whitespace only
	if !errors.Is(err, ErrWODNameRequired) {
		t.Errorf("expected ErrWODNameRequired for whitespace, got %v", err)
	}
}

// TestWODService_DeleteWOD_NotFound tests deletion of non-existent WOD
func TestWODService_DeleteWOD_NotFound(t *testing.T) {
	wodRepo := newMockWODRepo()
	service := NewWODService(wodRepo, nil, nil)

	err := service.Delete(999, 1, "test@test.com")
	if !errors.Is(err, ErrWODNotFound) {
		t.Errorf("expected ErrWODNotFound, got %v", err)
	}
}

// TestWODService_DeleteWOD_NilCreatedBy tests deletion when CreatedBy is nil
func TestWODService_DeleteWOD_NilCreatedBy(t *testing.T) {
	wodRepo := newMockWODRepo()
	wodRepo.wods[1] = &domain.WOD{
		ID:         1,
		Name:       "Orphan WOD",
		IsStandard: false,
		CreatedBy:  nil, // No creator
	}

	service := NewWODService(wodRepo, nil, nil)

	err := service.Delete(1, 1, "test@test.com")
	if !errors.Is(err, ErrWODOwnership) {
		t.Errorf("expected ErrWODOwnership, got %v", err)
	}
}

// TestWODService_ListAll_NilUserID tests ListAll with nil user ID
func TestWODService_ListAll_NilUserID(t *testing.T) {
	wodRepo := newMockWODRepo()
	wodRepo.wods[1] = &domain.WOD{ID: 1, Name: "Fran", IsStandard: true}
	wodRepo.wods[2] = &domain.WOD{ID: 2, Name: "Murph", IsStandard: true}

	service := NewWODService(wodRepo, nil, nil)

	wods, err := service.ListAll(nil, 100, 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Should only return standard WODs
	if len(wods) != 2 {
		t.Errorf("expected 2 WODs, got %d", len(wods))
	}
}

// TestWODService_PaginateWODs tests the pagination helper
func TestWODService_PaginateWODs(t *testing.T) {
	tests := []struct {
		name          string
		wodCount      int
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "default limit with no limit specified",
			wodCount:      150,
			limit:         0, // Should default to 100
			offset:        0,
			expectedCount: 100,
		},
		{
			name:          "offset beyond list",
			wodCount:      10,
			limit:         100,
			offset:        20,
			expectedCount: 0,
		},
		{
			name:          "negative offset treated as 0",
			wodCount:      10,
			limit:         100,
			offset:        -5,
			expectedCount: 10,
		},
		{
			name:          "partial page at end",
			wodCount:      25,
			limit:         10,
			offset:        20,
			expectedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			for i := 1; i <= tt.wodCount; i++ {
				wodRepo.wods[int64(i)] = &domain.WOD{
					ID:         int64(i),
					Name:       "Fran",
					IsStandard: true,
				}
			}

			service := NewWODService(wodRepo, nil, nil)
			wods, err := service.ListAll(nil, tt.limit, tt.offset)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wods) != tt.expectedCount {
				t.Errorf("expected %d WODs, got %d", tt.expectedCount, len(wods))
			}
		})
	}
}

// TestWODService_Create_DuplicateName tests creating WOD with duplicate name
func TestWODService_Create_DuplicateName(t *testing.T) {
	wodRepo := newMockWODRepo()
	wodRepo.wods[1] = &domain.WOD{
		ID:         1,
		Name:       "Existing WOD",
		IsStandard: true,
	}

	service := NewWODService(wodRepo, nil, nil)

	err := service.Create(&domain.WOD{
		Name:   "Existing WOD",
		Source: "Self-recorded",
		Type:   "Self-created",
	}, 1, "test@test.com")

	if !errors.Is(err, ErrWODDuplicateName) {
		t.Errorf("expected ErrWODDuplicateName, got %v", err)
	}
}

// TestWODService_Update_NotFound tests updating non-existent WOD
func TestWODService_Update_NotFound(t *testing.T) {
	wodRepo := newMockWODRepo()
	service := NewWODService(wodRepo, nil, nil)

	err := service.Update(&domain.WOD{
		ID:     999,
		Name:   "Test",
		Source: "Self-recorded",
		Type:   "Self-created",
	}, 1, "test@test.com")

	if !errors.Is(err, ErrWODNotFound) {
		t.Errorf("expected ErrWODNotFound, got %v", err)
	}
}

// TestWODService_Update_DuplicateName tests updating WOD with duplicate name
func TestWODService_Update_DuplicateName(t *testing.T) {
	wodRepo := newMockWODRepo()
	userID := int64(1)
	wodRepo.wods[1] = &domain.WOD{
		ID:         1,
		Name:       "WOD 1",
		Source:     "Self-recorded",
		Type:       "Self-created",
		IsStandard: false,
		CreatedBy:  &userID,
	}
	wodRepo.wods[2] = &domain.WOD{
		ID:         2,
		Name:       "WOD 2",
		Source:     "Self-recorded",
		Type:       "Self-created",
		IsStandard: false,
		CreatedBy:  &userID,
	}

	service := NewWODService(wodRepo, nil, nil)

	err := service.Update(&domain.WOD{
		ID:     1,
		Name:   "WOD 2", // Trying to rename to existing name
		Source: "Self-recorded",
		Type:   "Self-created",
	}, 1, "test@test.com")

	if !errors.Is(err, ErrWODDuplicateName) {
		t.Errorf("expected ErrWODDuplicateName, got %v", err)
	}
}

// TestWODService_Update_NilCreatedBy tests updating WOD with nil CreatedBy
func TestWODService_Update_NilCreatedBy(t *testing.T) {
	wodRepo := newMockWODRepo()
	wodRepo.wods[1] = &domain.WOD{
		ID:         1,
		Name:       "Orphan WOD",
		Source:     "Self-recorded",
		Type:       "Self-created",
		IsStandard: false,
		CreatedBy:  nil, // No creator
	}

	service := NewWODService(wodRepo, nil, nil)

	err := service.Update(&domain.WOD{
		ID:     1,
		Name:   "Updated Orphan",
		Source: "Self-recorded",
		Type:   "Self-created",
	}, 1, "test@test.com")

	if !errors.Is(err, ErrWODOwnership) {
		t.Errorf("expected ErrWODOwnership, got %v", err)
	}
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
