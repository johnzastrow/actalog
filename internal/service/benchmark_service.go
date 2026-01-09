package service

import (
	"bufio"
	"encoding/json"
	"fmt"
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

	result := &domain.BenchmarkSuiteResult{
		Timestamp: startTime,
		Version:   version.Version(),
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

	// Run database benchmarks
	s.runDatabaseBenchmarks(userID, result)

	// Run serialization benchmarks
	s.runSerializationBenchmarks(result)

	// Run business logic benchmarks
	s.runBusinessLogicBenchmarks(result)

	// Run concurrent benchmarks if requested
	if options.IncludeConcurrent {
		result.Concurrent = &domain.ConcurrentBenchmarkResult{}
		s.runConcurrentBenchmarks(userID, result)
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
func (s *BenchmarkService) runDatabaseBenchmarks(userID int64, result *domain.BenchmarkSuiteResult) {
	// 1. Single insert
	result.Database.Insert = s.benchmarkInsert(userID)

	// 2. Bulk insert (100 records)
	result.Database.BulkInsert = s.benchmarkBulkInsert(userID, 100)

	// 3. Select by ID
	result.Database.SelectByID = s.benchmarkSelectByID()

	// 4. Select by key
	result.Database.SelectByKey = s.benchmarkSelectByKey()

	// 5. Select list (pagination)
	result.Database.SelectList = s.benchmarkSelectList()

	// 6. Select filtered
	result.Database.SelectFiltered = s.benchmarkSelectFiltered(userID)

	// 7. Update
	result.Database.Update = s.benchmarkUpdate()

	// 8. Delete
	result.Database.Delete = s.benchmarkDelete()
}

// benchmarkInsert tests single record insertion
func (s *BenchmarkService) benchmarkInsert(userID int64) *domain.OperationResult {
	start := time.Now()
	data := &domain.BenchmarkData{
		TestKey:   fmt.Sprintf("bench_insert_%d", time.Now().UnixNano()),
		TestValue: "single insert test",
		NumValue:  3.14159,
		IntValue:  42,
		BoolValue: true,
		CreatedBy: userID,
	}

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

// benchmarkBulkInsert tests batch record insertion
func (s *BenchmarkService) benchmarkBulkInsert(userID int64, count int) *domain.OperationResult {
	start := time.Now()

	data := make([]*domain.BenchmarkData, count)
	for i := 0; i < count; i++ {
		data[i] = &domain.BenchmarkData{
			TestKey:   fmt.Sprintf("bench_bulk_%d_%d", time.Now().UnixNano(), i),
			TestValue: fmt.Sprintf("bulk insert test record %d", i),
			NumValue:  float64(i) * 1.5,
			IntValue:  i,
			BoolValue: i%2 == 0,
			CreatedBy: userID,
		}
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

// benchmarkSelectList tests pagination
func (s *BenchmarkService) benchmarkSelectList() *domain.OperationResult {
	start := time.Now()
	items, err := s.repo.List(50, 0)
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

// benchmarkSelectFiltered tests filtered queries
func (s *BenchmarkService) benchmarkSelectFiltered(userID int64) *domain.OperationResult {
	start := time.Now()
	filters := domain.BenchmarkFilters{
		CreatedBy: &userID,
	}
	items, err := s.repo.ListFiltered(filters, 50, 0)
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

// runSerializationBenchmarks runs JSON serialization benchmarks
func (s *BenchmarkService) runSerializationBenchmarks(result *domain.BenchmarkSuiteResult) {
	// Create sample data
	smallData := &domain.BenchmarkData{
		ID:        1,
		TestKey:   "serialization_test",
		TestValue: "test value for serialization",
		NumValue:  123.456,
		IntValue:  789,
		BoolValue: true,
		CreatedBy: 1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	largeData := make([]*domain.BenchmarkData, 100)
	for i := 0; i < 100; i++ {
		largeData[i] = &domain.BenchmarkData{
			ID:        int64(i),
			TestKey:   fmt.Sprintf("key_%d", i),
			TestValue: fmt.Sprintf("value_%d with some extra text to make it realistic", i),
			NumValue:  float64(i) * 1.23456,
			IntValue:  i * 10,
			BoolValue: i%2 == 0,
			CreatedBy: 1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// Marshal small
	result.Serialization.MarshalSmall = s.benchmarkMarshal("marshal_small", smallData)

	// Marshal large
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

// runBusinessLogicBenchmarks runs business logic benchmarks
func (s *BenchmarkService) runBusinessLogicBenchmarks(result *domain.BenchmarkSuiteResult) {
	// 1RM calculations
	result.BusinessLogic.OneRMCalcs = s.benchmark1RMCalculations(1000)

	// Intensity calculations
	result.BusinessLogic.IntensityCalcs = s.benchmarkIntensityCalculations(1000)

	// Input validation
	result.BusinessLogic.Validation = s.benchmarkValidation(100)

	// String operations
	result.BusinessLogic.StringOps = s.benchmarkStringOperations(1000)

	// Date operations
	result.BusinessLogic.DateOps = s.benchmarkDateOperations(1000)
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

// runConcurrentBenchmarks runs concurrent operation benchmarks
func (s *BenchmarkService) runConcurrentBenchmarks(userID int64, result *domain.BenchmarkSuiteResult) {
	// Parallel reads
	result.Concurrent.ParallelReads = s.benchmarkParallelReads(10)

	// Parallel writes
	result.Concurrent.ParallelWrites = s.benchmarkParallelWrites(userID, 5)

	// Mixed operations
	result.Concurrent.MixedOps = s.benchmarkMixedOperations(userID, 10)
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
