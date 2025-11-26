package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

var (
	ErrWODNotFound       = errors.New("wod not found")
	ErrWODUnauthorized   = errors.New("unauthorized: cannot modify standard WOD")
	ErrWODOwnership      = errors.New("unauthorized: not the owner of this WOD")
	ErrWODNameRequired   = errors.New("wod name is required")
	ErrWODSourceRequired = errors.New("wod source is required")
	ErrWODTypeRequired   = errors.New("wod type is required")
	ErrWODDuplicateName  = errors.New("wod with this name already exists")
)

// WODService handles WOD business logic
type WODService struct {
	wodRepo              domain.WODRepository
	dataChangeLogService *DataChangeLogService
}

// NewWODService creates a new WOD service
func NewWODService(wodRepo domain.WODRepository, dataChangeLogService *DataChangeLogService) *WODService {
	return &WODService{
		wodRepo:              wodRepo,
		dataChangeLogService: dataChangeLogService,
	}
}

// Create creates a new custom WOD with validation
func (s *WODService) Create(wod *domain.WOD, userID int64) error {
	// Validate required fields
	if err := s.validateWOD(wod); err != nil {
		return err
	}

	// Check for duplicate name
	existing, err := s.wodRepo.GetByName(wod.Name)
	if err != nil {
		return fmt.Errorf("failed to check for duplicate WOD name: %w", err)
	}
	if existing != nil {
		return ErrWODDuplicateName
	}

	// Set custom WOD attributes
	wod.IsStandard = false
	wod.CreatedBy = &userID
	now := time.Now()
	wod.CreatedAt = now
	wod.UpdatedAt = now

	// Create WOD
	err = s.wodRepo.Create(wod)
	if err != nil {
		return fmt.Errorf("failed to create wod: %w", err)
	}

	return nil
}

// GetByID retrieves a WOD by ID
func (s *WODService) GetByID(id int64) (*domain.WOD, error) {
	wod, err := s.wodRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get wod: %w", err)
	}
	if wod == nil {
		return nil, ErrWODNotFound
	}

	return wod, nil
}

// GetByName retrieves a WOD by name
func (s *WODService) GetByName(name string) (*domain.WOD, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrWODNameRequired
	}

	wod, err := s.wodRepo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get wod by name: %w", err)
	}
	if wod == nil {
		return nil, ErrWODNotFound
	}

	return wod, nil
}

// List retrieves WODs with optional filtering
func (s *WODService) List(filters map[string]interface{}, limit, offset int) ([]*domain.WOD, error) {
	wods, err := s.wodRepo.List(filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list wods: %w", err)
	}

	return wods, nil
}

// ListStandard retrieves all standard (pre-seeded) WODs
func (s *WODService) ListStandard(limit, offset int) ([]*domain.WOD, error) {
	wods, err := s.wodRepo.ListStandard(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard wods: %w", err)
	}

	return wods, nil
}

// ListByUser retrieves all custom WODs created by a specific user
func (s *WODService) ListByUser(userID int64, limit, offset int) ([]*domain.WOD, error) {
	wods, err := s.wodRepo.ListByUser(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user wods: %w", err)
	}

	return wods, nil
}

// ListAll retrieves all WODs (standard + user's custom) - convenience method
func (s *WODService) ListAll(userID *int64, limit, offset int) ([]*domain.WOD, error) {
	// Get standard WODs (with large limit to get all for combining)
	standard, err := s.wodRepo.ListStandard(10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list standard wods: %w", err)
	}

	// If no user ID, return only standard WODs with pagination
	if userID == nil {
		return s.paginateWODs(standard, limit, offset), nil
	}

	// Get user's custom WODs (with large limit to get all for combining)
	custom, err := s.wodRepo.ListByUser(*userID, 10000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list user wods: %w", err)
	}

	// Combine both lists and sort alphabetically by name
	wods := append(standard, custom...)
	sort.Slice(wods, func(i, j int) bool {
		return strings.ToLower(wods[i].Name) < strings.ToLower(wods[j].Name)
	})
	return s.paginateWODs(wods, limit, offset), nil
}

