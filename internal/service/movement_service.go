package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrMovementNotFound     = errors.New("movement not found")
	ErrMovementUnauthorized = errors.New("unauthorized: cannot modify standard movement")
	ErrMovementNameRequired = errors.New("movement name is required")
	ErrMovementTypeRequired = errors.New("movement type is required")
)

// MovementService handles movement business logic
type MovementService struct {
	movementRepo         domain.MovementRepository
	dataChangeLogService *DataChangeLogService
}

// NewMovementService creates a new movement service
func NewMovementService(movementRepo domain.MovementRepository, dataChangeLogService *DataChangeLogService) *MovementService {
	return &MovementService{
		movementRepo:         movementRepo,
		dataChangeLogService: dataChangeLogService,
	}
}

// Create creates a new custom movement
func (s *MovementService) Create(movement *domain.Movement) error {
	// Validate required fields
	if err := s.validateMovement(movement); err != nil {
		return err
	}

	// Set as custom movement
	movement.IsStandard = false
	now := time.Now()
	movement.CreatedAt = now
	movement.UpdatedAt = now

	return s.movementRepo.Create(movement)
}

// GetByID retrieves a movement by ID
func (s *MovementService) GetByID(id int64) (*domain.Movement, error) {
	movement, err := s.movementRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get movement: %w", err)
	}
	if movement == nil {
		return nil, ErrMovementNotFound
	}
	return movement, nil
}

// GetByName retrieves a movement by name
func (s *MovementService) GetByName(name string) (*domain.Movement, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrMovementNameRequired
	}
	return s.movementRepo.GetByName(name)
}

// ListAll retrieves all movements
func (s *MovementService) ListAll() ([]*domain.Movement, error) {
	return s.movementRepo.ListAll()
}

// ListStandard retrieves all standard movements
func (s *MovementService) ListStandard() ([]*domain.Movement, error) {
	return s.movementRepo.ListStandard()
}

// Search searches movements by name
func (s *MovementService) Search(query string, limit int) ([]*domain.Movement, error) {
	if strings.TrimSpace(query) == "" {
		return []*domain.Movement{}, nil
	}
	return s.movementRepo.Search(query, limit)
}

// Update updates an existing movement with data change logging
func (s *MovementService) Update(movement *domain.Movement, userID int64, userEmail string) error {
	// Validate required fields
	if err := s.validateMovement(movement); err != nil {
		return err
	}

	// Get existing movement for before values
	existing, err := s.movementRepo.GetByID(movement.ID)
	if err != nil {
		return fmt.Errorf("failed to get movement: %w", err)
	}
	if existing == nil {
		return ErrMovementNotFound
	}

	// Check if it's a standard movement (only admins can modify these)
	if existing.IsStandard {
		return ErrMovementUnauthorized
	}

	// Update timestamp
	movement.UpdatedAt = time.Now()
	movement.CreatedAt = existing.CreatedAt
	movement.IsStandard = existing.IsStandard

	// Perform the update
	if err := s.movementRepo.Update(movement); err != nil {
		return fmt.Errorf("failed to update movement: %w", err)
	}

	// Log the change
	if s.dataChangeLogService != nil {
		if logErr := s.dataChangeLogService.LogMovementUpdate(movement.ID, movement.Name, userID, userEmail, existing, movement, nil, nil); logErr != nil {
			fmt.Printf("Warning: failed to log movement update: %v\n", logErr)
		}
	}

	return nil
}

// UpdateAsAdmin updates any movement (for admin use)
func (s *MovementService) UpdateAsAdmin(movement *domain.Movement, userID int64, userEmail string) error {
	// Validate required fields
	if err := s.validateMovement(movement); err != nil {
		return err
	}

	// Get existing movement for before values
	existing, err := s.movementRepo.GetByID(movement.ID)
	if err != nil {
		return fmt.Errorf("failed to get movement: %w", err)
	}
	if existing == nil {
		return ErrMovementNotFound
	}

	// Update timestamp
	movement.UpdatedAt = time.Now()
	movement.CreatedAt = existing.CreatedAt
	movement.IsStandard = existing.IsStandard

	// Perform the update
	if err := s.movementRepo.Update(movement); err != nil {
		return fmt.Errorf("failed to update movement: %w", err)
	}

	// Log the change
	if s.dataChangeLogService != nil {
		if logErr := s.dataChangeLogService.LogMovementUpdate(movement.ID, movement.Name, userID, userEmail, existing, movement, nil, nil); logErr != nil {
			fmt.Printf("Warning: failed to log movement update: %v\n", logErr)
		}
	}

	return nil
}

