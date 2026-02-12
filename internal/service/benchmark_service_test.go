package service

import (
	"encoding/json"
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
	"github.com/johnzastrow/actalog/internal/repository"
)

func TestNewBenchmarkService(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	if svc == nil {
		t.Error("NewBenchmarkService() should return a non-nil service")
	}
}

func TestBenchmarkService_RunBenchmark(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user for foreign key
	userRepo := repository.NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "benchmark-test@example.com",
		PasswordHash: "hash",
		Role:         "athlete",
	}
	err = userRepo.Create(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	tests := []struct {
		name        string
		userID      int64
		options     domain.BenchmarkOptions
		wantErr     bool
		checkResult func(*domain.BenchmarkSuiteResult) bool
	}{
		{
			name:   "basic benchmark with defaults",
			userID: testUser.ID,
			options: domain.BenchmarkOptions{
				Cleanup: true,
			},
			wantErr: false,
			checkResult: func(r *domain.BenchmarkSuiteResult) bool {
				return r != nil && r.Overall != "" && r.TotalOperations > 0
			},
		},
		{
			name:   "benchmark with custom record count",
			userID: testUser.ID,
			options: domain.BenchmarkOptions{
				RecordCount: 10,
				Cleanup:     true,
			},
			wantErr: false,
			checkResult: func(r *domain.BenchmarkSuiteResult) bool {
				return r != nil && r.RecordCount == 10
			},
		},
		{
			name:   "benchmark with concurrent operations",
			userID: testUser.ID,
			options: domain.BenchmarkOptions{
				RecordCount:       5,
				IncludeConcurrent: true,
				Cleanup:           true,
			},
			wantErr: false,
			checkResult: func(r *domain.BenchmarkSuiteResult) bool {
				return r != nil && r.Concurrent != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			repo.DeleteByUserID(tt.userID)

			result, err := svc.RunBenchmark(tt.userID, tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunBenchmark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("RunBenchmark() returned nil result")
					return
				}
				if !tt.checkResult(result) {
					t.Errorf("RunBenchmark() result check failed")
				}

				// Verify basic result structure
				if result.Version == "" {
					t.Error("RunBenchmark() result missing Version")
				}
				if result.SystemInfo == nil {
					t.Error("RunBenchmark() result missing SystemInfo")
				}
				if result.Database == nil {
					t.Error("RunBenchmark() result missing Database")
				}
				if result.Serialization == nil {
					t.Error("RunBenchmark() result missing Serialization")
				}
				if result.BusinessLogic == nil {
					t.Error("RunBenchmark() result missing BusinessLogic")
				}
			}

			// Cleanup after test
			repo.DeleteByUserID(tt.userID)
		})
	}
}

func TestBenchmarkService_CleanupAllBenchmarkData(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "cleanup-test@example.com",
		PasswordHash: "hash",
		Role:         "admin",
	}
	userRepo.Create(testUser)

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	// Create some test data
	for i := 0; i < 5; i++ {
		repo.Create(&domain.BenchmarkData{
			TestKey:   "cleanup-test-key",
			CreatedBy: testUser.ID,
		})
	}

	// Verify data exists
	count, _ := repo.Count()
	if count == 0 {
		t.Fatal("Test data was not created")
	}

	// Cleanup
	err = svc.CleanupAllBenchmarkData()
	if err != nil {
		t.Fatalf("CleanupAllBenchmarkData() error = %v", err)
	}

	// Verify all data is gone
	count, _ = repo.Count()
	if count != 0 {
		t.Errorf("CleanupAllBenchmarkData() did not delete all data, count = %d", count)
	}
}

