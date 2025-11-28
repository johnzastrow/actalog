package service

import (
	"errors"
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

			service := NewWODService(wodRepo, nil)

			err := service.Create(tt.wod, tt.userID)

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

			service := NewWODService(wodRepo, nil)

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

			service := NewWODService(wodRepo, nil)

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

			service := NewWODService(wodRepo, nil)

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

			service := NewWODService(wodRepo, nil)

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

			service := NewWODService(wodRepo, nil)

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
			name:  "empty query returns all",
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
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil)

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
				Description: "Updated description",
			},
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:          1,
					Name:        "Original WOD",
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
				Name: "Updated WOD",
			},
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Original WOD",
					IsStandard: false,
					CreatedBy:  &userID,
				}
			},
			expectedError: ErrUnauthorized,
		},
		{
			name:   "cannot update standard WOD",
			wodID:  1,
			userID: 1,
			updates: &domain.WOD{
				Name: "Updated WOD",
			},
			setupMock: func(m *mockWODRepo) {
				userID := int64(1)
				m.wods[1] = &domain.WOD{
					ID:         1,
					Name:       "Fran",
					IsStandard: true,
					CreatedBy:  &userID,
				}
			},
			expectedError: nil, // Should get error about standard WODs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil)

			tt.updates.ID = tt.wodID
			err := service.Update(tt.updates, tt.userID, "test@test.com")

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if tt.name == "successful update" {
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
			} else if tt.name == "cannot update standard WOD" {
				if err == nil {
					t.Error("expected error for updating standard WOD, got nil")
				}
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
			expectedError: ErrUnauthorized,
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
			expectedError: nil, // Should get error about standard WODs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wodRepo := newMockWODRepo()
			if tt.setupMock != nil {
				tt.setupMock(wodRepo)
			}

			service := NewWODService(wodRepo, nil)

			err := service.Delete(tt.wodID, tt.userID, "test@test.com")

			if tt.expectedError != nil {
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if tt.name == "successful deletion" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				// Verify deletion
				if _, exists := wodRepo.wods[tt.wodID]; exists {
					t.Error("WOD should have been deleted")
				}
			} else if tt.name == "cannot delete standard WOD" {
				if err == nil {
					t.Error("expected error for deleting standard WOD, got nil")
				}
			}
		})
	}
}
