package domain

import "time"

// BenchmarkData - isolated table for safe read/write testing
// This table is separate from production data and used for benchmarking operations
type BenchmarkData struct {
	ID        int64     `json:"id" db:"id"`
	TestKey   string    `json:"test_key" db:"test_key"`
	TestValue string    `json:"test_value" db:"test_value"`
	NumValue  float64   `json:"num_value" db:"num_value"`
	IntValue  int       `json:"int_value" db:"int_value"`
	BoolValue bool      `json:"bool_value" db:"bool_value"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// BenchmarkRepository defines the interface for benchmark data access
type BenchmarkRepository interface {
	// Create creates a new benchmark data entry
	Create(data *BenchmarkData) error

	// CreateBatch creates multiple benchmark data entries
	CreateBatch(data []*BenchmarkData) error

	// GetByID retrieves a benchmark data entry by ID
	GetByID(id int64) (*BenchmarkData, error)

	// GetByKey retrieves a benchmark data entry by test key
	GetByKey(testKey string) (*BenchmarkData, error)

	// List retrieves benchmark data entries with pagination
	List(limit, offset int) ([]*BenchmarkData, error)

	// ListFiltered retrieves benchmark data with filters
	ListFiltered(filters BenchmarkFilters, limit, offset int) ([]*BenchmarkData, error)

	// Update updates a benchmark data entry
	Update(data *BenchmarkData) error

	// Delete deletes a benchmark data entry by ID
	Delete(id int64) error

	// DeleteByUserID deletes all benchmark data for a specific user
	DeleteByUserID(userID int64) error

	// DeleteAll deletes all benchmark data (admin only)
	DeleteAll() error

	// Count returns the total number of benchmark data entries
	Count() (int64, error)
}

// BenchmarkFilters represents filter options for querying benchmark data
type BenchmarkFilters struct {
	CreatedBy *int64  // Filter by creator user ID
	TestKey   *string // Filter by test key prefix
	BoolValue *bool   // Filter by boolean value
}

// OperationResult represents the result of a single benchmark operation
type OperationResult struct {
	Operation       string  `json:"operation"`
	Success         bool    `json:"success"`
	DurationMs      float64 `json:"duration_ms"`
	RecordsAffected int     `json:"records_affected,omitempty"`
	Error           string  `json:"error,omitempty"`
}

// DatabaseBenchmarkResult contains results of database operations
type DatabaseBenchmarkResult struct {
	Insert         *OperationResult `json:"insert"`
	BulkInsert     *OperationResult `json:"bulk_insert"`
	SelectByID     *OperationResult `json:"select_by_id"`
	SelectByKey    *OperationResult `json:"select_by_key"`
	SelectList     *OperationResult `json:"select_list"`
	SelectFiltered *OperationResult `json:"select_filtered"`
	Update         *OperationResult `json:"update"`
	Delete         *OperationResult `json:"delete"`
	Cleanup        *OperationResult `json:"cleanup"`
}

// SerializationBenchmarkResult contains results of JSON serialization operations
type SerializationBenchmarkResult struct {
	MarshalSmall   *OperationResult `json:"marshal_small"`
	MarshalLarge   *OperationResult `json:"marshal_large"`
	UnmarshalSmall *OperationResult `json:"unmarshal_small"`
	UnmarshalLarge *OperationResult `json:"unmarshal_large"`
}

// BusinessLogicBenchmarkResult contains results of business logic operations
type BusinessLogicBenchmarkResult struct {
	OneRMCalcs     *OperationResult `json:"one_rm_calculations"`
	IntensityCalcs *OperationResult `json:"intensity_calculations"`
	Validation     *OperationResult `json:"validation"`
	StringOps      *OperationResult `json:"string_operations"`
	DateOps        *OperationResult `json:"date_operations"`
}

// ConcurrentBenchmarkResult contains results of concurrent operations
type ConcurrentBenchmarkResult struct {
	ParallelReads  *OperationResult `json:"parallel_reads"`
	ParallelWrites *OperationResult `json:"parallel_writes"`
	MixedOps       *OperationResult `json:"mixed_operations"`
}

// BenchmarkSuiteResult contains the complete benchmark results
type BenchmarkSuiteResult struct {
	Timestamp       time.Time `json:"timestamp"`
	Version         string    `json:"version"`
	DatabaseDriver  string    `json:"database_driver"`
	TotalDurationMs float64   `json:"total_duration_ms"`
	Overall         string    `json:"overall"` // "pass", "fail", "degraded"

	Database      *DatabaseBenchmarkResult      `json:"database"`
	Serialization *SerializationBenchmarkResult `json:"serialization"`
	BusinessLogic *BusinessLogicBenchmarkResult `json:"business_logic"`
	Concurrent    *ConcurrentBenchmarkResult    `json:"concurrent,omitempty"` // optional

	TotalOperations      int `json:"total_operations"`
	SuccessfulOperations int `json:"successful_operations"`
	FailedOperations     int `json:"failed_operations"`
}

// BenchmarkOptions configures the benchmark run
type BenchmarkOptions struct {
	IncludeConcurrent bool // Include concurrent operation tests
	Cleanup           bool // Clean up benchmark data after run
}
