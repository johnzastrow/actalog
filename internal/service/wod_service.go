package service

import (
	"fmt"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WODService handles WOD (Workout of the Day) management
type WODService struct {
	wodRepo domain.WODRepository
}

// NewWODService creates a new WOD service
func NewWODService(wodRepo domain.WODRepository) *WODService {
	return &WODService{
		wodRepo: wodRepo,
	}
}

// CreateWOD creates a custom WOD for a user
func (s *WODService) CreateWOD(userID int64, wod *domain.WOD) error {
	// Set creator and timestamps
	wod.CreatedBy = &userID
	wod.IsStandard = false
	now := time.Now()
	wod.CreatedAt = now
	wod.UpdatedAt = now

	// Create WOD
	err := s.wodRepo.Create(wod)
	if err != nil {
		return fmt.Errorf("failed to create WOD: %w", err)
	}

	return nil
}

// GetWOD retrieves a WOD by ID
func (s *WODService) GetWOD(wodID int64) (*domain.WOD, error) {
	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return nil, fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return nil, fmt.Errorf("WOD not found")
	}

	return wod, nil
}

// GetWODByName retrieves a WOD by name (e.g., "Fran", "Murph")
func (s *WODService) GetWODByName(name string) (*domain.WOD, error) {
	wod, err := s.wodRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get WOD by name: %w", err)
	}
	if wod == nil {
		return nil, fmt.Errorf("WOD not found")
	}

	return wod, nil
}

// ListStandardWODs retrieves all standard CrossFit benchmark WODs
func (s *WODService) ListStandardWODs() ([]*domain.WOD, error) {
	wods, err := s.wodRepo.ListStandard()
	if err != nil {
		return nil, fmt.Errorf("failed to list standard WODs: %w", err)
	}

	return wods, nil
}

// ListUserWODs retrieves all custom WODs created by a user
func (s *WODService) ListUserWODs(userID int64) ([]*domain.WOD, error) {
	wods, err := s.wodRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user WODs: %w", err)
	}

	return wods, nil
}

// ListAllWODs retrieves combined list of standard and user WODs
func (s *WODService) ListAllWODs(userID int64) ([]*domain.WOD, error) {
	// Get standard WODs
	standard, err := s.ListStandardWODs()
	if err != nil {
		return nil, err
	}

	// Get user's custom WODs
	custom, err := s.ListUserWODs(userID)
	if err != nil {
		return nil, err
	}

	// Combine both lists
	wods := append(standard, custom...)
	return wods, nil
}

// SearchWODs searches for WODs by name
func (s *WODService) SearchWODs(query string) ([]*domain.WOD, error) {
	wods, err := s.wodRepo.Search(query)
	if err != nil {
		return nil, fmt.Errorf("failed to search WODs: %w", err)
	}

	return wods, nil
}

// UpdateWOD updates a custom WOD with authorization check
func (s *WODService) UpdateWOD(wodID, userID int64, updates *domain.WOD) error {
	// Get existing WOD
	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return fmt.Errorf("WOD not found")
	}

	// Authorization check: only creator can modify custom WOD
	if wod.CreatedBy == nil || *wod.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Can't modify standard WODs
	if wod.IsStandard {
		return fmt.Errorf("cannot modify standard WODs")
	}

	// Update fields
	wod.Name = updates.Name
	wod.Source = updates.Source
	wod.Type = updates.Type
	wod.Regime = updates.Regime
	wod.ScoreType = updates.ScoreType
	wod.Description = updates.Description
	wod.UpdatedAt = time.Now()

	err = s.wodRepo.Update(wod)
	if err != nil {
		return fmt.Errorf("failed to update WOD: %w", err)
	}

	return nil
}

// DeleteWOD deletes a custom WOD with authorization check
func (s *WODService) DeleteWOD(wodID, userID int64) error {
	// Get existing WOD
	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return fmt.Errorf("WOD not found")
	}

	// Authorization check: only creator can delete custom WOD
	if wod.CreatedBy == nil || *wod.CreatedBy != userID {
		return ErrUnauthorized
	}

	// Can't delete standard WODs
	if wod.IsStandard {
		return fmt.Errorf("cannot delete standard WODs")
	}

	// Delete WOD
	err = s.wodRepo.Delete(wodID)
	if err != nil {
		return fmt.Errorf("failed to delete WOD: %w", err)
	}

	return nil
}