// Delete deletes a movement with data change logging
func (s *MovementService) Delete(id int64, userID int64, userEmail string) error {
	// Get existing movement for before values
	existing, err := s.movementRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get movement: %w", err)
	}
	if existing == nil {
		return ErrMovementNotFound
	}

	// Check if it's a standard movement
	if existing.IsStandard {
		return ErrMovementUnauthorized
	}

	// Delete the movement
	if err := s.movementRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete movement: %w", err)
	}

	// Log the deletion
	if s.dataChangeLogService != nil {
		if logErr := s.dataChangeLogService.LogMovementDelete(id, existing.Name, userID, userEmail, existing, nil, nil); logErr != nil {
			fmt.Printf("Warning: failed to log movement delete: %v\n", logErr)
		}
	}

	return nil
}

// ListAllUserCreated retrieves all user-created movements across all users (admin only)
func (s *MovementService) ListAllUserCreated() ([]*domain.Movement, int64, error) {
	// Get the list
	movements, err := s.movementRepo.ListAllUserCreated()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all user-created movements: %w", err)
	}

	// Get the count
	count, err := s.movementRepo.CountAllUserCreated()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user-created movements: %w", err)
	}

	return movements, count, nil
}

// ListAllUserCreatedWithUserInfo retrieves all user-created movements with creator info (admin only)
func (s *MovementService) ListAllUserCreatedWithUserInfo() ([]*domain.MovementWithCreator, int64, error) {
	// Get the list with user info
	movements, err := s.movementRepo.ListAllUserCreatedWithUserInfo()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all user-created movements with user info: %w", err)
	}

	// Get the count
	count, err := s.movementRepo.CountAllUserCreated()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user-created movements: %w", err)
	}

	return movements, count, nil
}

// ListAllUserCreatedWithUserInfoFiltered retrieves all user-created movements with creator info and filters (admin only)
func (s *MovementService) ListAllUserCreatedWithUserInfoFiltered(limit, offset int, search, movementType, creator string) ([]*domain.MovementWithCreator, int64, error) {
	// Get the list with user info and filters
	movements, count, err := s.movementRepo.ListAllUserCreatedWithUserInfoFiltered(limit, offset, search, movementType, creator)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list all user-created movements with filters: %w", err)
	}

	return movements, count, nil
}

// CopyToStandard creates a standard movement from a user-created one (admin only)
func (s *MovementService) CopyToStandard(id int64, newName string) (*domain.Movement, error) {
	// Validate new name
	if strings.TrimSpace(newName) == "" {
		return nil, ErrMovementNameRequired
	}

	// Check if a STANDARD movement with this name already exists
	// User-created movements with the same name are OK - we only prevent duplicate standard movements
	existing, err := s.movementRepo.GetByName(newName)
	if err != nil {
		return nil, fmt.Errorf("failed to check for duplicate movement name: %w", err)
	}
	if existing != nil && existing.IsStandard {
		return nil, fmt.Errorf("standard movement with name '%s' already exists", newName)
	}

	// Copy to standard
	movement, err := s.movementRepo.CopyToStandard(id, newName)
	if err != nil {
		return nil, fmt.Errorf("failed to copy movement to standard: %w", err)
	}

	return movement, nil
}

// validateMovement validates movement required fields
func (s *MovementService) validateMovement(movement *domain.Movement) error {
	if strings.TrimSpace(movement.Name) == "" {
		return ErrMovementNameRequired
	}
	if movement.Type == "" {
		return ErrMovementTypeRequired
	}
	return nil
}