func TestBenchmarkService_GetBenchmarkStatus(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "status-test@example.com",
		PasswordHash: "hash",
		Role:         "athlete",
	}
	userRepo.Create(testUser)

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	// Create some test data
	for i := 0; i < 3; i++ {
		repo.Create(&domain.BenchmarkData{
			TestKey:   "status-test-key",
			CreatedBy: testUser.ID,
		})
	}

	status, err := svc.GetBenchmarkStatus()
	if err != nil {
		t.Fatalf("GetBenchmarkStatus() error = %v", err)
	}

	if status == nil {
		t.Error("GetBenchmarkStatus() returned nil")
	}

	// Check for expected fields
	if _, ok := status["database_driver"]; !ok {
		t.Error("GetBenchmarkStatus() missing database_driver")
	}
	if _, ok := status["version"]; !ok {
		t.Error("GetBenchmarkStatus() missing version")
	}
	if _, ok := status["record_count"]; !ok {
		t.Error("GetBenchmarkStatus() missing record_count")
	}
	if _, ok := status["status"]; !ok {
		t.Error("GetBenchmarkStatus() missing status")
	}

	// Verify record count
	if count, ok := status["record_count"].(int64); ok {
		if count != 3 {
			t.Errorf("GetBenchmarkStatus() record_count = %d, want 3", count)
		}
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{name: "short string", length: 5},
		{name: "medium string", length: 16},
		{name: "long string", length: 64},
		{name: "very long string", length: 256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateRandomString(tt.length)
			if len(result) != tt.length {
				t.Errorf("generateRandomString(%d) returned string of length %d", tt.length, len(result))
			}

			// Check it only contains valid characters
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			for _, c := range result {
				found := false
				for _, valid := range charset {
					if c == valid {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("generateRandomString() returned invalid character: %c", c)
				}
			}
		})
	}
}

func TestGenerateRandomFloat(t *testing.T) {
	for i := 0; i < 10; i++ {
		result := generateRandomFloat()
		if result < 0 || result >= 1000 {
			t.Errorf("generateRandomFloat() returned out of range value: %f", result)
		}
	}
}

func TestGenerateRandomInt(t *testing.T) {
	tests := []struct {
		name   string
		maxVal int
	}{
		{name: "small max", maxVal: 10},
		{name: "medium max", maxVal: 100},
		{name: "large max", maxVal: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 10; i++ {
				result := generateRandomInt(tt.maxVal)
				if result < 0 || result >= tt.maxVal {
					t.Errorf("generateRandomInt(%d) returned out of range value: %d", tt.maxVal, result)
				}
			}
		})
	}
}

func TestGenerateComplexJSON(t *testing.T) {
	for seed := 0; seed < 5; seed++ {
		result := generateComplexJSON(seed)

		if result == "" {
			t.Error("generateComplexJSON() returned empty string")
		}

		// Verify it's valid JSON
		var data complexJSONData
		err := json.Unmarshal([]byte(result), &data)
		if err != nil {
			t.Errorf("generateComplexJSON() returned invalid JSON: %v", err)
		}

		// Verify structure
		if data.Metadata == nil {
			t.Error("generateComplexJSON() missing metadata")
		}
		if len(data.Measurements) != 10 {
			t.Errorf("generateComplexJSON() measurements count = %d, want 10", len(data.Measurements))
		}
		if len(data.Tags) != 20 {
			t.Errorf("generateComplexJSON() tags count = %d, want 20", len(data.Tags))
		}
		if len(data.Attributes) != 30 {
			t.Errorf("generateComplexJSON() attributes count = %d, want 30", len(data.Attributes))
		}
	}
}

func TestBenchmarkService_RunBenchmark_DatabaseOperations(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "db-ops-test@example.com",
		PasswordHash: "hash",
		Role:         "athlete",
	}
	userRepo.Create(testUser)

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	// Run a benchmark with small record count for faster test
	result, err := svc.RunBenchmark(testUser.ID, domain.BenchmarkOptions{
		RecordCount: 5,
		Cleanup:     true,
	})

	if err != nil {
		t.Fatalf("RunBenchmark() error = %v", err)
	}

	// Verify database benchmark results
	if result.Database == nil {
		t.Fatal("Database benchmark results are nil")
	}

	// Check each operation has results
	if result.Database.Insert == nil {
		t.Error("Insert benchmark result is nil")
	} else if !result.Database.Insert.Success {
		t.Errorf("Insert benchmark failed: %s", result.Database.Insert.Error)
	}

	if result.Database.BulkInsert == nil {
		t.Error("BulkInsert benchmark result is nil")
	} else if !result.Database.BulkInsert.Success {
		t.Errorf("BulkInsert benchmark failed: %s", result.Database.BulkInsert.Error)
	}

	if result.Database.SelectByID == nil {
		t.Error("SelectByID benchmark result is nil")
	}

	if result.Database.SelectByKey == nil {
		t.Error("SelectByKey benchmark result is nil")
	}

	if result.Database.SelectList == nil {
		t.Error("SelectList benchmark result is nil")
	}

	if result.Database.Update == nil {
		t.Error("Update benchmark result is nil")
	}

	if result.Database.Delete == nil {
		t.Error("Delete benchmark result is nil")
	}
}