// Update updates an existing custom WOD with authorization checks
func (s *WODService) Update(wod *domain.WOD, userID int64, userEmail string) error {
	// Validate required fields
	if err := s.validateWOD(wod); err != nil {
		return err
	}

	// Get existing WOD
	existing, err := s.wodRepo.GetByID(wod.ID)
	if err != nil {
		return fmt.Errorf("failed to get wod: %w", err)
	}
	if existing == nil {
		return ErrWODNotFound
	}

	// Check if it's a standard WOD
	if existing.IsStandard {
		return ErrWODUnauthorized
	}

	// Check ownership
	if existing.CreatedBy == nil || *existing.CreatedBy != userID {
		return ErrWODOwnership
	}

	// Check for duplicate name (if name changed)
	if existing.Name != wod.Name {
		duplicate, err := s.wodRepo.GetByName(wod.Name)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate WOD name: %w", err)
		}
		if duplicate != nil && duplicate.ID != wod.ID {
			return ErrWODDuplicateName
		}
	}

	// Update timestamp
	wod.UpdatedAt = time.Now()

	// Preserve original creation info
	wod.IsStandard = existing.IsStandard
	wod.CreatedBy = existing.CreatedBy
	wod.CreatedAt = existing.CreatedAt

	// Update WOD
	err = s.wodRepo.Update(wod)
	if err != nil {
		return fmt.Errorf("failed to update wod: %w", err)
	}

	// Log the change (after successful update)
	if s.dataChangeLogService != nil {
		if logErr := s.dataChangeLogService.LogWODUpdate(wod.ID, wod.Name, userID, userEmail, existing, wod, nil, nil); logErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to log WOD update: %v\n", logErr)
		}
	}

	return nil
}

// UpdateAsAdmin updates any WOD (for admin use, can update standard WODs)
func (s *WODService) UpdateAsAdmin(wod *domain.WOD, userID int64, userEmail string) error {
	// Validate required fields
	if err := s.validateWOD(wod); err != nil {
		return err
	}

	// Get existing WOD
	existing, err := s.wodRepo.GetByID(wod.ID)
	if err != nil {
		return fmt.Errorf("failed to get wod: %w", err)
	}
	if existing == nil {
		return ErrWODNotFound
	}

	// Check for duplicate name (if name changed)
	if existing.Name != wod.Name {
		duplicate, err := s.wodRepo.GetByName(wod.Name)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate WOD name: %w", err)
		}
		if duplicate != nil && duplicate.ID != wod.ID {
			return ErrWODDuplicateName
		}
	}

	// Update timestamp
	wod.UpdatedAt = time.Now()

	// Preserve original creation info
	wod.IsStandard = existing.IsStandard
	wod.CreatedBy = existing.CreatedBy
	wod.CreatedAt = existing.CreatedAt

	// Use appropriate update method based on whether WOD is standard
	if existing.IsStandard {
		err = s.wodRepo.UpdateStandard(wod)
	} else {
		err = s.wodRepo.Update(wod)
	}
	if err != nil {
		return fmt.Errorf("failed to update wod: %w", err)
	}

	// Log the change (after successful update)
	if s.dataChangeLogService != nil {
		if logErr := s.dataChangeLogService.LogWODUpdate(wod.ID, wod.Name, userID, userEmail, existing, wod, nil, nil); logErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to log WOD update: %v\n", logErr)
		}
	}

	return nil
}

