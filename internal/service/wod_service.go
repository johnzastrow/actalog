package service

import (
	"errors"
	"fmt"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	// ErrWODNotFound is returned when a WOD is not found
	ErrWODNotFound = errors.New("WOD not found")
	// ErrUnauthorizedWODAccess is returned when a user tries to modify a WOD they don't own
	ErrUnauthorizedWODAccess = errors.New("unauthorized WOD access")
)

// WODService handles business logic for WODs (Workout of the Day benchmarks)
type WODService struct {
	wodRepo domain.WODRepository
}

// NewWODService creates a new WOD service
func NewWODService(wodRepo domain.WODRepository) *WODService {
	return &WODService{
		wodRepo: wodRepo,
	}
}

// CreateWOD creates a new custom WOD
func (s *WODService) CreateWOD(userID int64, wod *domain.WOD) error {
	// Set created_by to user (custom WODs)
	wod.CreatedBy = &userID

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
		return nil, ErrWODNotFound
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
		return nil, ErrWODNotFound
	}

	return wod, nil
}

// ListStandardWODs lists all standard CrossFit benchmark WODs (created_by IS NULL)
func (s *WODService) ListStandardWODs() ([]*domain.WOD, error) {
	wods, err := s.wodRepo.ListStandard()
	if err != nil {
		return nil, fmt.Errorf("failed to list standard WODs: %w", err)
	}

	return wods, nil
}

// ListUserWODs lists all custom WODs created by a specific user
func (s *WODService) ListUserWODs(userID int64) ([]*domain.WOD, error) {
	wods, err := s.wodRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user WODs: %w", err)
	}

	return wods, nil
}

// ListAllWODs lists all WODs available to a user (standard + user's custom)
func (s *WODService) ListAllWODs(userID int64) ([]*domain.WOD, error) {
	// Get standard WODs
	standard, err := s.wodRepo.ListStandard()
	if err != nil {
		return nil, fmt.Errorf("failed to list standard WODs: %w", err)
	}

	// Get user's custom WODs
	custom, err := s.wodRepo.ListByUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user WODs: %w", err)
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

// UpdateWOD updates a custom WOD (only if user is the creator)
func (s *WODService) UpdateWOD(wodID, userID int64, updates *domain.WOD) error {
	// Get existing WOD
	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return ErrWODNotFound
	}

	// Authorization check - can only update custom WODs that user created
	if wod.CreatedBy == nil || *wod.CreatedBy != userID {
		return ErrUnauthorizedWODAccess
	}

	// Update fields
	wod.Name = updates.Name
	wod.Description = updates.Description
	wod.Source = updates.Source
	wod.Type = updates.Type
	wod.Regime = updates.Regime
	wod.ScoreType = updates.ScoreType

	err = s.wodRepo.Update(wod)
	if err != nil {
		return fmt.Errorf("failed to update WOD: %w", err)
	}

	return nil
}

// DeleteWOD deletes a custom WOD (only if user is the creator)
func (s *WODService) DeleteWOD(wodID, userID int64) error {
	// Get existing WOD
	wod, err := s.wodRepo.GetByID(wodID)
	if err != nil {
		return fmt.Errorf("failed to get WOD: %w", err)
	}
	if wod == nil {
		return ErrWODNotFound
	}

	// Authorization check - can only delete custom WODs that user created
	if wod.CreatedBy == nil || *wod.CreatedBy != userID {
		return ErrUnauthorizedWODAccess
	}

	err = s.wodRepo.Delete(wodID)
	if err != nil {
		return fmt.Errorf("failed to delete WOD: %w", err)
	}

	return nil
}