func TestBenchmarkService_RunBenchmark_SerializationOperations(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "serialization-test@example.com",
		PasswordHash: "hash",
		Role:         "athlete",
	}
	userRepo.Create(testUser)

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	result, err := svc.RunBenchmark(testUser.ID, domain.BenchmarkOptions{
		RecordCount: 5,
		Cleanup:     true,
	})

	if err != nil {
		t.Fatalf("RunBenchmark() error = %v", err)
	}

	// Verify serialization benchmark results
	if result.Serialization == nil {
		t.Fatal("Serialization benchmark results are nil")
	}

	if result.Serialization.MarshalSmall == nil {
		t.Error("MarshalSmall benchmark result is nil")
	} else if !result.Serialization.MarshalSmall.Success {
		t.Errorf("MarshalSmall benchmark failed: %s", result.Serialization.MarshalSmall.Error)
	}

	if result.Serialization.MarshalLarge == nil {
		t.Error("MarshalLarge benchmark result is nil")
	}

	if result.Serialization.UnmarshalSmall == nil {
		t.Error("UnmarshalSmall benchmark result is nil")
	}

	if result.Serialization.UnmarshalLarge == nil {
		t.Error("UnmarshalLarge benchmark result is nil")
	}
}

func TestBenchmarkService_RunBenchmark_BusinessLogicOperations(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "business-logic-test@example.com",
		PasswordHash: "hash",
		Role:         "athlete",
	}
	userRepo.Create(testUser)

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	result, err := svc.RunBenchmark(testUser.ID, domain.BenchmarkOptions{
		RecordCount: 5,
		Cleanup:     true,
	})

	if err != nil {
		t.Fatalf("RunBenchmark() error = %v", err)
	}

	// Verify business logic benchmark results
	if result.BusinessLogic == nil {
		t.Fatal("BusinessLogic benchmark results are nil")
	}

	if result.BusinessLogic.OneRMCalcs == nil {
		t.Error("OneRMCalcs benchmark result is nil")
	} else if !result.BusinessLogic.OneRMCalcs.Success {
		t.Errorf("OneRMCalcs benchmark failed: %s", result.BusinessLogic.OneRMCalcs.Error)
	}

	if result.BusinessLogic.IntensityCalcs == nil {
		t.Error("IntensityCalcs benchmark result is nil")
	}

	if result.BusinessLogic.Validation == nil {
		t.Error("Validation benchmark result is nil")
	}

	if result.BusinessLogic.StringOps == nil {
		t.Error("StringOps benchmark result is nil")
	}

	if result.BusinessLogic.DateOps == nil {
		t.Error("DateOps benchmark result is nil")
	}
}

func TestBenchmarkService_RunBenchmark_ConcurrentOperations(t *testing.T) {
	db, cleanup, err := repository.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	// Create test user
	userRepo := repository.NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "concurrent-test@example.com",
		PasswordHash: "hash",
		Role:         "athlete",
	}
	userRepo.Create(testUser)

	repo := repository.NewBenchmarkRepository(db, "sqlite3")
	svc := NewBenchmarkService(repo)

	result, err := svc.RunBenchmark(testUser.ID, domain.BenchmarkOptions{
		RecordCount:       5,
		IncludeConcurrent: true,
		Cleanup:           true,
	})

	if err != nil {
		t.Fatalf("RunBenchmark() error = %v", err)
	}

	// Verify concurrent benchmark results
	if result.Concurrent == nil {
		t.Fatal("Concurrent benchmark results are nil")
	}

	if result.Concurrent.ParallelReads == nil {
		t.Error("ParallelReads benchmark result is nil")
	}

	if result.Concurrent.ParallelWrites == nil {
		t.Error("ParallelWrites benchmark result is nil")
	}

	if result.Concurrent.MixedOps == nil {
		t.Error("MixedOps benchmark result is nil")
	}
}