// Delete deletes a custom WOD with authorization checks
func (s *WODService) Delete(id int64, userID int64, userEmail string) error {
	// Get existing WOD
	wod, err := s.wodRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get wod: %w", err)
	}
	if wod == nil {
		return ErrWODNotFound
	}

	// Check if it's a standard WOD
	if wod.IsStandard {
		return ErrWODUnauthorized
	}

	// Check ownership
	if wod.CreatedBy == nil || *wod.CreatedBy != userID {
		return ErrWODOwnership
	}

	// Delete WOD
	err = s.wodRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete wod: %w", err)
	}

	// Log the deletion (after successful delete)
	if s.dataChangeLogService != nil {
		if logErr := s.dataChangeLogService.LogWODDelete(id, wod.Name, userID, userEmail, wod, nil, nil); logErr != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to log WOD delete: %v\n", logErr)
		}
	}

	return nil
}

// Search searches for WODs by name (partial match)
func (s *WODService) Search(query string, limit int) ([]*domain.WOD, error) {
	// Validate query
	if strings.TrimSpace(query) == "" {
		return []*domain.WOD{}, nil
	}

	wods, err := s.wodRepo.Search(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search wods: %w", err)
	}

	return wods, nil
}

// Count returns the total count of WODs
func (s *WODService) Count(isStandard *bool) (int64, error) {
	// Count all WODs or filtered by standard status
	var filters map[string]interface{}
	if isStandard != nil {
		filters = map[string]interface{}{
			"is_standard": *isStandard,
		}
	}

	wods, err := s.wodRepo.List(filters, 10000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to count wods: %w", err)
	}

	return int64(len(wods)), nil
}

// validateWOD validates WOD required fields and business rules
func (s *WODService) validateWOD(wod *domain.WOD) error {
	// Validate name
	if strings.TrimSpace(wod.Name) == "" {
		return ErrWODNameRequired
	}

	// Validate source
	if strings.TrimSpace(wod.Source) == "" {
		return ErrWODSourceRequired
	}

	// Validate type
	if strings.TrimSpace(wod.Type) == "" {
		return ErrWODTypeRequired
	}

	// Validate allowed values for enums
	validSources := map[string]bool{
		"CrossFit":      true,
		"Other Coach":   true,
		"Self-recorded": true,
	}
	if !validSources[wod.Source] {
		return fmt.Errorf("invalid source: must be one of [CrossFit, Other Coach, Self-recorded]")
	}

	validTypes := map[string]bool{
		"Benchmark":    true,
		"Hero":         true,
		"Girl":         true,
		"Notables":     true,
		"Games":        true,
		"Endurance":    true,
		"Self-created": true,
	}
	if !validTypes[wod.Type] {
		return fmt.Errorf("invalid type: must be one of [Benchmark, Hero, Girl, Notables, Games, Endurance, Self-created]")
	}

	// Validate regime (optional but if provided must be valid)
	if wod.Regime != "" {
		validRegimes := map[string]bool{
			"EMOM":          true,
			"AMRAP":         true,
			"Fastest Time":  true,
			"Slowest Round": true,
			"Get Stronger":  true,
			"Skills":        true,
		}
		if !validRegimes[wod.Regime] {
			return fmt.Errorf("invalid regime: must be one of [EMOM, AMRAP, Fastest Time, Slowest Round, Get Stronger, Skills]")
		}
	}

	// Validate score type (optional but if provided must be valid)
	if wod.ScoreType != "" {
		validScoreTypes := map[string]bool{
			"Time (HH:MM:SS)": true,
			"Rounds+Reps":     true,
			"Max Weight":      true,
		}
		if !validScoreTypes[wod.ScoreType] {
			return fmt.Errorf("invalid score type: must be one of [Time (HH:MM:SS), Rounds+Reps, Max Weight]")
		}
	}

	return nil
}

// paginateWODs applies limit and offset to a slice of WODs
func (s *WODService) paginateWODs(wods []*domain.WOD, limit, offset int) []*domain.WOD {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 100
	}

	// Calculate start and end indices
	start := offset
	if start < 0 {
		start = 0
	}
	if start >= len(wods) {
		return []*domain.WOD{}
	}

	end := start + limit
	if end > len(wods) {
		end = len(wods)
	}

	return wods[start:end]
}
