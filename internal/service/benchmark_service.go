package service

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
	"github.com/johnzastrow/actalog/pkg/prmath"
	"github.com/johnzastrow/actalog/pkg/version"
)

// Categories for benchmark data
var benchmarkCategories = []string{
	"strength", "cardio", "flexibility", "endurance", "power",
	"balance", "agility", "speed", "coordination", "reaction",
}

// JSON blob templates for complex data generation
type complexJSONData struct {
	Metadata    map[string]interface{} `json:"metadata"`
	Measurements []measurement         `json:"measurements"`
	Tags        []string               `json:"tags"`
	Attributes  map[string]string      `json:"attributes"`
}

type measurement struct {
	Name      string    `json:"name"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
}

// BenchmarkService handles benchmark operations
type BenchmarkService struct {
	repo *repository.BenchmarkRepository
}

// NewBenchmarkService creates a new benchmark service
func NewBenchmarkService(repo *repository.BenchmarkRepository) *BenchmarkService {
	return &BenchmarkService{
		repo: repo,
	}
}

// RunBenchmark executes the complete benchmark suite
func (s *BenchmarkService) RunBenchmark(userID int64, options domain.BenchmarkOptions) (*domain.BenchmarkSuiteResult, error) {
	startTime := time.Now()

	// Validate and set record count
	recordCount := options.RecordCount
	if recordCount <= 0 {
		recordCount = domain.DefaultRecordCount
	}
	if recordCount > domain.MaxRecordCount {
		recordCount = domain.MaxRecordCount
	}

	result := &domain.BenchmarkSuiteResult{
		Timestamp:   startTime,
		Version:     version.Version(),
		RecordCount: recordCount,
		SystemInfo: &domain.SystemInfo{
			GoVersion:       runtime.Version(),
			GoOS:            runtime.GOOS,
			GoArch:          runtime.GOARCH,
			OSVersion:       getOSVersion(),
			NumCPU:          runtime.NumCPU(),
			DatabaseDriver:  s.repo.Driver(),
			DatabaseVersion: s.repo.DatabaseVersion(),
		},
		Database:      &domain.DatabaseBenchmarkResult{},
		Serialization: &domain.SerializationBenchmarkResult{},
		BusinessLogic: &domain.BusinessLogicBenchmarkResult{},
	}

	// Run database benchmarks with configurable record count
	s.runDatabaseBenchmarks(userID, recordCount, result)

	// Run serialization benchmarks with complex data
	s.runSerializationBenchmarks(recordCount, result)

	// Run business logic benchmarks scaled by record count
	s.runBusinessLogicBenchmarks(recordCount, result)

	// Run concurrent benchmarks if requested
	if options.IncludeConcurrent {
		result.Concurrent = &domain.ConcurrentBenchmarkResult{}
		s.runConcurrentBenchmarks(userID, recordCount, result)
	}

	// Cleanup if requested (default behavior)
	if options.Cleanup {
		cleanupResult := s.cleanupBenchmarkData(userID)
		result.Database.Cleanup = cleanupResult
	}

	// Calculate totals
	result.TotalDurationMs = float64(time.Since(startTime).Microseconds()) / 1000.0
	s.calculateTotals(result)

	// Determine overall status
	if result.FailedOperations == 0 {
		result.Overall = "pass"
	} else if result.FailedOperations < result.TotalOperations/2 {
		result.Overall = "degraded"
	} else {
		result.Overall = "fail"
	}

	return result, nil
}

// runDatabaseBenchmarks runs all database-related benchmarks
func (s *BenchmarkService) runDatabaseBenchmarks(userID int64, recordCount int, result *domain.BenchmarkSuiteResult) {
	// 1. Single insert with complex data
	result.Database.Insert = s.benchmarkInsert(userID)

	// 2. Bulk insert (configurable record count)
	result.Database.BulkInsert = s.benchmarkBulkInsert(userID, recordCount)

	// 3. Select by ID
	result.Database.SelectByID = s.benchmarkSelectByID()

	// 4. Select by key
	result.Database.SelectByKey = s.benchmarkSelectByKey()

	// 5. Select list (pagination) - scale with record count
	listSize := min(recordCount, 1000)
	result.Database.SelectList = s.benchmarkSelectList(listSize)

	// 6. Select filtered
	result.Database.SelectFiltered = s.benchmarkSelectFiltered(userID, listSize)

	// 7. Update
	result.Database.Update = s.benchmarkUpdate()

	// 8. Delete
	result.Database.Delete = s.benchmarkDelete()
}

// benchmarkInsert tests single record insertion with complex data
func (s *BenchmarkService) benchmarkInsert(userID int64) *domain.OperationResult {
	start := time.Now()
	data := generateComplexBenchmarkData(userID, 0)

	err := s.repo.Create(data)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "insert",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "insert",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: 1,
	}
}

// benchmarkBulkInsert tests batch record insertion with complex data
func (s *BenchmarkService) benchmarkBulkInsert(userID int64, count int) *domain.OperationResult {
	start := time.Now()

	data := make([]*domain.BenchmarkData, count)
	for i := 0; i < count; i++ {
		data[i] = generateComplexBenchmarkData(userID, i)
	}

	err := s.repo.CreateBatch(data)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "bulk_insert",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "bulk_insert",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: count,
	}
}

// benchmarkSelectByID tests primary key lookup
func (s *BenchmarkService) benchmarkSelectByID() *domain.OperationResult {
	// First, get a valid ID
	items, err := s.repo.List(1, 0)
	if err != nil || len(items) == 0 {
		return &domain.OperationResult{
			Operation:  "select_by_id",
			Success:    false,
			DurationMs: 0,
			Error:      "no records available for lookup",
		}
	}

	start := time.Now()
	_, err = s.repo.GetByID(items[0].ID)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "select_by_id",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "select_by_id",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: 1,
	}
}

// benchmarkSelectByKey tests index lookup
func (s *BenchmarkService) benchmarkSelectByKey() *domain.OperationResult {
	// First, get a valid key
	items, err := s.repo.List(1, 0)
	if err != nil || len(items) == 0 {
		return &domain.OperationResult{
			Operation:  "select_by_key",
			Success:    false,
			DurationMs: 0,
			Error:      "no records available for lookup",
		}
	}

	start := time.Now()
	_, err = s.repo.GetByKey(items[0].TestKey)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "select_by_key",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "select_by_key",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: 1,
	}
}

// benchmarkSelectList tests pagination with configurable size
func (s *BenchmarkService) benchmarkSelectList(listSize int) *domain.OperationResult {
	start := time.Now()
	items, err := s.repo.List(listSize, 0)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "select_list",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "select_list",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: len(items),
	}
}

// benchmarkSelectFiltered tests filtered queries with multiple conditions
func (s *BenchmarkService) benchmarkSelectFiltered(userID int64, listSize int) *domain.OperationResult {
	start := time.Now()

	// Use multiple filters to stress test query building
	category := benchmarkCategories[0]
	minPriority := 50
	filters := domain.BenchmarkFilters{
		CreatedBy:   &userID,
		Category:    &category,
		MinPriority: &minPriority,
	}
	items, err := s.repo.ListFiltered(filters, listSize, 0)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "select_filtered",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "select_filtered",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: len(items),
	}
}

// benchmarkUpdate tests record updates
func (s *BenchmarkService) benchmarkUpdate() *domain.OperationResult {
	// Get a record to update
	items, err := s.repo.List(1, 0)
	if err != nil || len(items) == 0 {
		return &domain.OperationResult{
			Operation:  "update",
			Success:    false,
			DurationMs: 0,
			Error:      "no records available for update",
		}
	}

	item := items[0]
	item.TestValue = fmt.Sprintf("updated at %s", time.Now().Format(time.RFC3339))
	item.NumValue = item.NumValue + 1.0

	start := time.Now()
	err = s.repo.Update(item)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "update",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "update",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: 1,
	}
}

// benchmarkDelete tests record deletion
func (s *BenchmarkService) benchmarkDelete() *domain.OperationResult {
	// Get a record to delete
	items, err := s.repo.List(1, 0)
	if err != nil || len(items) == 0 {
		return &domain.OperationResult{
			Operation:  "delete",
			Success:    false,
			DurationMs: 0,
			Error:      "no records available for deletion",
		}
	}

	start := time.Now()
	err = s.repo.Delete(items[0].ID)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "delete",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "delete",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: 1,
	}
}

// cleanupBenchmarkData removes benchmark data for the user
func (s *BenchmarkService) cleanupBenchmarkData(userID int64) *domain.OperationResult {
	start := time.Now()
	err := s.repo.DeleteByUserID(userID)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  "cleanup",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:  "cleanup",
		Success:    true,
		DurationMs: duration,
	}
}

// runSerializationBenchmarks runs JSON serialization benchmarks with complex data
func (s *BenchmarkService) runSerializationBenchmarks(recordCount int, result *domain.BenchmarkSuiteResult) {
	// Create complex sample data
	smallData := generateComplexBenchmarkData(1, 0)
	smallData.ID = 1

	// Scale the large data set based on record count (use 10% of record count, min 100, max 10000)
	largeCount := max(100, min(recordCount/10, 10000))
	largeData := make([]*domain.BenchmarkData, largeCount)
	for i := 0; i < largeCount; i++ {
		largeData[i] = generateComplexBenchmarkData(1, i)
		largeData[i].ID = int64(i)
	}

	// Marshal small (single complex record)
	result.Serialization.MarshalSmall = s.benchmarkMarshal("marshal_small", smallData)

	// Marshal large (many complex records)
	result.Serialization.MarshalLarge = s.benchmarkMarshal("marshal_large", largeData)

	// Unmarshal small
	smallJSON, _ := json.Marshal(smallData)
	result.Serialization.UnmarshalSmall = s.benchmarkUnmarshal("unmarshal_small", smallJSON, &domain.BenchmarkData{})

	// Unmarshal large
	largeJSON, _ := json.Marshal(largeData)
	result.Serialization.UnmarshalLarge = s.benchmarkUnmarshal("unmarshal_large", largeJSON, &[]*domain.BenchmarkData{})
}

func (s *BenchmarkService) benchmarkMarshal(operation string, data interface{}) *domain.OperationResult {
	start := time.Now()
	_, err := json.Marshal(data)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  operation,
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:  operation,
		Success:    true,
		DurationMs: duration,
	}
}

func (s *BenchmarkService) benchmarkUnmarshal(operation string, data []byte, target interface{}) *domain.OperationResult {
	start := time.Now()
	err := json.Unmarshal(data, target)
	duration := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return &domain.OperationResult{
			Operation:  operation,
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:  operation,
		Success:    true,
		DurationMs: duration,
	}
}

// runBusinessLogicBenchmarks runs business logic benchmarks scaled by record count
func (s *BenchmarkService) runBusinessLogicBenchmarks(recordCount int, result *domain.BenchmarkSuiteResult) {
	// Scale iterations based on record count (use record count directly for compute benchmarks)
	iterations := recordCount

	// 1RM calculations
	result.BusinessLogic.OneRMCalcs = s.benchmark1RMCalculations(iterations)

	// Intensity calculations
	result.BusinessLogic.IntensityCalcs = s.benchmarkIntensityCalculations(iterations)

	// Input validation (1/10 of iterations as it's more complex)
	result.BusinessLogic.Validation = s.benchmarkValidation(iterations / 10)

	// String operations
	result.BusinessLogic.StringOps = s.benchmarkStringOperations(iterations)

	// Date operations
	result.BusinessLogic.DateOps = s.benchmarkDateOperations(iterations)
}

func (s *BenchmarkService) benchmark1RMCalculations(iterations int) *domain.OperationResult {
	start := time.Now()

	for i := 0; i < iterations; i++ {
		weight := float64(100 + i%100)
		reps := 1 + i%15
		prmath.Calculate1RM(weight, reps)
	}

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	return &domain.OperationResult{
		Operation:       "one_rm_calculations",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: iterations,
	}
}

func (s *BenchmarkService) benchmarkIntensityCalculations(iterations int) *domain.OperationResult {
	start := time.Now()

	for i := 0; i < iterations; i++ {
		weight := float64(80 + i%50)
		oneRM := float64(120 + i%30)
		prmath.CalculateIntensity(weight, oneRM)
	}

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	return &domain.OperationResult{
		Operation:       "intensity_calculations",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: iterations,
	}
}

func (s *BenchmarkService) benchmarkValidation(iterations int) *domain.OperationResult {
	start := time.Now()

	testInputs := []string{
		"valid_input",
		"",
		"a",
		"this is a very long input string that needs to be validated",
		"special!@#$%chars",
		"  trimmed  ",
	}

	for i := 0; i < iterations; i++ {
		input := testInputs[i%len(testInputs)]
		// Simulate validation operations
		_ = len(input) > 0
		_ = len(input) < 100
		_ = strings.TrimSpace(input)
		_ = strings.ToLower(input)
	}

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	return &domain.OperationResult{
		Operation:       "validation",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: iterations,
	}
}

func (s *BenchmarkService) benchmarkStringOperations(iterations int) *domain.OperationResult {
	start := time.Now()

	for i := 0; i < iterations; i++ {
		s1 := fmt.Sprintf("string_%d", i)
		s2 := fmt.Sprintf("another_%d_string", i)

		// Various string operations
		_ = strings.Contains(s1, "string")
		_ = strings.HasPrefix(s1, "str")
		_ = strings.HasSuffix(s1, fmt.Sprintf("%d", i))
		_ = strings.ToUpper(s1)
		_ = strings.Replace(s2, "another", "replaced", 1)
		_ = strings.Split(s2, "_")
		_ = strings.Join([]string{s1, s2}, " ")
	}

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	return &domain.OperationResult{
		Operation:       "string_operations",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: iterations,
	}
}

func (s *BenchmarkService) benchmarkDateOperations(iterations int) *domain.OperationResult {
	start := time.Now()

	now := time.Now()
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"Jan 2, 2006",
	}

	for i := 0; i < iterations; i++ {
		// Date formatting
		format := formats[i%len(formats)]
		formatted := now.Format(format)

		// Date parsing
		_, _ = time.Parse(format, formatted)

		// Date arithmetic
		_ = now.Add(time.Duration(i) * time.Hour)
		_ = now.AddDate(0, 0, i%30)

		// Timezone operations
		_ = now.UTC()
		_ = now.Local()
	}

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	return &domain.OperationResult{
		Operation:       "date_operations",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: iterations,
	}
}

// runConcurrentBenchmarks runs concurrent operation benchmarks scaled by record count
func (s *BenchmarkService) runConcurrentBenchmarks(userID int64, recordCount int, result *domain.BenchmarkSuiteResult) {
	// Scale goroutines based on record count (min 10, max 100)
	goroutines := max(10, min(recordCount/100, 100))
	writeGoroutines := max(5, min(recordCount/200, 50))

	// Parallel reads
	result.Concurrent.ParallelReads = s.benchmarkParallelReads(goroutines)

	// Parallel writes
	result.Concurrent.ParallelWrites = s.benchmarkParallelWrites(userID, writeGoroutines)

	// Mixed operations
	result.Concurrent.MixedOps = s.benchmarkMixedOperations(userID, goroutines)
}

func (s *BenchmarkService) benchmarkParallelReads(goroutines int) *domain.OperationResult {
	start := time.Now()

	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.repo.List(10, 0)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	// Check for errors
	for err := range errors {
		return &domain.OperationResult{
			Operation:  "parallel_reads",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "parallel_reads",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: goroutines,
	}
}

func (s *BenchmarkService) benchmarkParallelWrites(userID int64, goroutines int) *domain.OperationResult {
	start := time.Now()

	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			data := &domain.BenchmarkData{
				TestKey:   fmt.Sprintf("parallel_write_%d_%d", time.Now().UnixNano(), idx),
				TestValue: fmt.Sprintf("parallel write test %d", idx),
				NumValue:  float64(idx),
				IntValue:  idx,
				BoolValue: idx%2 == 0,
				CreatedBy: userID,
			}
			if err := s.repo.Create(data); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	// Check for errors
	for err := range errors {
		return &domain.OperationResult{
			Operation:  "parallel_writes",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "parallel_writes",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: goroutines,
	}
}

func (s *BenchmarkService) benchmarkMixedOperations(userID int64, goroutines int) *domain.OperationResult {
	start := time.Now()

	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Mix of read and write operations
			if idx%2 == 0 {
				// Read operation
				_, err := s.repo.List(5, 0)
				if err != nil {
					errors <- err
				}
			} else {
				// Write operation
				data := &domain.BenchmarkData{
					TestKey:   fmt.Sprintf("mixed_op_%d_%d", time.Now().UnixNano(), idx),
					TestValue: fmt.Sprintf("mixed operation test %d", idx),
					NumValue:  float64(idx),
					IntValue:  idx,
					BoolValue: true,
					CreatedBy: userID,
				}
				if err := s.repo.Create(data); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	// Check for errors
	for err := range errors {
		return &domain.OperationResult{
			Operation:  "mixed_operations",
			Success:    false,
			DurationMs: duration,
			Error:      err.Error(),
		}
	}

	return &domain.OperationResult{
		Operation:       "mixed_operations",
		Success:         true,
		DurationMs:      duration,
		RecordsAffected: goroutines,
	}
}

// calculateTotals calculates the total operations and success/failure counts
func (s *BenchmarkService) calculateTotals(result *domain.BenchmarkSuiteResult) {
	ops := []*domain.OperationResult{}

	// Database operations
	if result.Database != nil {
		ops = append(ops,
			result.Database.Insert,
			result.Database.BulkInsert,
			result.Database.SelectByID,
			result.Database.SelectByKey,
			result.Database.SelectList,
			result.Database.SelectFiltered,
			result.Database.Update,
			result.Database.Delete,
			result.Database.Cleanup,
		)
	}

	// Serialization operations
	if result.Serialization != nil {
		ops = append(ops,
			result.Serialization.MarshalSmall,
			result.Serialization.MarshalLarge,
			result.Serialization.UnmarshalSmall,
			result.Serialization.UnmarshalLarge,
		)
	}

	// Business logic operations
	if result.BusinessLogic != nil {
		ops = append(ops,
			result.BusinessLogic.OneRMCalcs,
			result.BusinessLogic.IntensityCalcs,
			result.BusinessLogic.Validation,
			result.BusinessLogic.StringOps,
			result.BusinessLogic.DateOps,
		)
	}

	// Concurrent operations
	if result.Concurrent != nil {
		ops = append(ops,
			result.Concurrent.ParallelReads,
			result.Concurrent.ParallelWrites,
			result.Concurrent.MixedOps,
		)
	}

	for _, op := range ops {
		if op != nil {
			result.TotalOperations++
			if op.Success {
				result.SuccessfulOperations++
			} else {
				result.FailedOperations++
			}
		}
	}
}

// GetBenchmarkStatus returns current benchmark data status
func (s *BenchmarkService) GetBenchmarkStatus() (map[string]interface{}, error) {
	count, err := s.repo.Count()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"version":         version.Version(),
		"database_driver": s.repo.Driver(),
		"record_count":    count,
		"status":          "ready",
	}, nil
}

// CleanupAllBenchmarkData removes all benchmark data (admin only)
func (s *BenchmarkService) CleanupAllBenchmarkData() error {
	return s.repo.DeleteAll()
}

// getOSVersion returns a detailed OS version string
func getOSVersion() string {
	switch runtime.GOOS {
	case "linux":
		return getLinuxVersion()
	case "darwin":
		return getDarwinVersion()
	case "windows":
		return getWindowsVersion()
	default:
		return runtime.GOOS
	}
}

// getLinuxVersion reads /etc/os-release for Linux distribution info
func getLinuxVersion() string {
	// Try /etc/os-release first (most modern distros)
	file, err := os.Open("/etc/os-release")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		var prettyName string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				prettyName = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
		if prettyName != "" {
			// Try to get kernel version too
			if out, err := exec.Command("uname", "-r").Output(); err == nil {
				return prettyName + " (kernel " + strings.TrimSpace(string(out)) + ")"
			}
			return prettyName
		}
	}

	// Fallback to lsb_release
	if out, err := exec.Command("lsb_release", "-d", "-s").Output(); err == nil {
		desc := strings.TrimSpace(string(out))
		if out, err := exec.Command("uname", "-r").Output(); err == nil {
			return desc + " (kernel " + strings.TrimSpace(string(out)) + ")"
		}
		return desc
	}

	// Final fallback to uname
	if out, err := exec.Command("uname", "-a").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}

	return "Linux"
}

// getDarwinVersion returns macOS version info
func getDarwinVersion() string {
	if out, err := exec.Command("sw_vers", "-productName").Output(); err == nil {
		name := strings.TrimSpace(string(out))
		if ver, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			return name + " " + strings.TrimSpace(string(ver))
		}
		return name
	}
	return "macOS"
}

// getWindowsVersion returns Windows version info
func getWindowsVersion() string {
	if out, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
		return strings.TrimSpace(string(out))
	}
	return "Windows"
}

// generateComplexBenchmarkData creates a complex benchmark data record with large, random data
func generateComplexBenchmarkData(userID int64, index int) *domain.BenchmarkData {
	now := time.Now()

	return &domain.BenchmarkData{
		TestKey:    fmt.Sprintf("bench_%d_%d_%s", now.UnixNano(), index, generateRandomString(8)),
		TestValue:  fmt.Sprintf("benchmark test value %d with random: %s", index, generateRandomString(32)),
		NumValue:   generateRandomFloat(),
		IntValue:   generateRandomInt(1000000),
		BoolValue:  index%2 == 0,
		LargeText:  generateLargeText(index),
		JsonBlob:   generateComplexJSON(index),
		ExtraFloat: generateRandomFloat() * 1000,
		ExtraInt:   generateRandomInt(10000000),
		Category:   benchmarkCategories[index%len(benchmarkCategories)],
		Priority:   generateRandomInt(100),
		CreatedBy:  userID,
	}
}

// generateLargeText creates a large text field (5-10KB) for stress testing
func generateLargeText(seed int) string {
	// Generate 5-10KB of text
	targetSize := 5000 + (seed % 5000)
	var sb strings.Builder
	sb.Grow(targetSize)

	paragraphs := []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. ",
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. ",
		"The quick brown fox jumps over the lazy dog. Pack my box with five dozen liquor jugs. How vexingly quick daft zebras jump! Sphinx of black quartz, judge my vow. ",
		"Benchmark data record %d: This is a test string containing various characters and numbers like 12345 and special chars @#$%. Testing database performance with complex text data. ",
		"Performance metrics collection in progress. Measuring insert, update, select, and delete operations. Analyzing query execution times and database throughput. ",
	}

	for sb.Len() < targetSize {
		para := paragraphs[(seed+sb.Len())%len(paragraphs)]
		if strings.Contains(para, "%d") {
			para = fmt.Sprintf(para, seed)
		}
		sb.WriteString(para)
	}

	return sb.String()[:targetSize]
}

// generateComplexJSON creates a complex JSON blob for stress testing
func generateComplexJSON(seed int) string {
	now := time.Now()

	data := complexJSONData{
		Metadata: map[string]interface{}{
			"version":    "1.0",
			"created_at": now.Format(time.RFC3339),
			"seed":       seed,
			"random_id":  generateRandomString(16),
			"nested": map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"value": generateRandomFloat(),
						"text":  generateRandomString(64),
					},
				},
			},
		},
		Measurements: make([]measurement, 10),
		Tags:         make([]string, 20),
		Attributes:   make(map[string]string),
	}

	// Generate measurements
	for i := 0; i < 10; i++ {
		data.Measurements[i] = measurement{
			Name:      fmt.Sprintf("measurement_%d", i),
			Value:     generateRandomFloat() * 100,
			Unit:      []string{"kg", "lbs", "reps", "seconds", "meters"}[i%5],
			Timestamp: now.Add(time.Duration(-i) * time.Hour),
		}
	}

	// Generate tags
	for i := 0; i < 20; i++ {
		data.Tags[i] = fmt.Sprintf("tag_%d_%s", i, generateRandomString(8))
	}

	// Generate attributes
	for i := 0; i < 30; i++ {
		data.Attributes[fmt.Sprintf("attr_%d", i)] = generateRandomString(32)
	}

	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}

// generateRandomString creates a random string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to math/rand if crypto/rand fails
			result[i] = charset[mathrand.Intn(len(charset))]
		} else {
			result[i] = charset[n.Int64()]
		}
	}
	return string(result)
}

// generateRandomFloat creates a random float64 value
func generateRandomFloat() float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return mathrand.Float64() * 1000
	}
	return float64(n.Int64()) / 1000.0
}

// generateRandomInt creates a random integer value up to max
func generateRandomInt(maxVal int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxVal)))
	if err != nil {
		return mathrand.Intn(maxVal)
	}
	return int(n.Int64())
}
