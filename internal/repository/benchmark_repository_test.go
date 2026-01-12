package repository

import (
	"testing"

	"github.com/johnzastrow/actalog/internal/domain"
)

func TestNewBenchmarkRepository(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")
	if repo == nil {
		t.Fatal("NewBenchmarkRepository() should return a non-nil repository")
	}

	if repo.Driver() != "sqlite3" {
		t.Errorf("Driver() = %q, want %q", repo.Driver(), "sqlite3")
	}
}

func TestBenchmarkRepository_Create(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user for foreign key
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "benchmark@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	err = userRepo.Create(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name    string
		data    *domain.BenchmarkData
		wantErr bool
	}{
		{
			name: "create with all fields",
			data: &domain.BenchmarkData{
				TestKey:    "test-key-1",
				TestValue:  "test-value-1",
				NumValue:   123.45,
				IntValue:   42,
				BoolValue:  true,
				LargeText:  "some large text content",
				JsonBlob:   `{"key": "value"}`,
				ExtraFloat: 99.99,
				ExtraInt:   100,
				Category:   "test-category",
				Priority:   5,
				CreatedBy:  testUser.ID,
			},
			wantErr: false,
		},
		{
			name: "create with minimal fields",
			data: &domain.BenchmarkData{
				TestKey:   "test-key-2",
				CreatedBy: testUser.ID,
			},
			wantErr: false,
		},
		{
			name: "create with empty values",
			data: &domain.BenchmarkData{
				TestKey:   "test-key-3",
				TestValue: "",
				NumValue:  0,
				IntValue:  0,
				BoolValue: false,
				CreatedBy: testUser.ID,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.data.ID == 0 {
					t.Error("Create() should set ID after creation")
				}
				if tt.data.CreatedAt.IsZero() {
					t.Error("Create() should set CreatedAt")
				}
				if tt.data.UpdatedAt.IsZero() {
					t.Error("Create() should set UpdatedAt")
				}
			}
		})
	}
}

func TestBenchmarkRepository_Create_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	data := &domain.BenchmarkData{
		TestKey:   "test-key",
		CreatedBy: 1,
	}

	err = repo.Create(data)
	if err == nil {
		t.Error("Create() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_CreateBatch(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "batch@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	err = userRepo.Create(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name      string
		data      []*domain.BenchmarkData
		wantErr   bool
		wantCount int64
	}{
		{
			name:      "empty batch",
			data:      []*domain.BenchmarkData{},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "single item batch",
			data: []*domain.BenchmarkData{
				{TestKey: "batch-1", TestValue: "value-1", CreatedBy: testUser.ID},
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "multiple items batch",
			data: []*domain.BenchmarkData{
				{TestKey: "batch-2", TestValue: "value-2", Priority: 1, CreatedBy: testUser.ID},
				{TestKey: "batch-3", TestValue: "value-3", Priority: 2, CreatedBy: testUser.ID},
				{TestKey: "batch-4", TestValue: "value-4", Priority: 3, CreatedBy: testUser.ID},
			},
			wantErr:   false,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			repo.DeleteAll()

			err := repo.CreateBatch(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				count, err := repo.Count()
				if err != nil {
					t.Errorf("Count() error = %v", err)
					return
				}
				if count != tt.wantCount {
					t.Errorf("Count() = %d, want %d", count, tt.wantCount)
				}
			}
		})
	}
}

func TestBenchmarkRepository_CreateBatch_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	data := []*domain.BenchmarkData{
		{TestKey: "test", CreatedBy: 1},
	}

	err = repo.CreateBatch(data)
	if err == nil {
		t.Error("CreateBatch() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_GetByID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "getbyid@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser)

	// Create test data
	testData := &domain.BenchmarkData{
		TestKey:    "get-by-id-key",
		TestValue:  "get-by-id-value",
		NumValue:   42.5,
		IntValue:   100,
		BoolValue:  true,
		LargeText:  "large text here",
		JsonBlob:   `{"test": true}`,
		ExtraFloat: 1.5,
		ExtraInt:   50,
		Category:   "testing",
		Priority:   3,
		CreatedBy:  testUser.ID,
	}
	repo.Create(testData)

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "existing ID",
			id:      testData.ID,
			wantErr: false,
		},
		{
			name:    "non-existing ID",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.ID != testData.ID {
					t.Errorf("GetByID() ID = %d, want %d", result.ID, testData.ID)
				}
				if result.TestKey != testData.TestKey {
					t.Errorf("GetByID() TestKey = %q, want %q", result.TestKey, testData.TestKey)
				}
				if result.TestValue != testData.TestValue {
					t.Errorf("GetByID() TestValue = %q, want %q", result.TestValue, testData.TestValue)
				}
				if result.NumValue != testData.NumValue {
					t.Errorf("GetByID() NumValue = %v, want %v", result.NumValue, testData.NumValue)
				}
				if result.IntValue != testData.IntValue {
					t.Errorf("GetByID() IntValue = %d, want %d", result.IntValue, testData.IntValue)
				}
				if result.BoolValue != testData.BoolValue {
					t.Errorf("GetByID() BoolValue = %v, want %v", result.BoolValue, testData.BoolValue)
				}
				if result.Category != testData.Category {
					t.Errorf("GetByID() Category = %q, want %q", result.Category, testData.Category)
				}
				if result.Priority != testData.Priority {
					t.Errorf("GetByID() Priority = %d, want %d", result.Priority, testData.Priority)
				}
			}
		})
	}
}

func TestBenchmarkRepository_GetByID_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	_, err = repo.GetByID(1)
	if err == nil {
		t.Error("GetByID() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_GetByKey(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "getbykey@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser)

	// Create test data
	testData := &domain.BenchmarkData{
		TestKey:   "unique-test-key",
		TestValue: "unique-value",
		NumValue:  123.0,
		CreatedBy: testUser.ID,
	}
	repo.Create(testData)

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "existing key",
			key:     "unique-test-key",
			wantErr: false,
		},
		{
			name:    "non-existing key",
			key:     "non-existent-key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.TestKey != tt.key {
					t.Errorf("GetByKey() TestKey = %q, want %q", result.TestKey, tt.key)
				}
			}
		})
	}
}

func TestBenchmarkRepository_GetByKey_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	_, err = repo.GetByKey("test")
	if err == nil {
		t.Error("GetByKey() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_List(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "list@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser)

	// Clean up and create test data
	repo.DeleteAll()

	for i := 0; i < 5; i++ {
		data := &domain.BenchmarkData{
			TestKey:   "list-key-" + string(rune('a'+i)),
			TestValue: "list-value",
			Priority:  i,
			CreatedBy: testUser.ID,
		}
		repo.Create(data)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
	}{
		{
			name:      "list all",
			limit:     100,
			offset:    0,
			wantCount: 5,
		},
		{
			name:      "list with limit",
			limit:     2,
			offset:    0,
			wantCount: 2,
		},
		{
			name:      "list with offset",
			limit:     100,
			offset:    3,
			wantCount: 2,
		},
		{
			name:      "list with limit and offset",
			limit:     2,
			offset:    2,
			wantCount: 2,
		},
		{
			name:      "list with high offset",
			limit:     10,
			offset:    100,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(result) != tt.wantCount {
				t.Errorf("List() returned %d items, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestBenchmarkRepository_List_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	_, err = repo.List(10, 0)
	if err == nil {
		t.Error("List() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_ListFiltered(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test users
	userRepo := NewSQLiteUserRepository(db)
	testUser1 := &domain.User{
		Email:        "filter1@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	testUser2 := &domain.User{
		Email:        "filter2@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser1)
	userRepo.Create(testUser2)

	// Clean up and create test data
	repo.DeleteAll()

	testDataSet := []*domain.BenchmarkData{
		{TestKey: "filter-a-1", Category: "category-a", BoolValue: true, Priority: 5, CreatedBy: testUser1.ID},
		{TestKey: "filter-a-2", Category: "category-a", BoolValue: false, Priority: 3, CreatedBy: testUser1.ID},
		{TestKey: "filter-b-1", Category: "category-b", BoolValue: true, Priority: 1, CreatedBy: testUser2.ID},
		{TestKey: "filter-b-2", Category: "category-b", BoolValue: false, Priority: 2, CreatedBy: testUser2.ID},
	}

	for _, data := range testDataSet {
		repo.Create(data)
	}

	// Helper functions for pointer values (local to avoid redeclaration)
	strPtr := func(s string) *string { return &s }
	boolPtr := func(b bool) *bool { return &b }
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name      string
		filters   domain.BenchmarkFilters
		wantCount int
	}{
		{
			name:      "no filters",
			filters:   domain.BenchmarkFilters{},
			wantCount: 4,
		},
		{
			name:      "filter by created_by",
			filters:   domain.BenchmarkFilters{CreatedBy: &testUser1.ID},
			wantCount: 2,
		},
		{
			name:      "filter by test_key prefix",
			filters:   domain.BenchmarkFilters{TestKey: strPtr("filter-a")},
			wantCount: 2,
		},
		{
			name:      "filter by bool_value true",
			filters:   domain.BenchmarkFilters{BoolValue: boolPtr(true)},
			wantCount: 2,
		},
		{
			name:      "filter by bool_value false",
			filters:   domain.BenchmarkFilters{BoolValue: boolPtr(false)},
			wantCount: 2,
		},
		{
			name:      "filter by category",
			filters:   domain.BenchmarkFilters{Category: strPtr("category-a")},
			wantCount: 2,
		},
		{
			name:      "filter by min_priority",
			filters:   domain.BenchmarkFilters{MinPriority: intPtr(3)},
			wantCount: 2,
		},
		{
			name: "combined filters",
			filters: domain.BenchmarkFilters{
				CreatedBy: &testUser1.ID,
				BoolValue: boolPtr(true),
			},
			wantCount: 1,
		},
		{
			name:      "filter with no results",
			filters:   domain.BenchmarkFilters{Category: strPtr("non-existent")},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.ListFiltered(tt.filters, 100, 0)
			if err != nil {
				t.Errorf("ListFiltered() error = %v", err)
				return
			}
			if len(result) != tt.wantCount {
				t.Errorf("ListFiltered() returned %d items, want %d", len(result), tt.wantCount)
			}
		})
	}
}

func TestBenchmarkRepository_Update(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "update@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser)

	// Create test data
	testData := &domain.BenchmarkData{
		TestKey:   "update-key",
		TestValue: "original-value",
		NumValue:  10.0,
		IntValue:  5,
		BoolValue: false,
		Category:  "original",
		Priority:  1,
		CreatedBy: testUser.ID,
	}
	repo.Create(testData)

	// Update the data
	testData.TestValue = "updated-value"
	testData.NumValue = 99.9
	testData.IntValue = 100
	testData.BoolValue = true
	testData.Category = "updated"
	testData.Priority = 10

	err = repo.Update(testData)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	result, err := repo.GetByID(testData.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	if result.TestValue != "updated-value" {
		t.Errorf("Update() TestValue = %q, want %q", result.TestValue, "updated-value")
	}
	if result.NumValue != 99.9 {
		t.Errorf("Update() NumValue = %v, want %v", result.NumValue, 99.9)
	}
	if result.IntValue != 100 {
		t.Errorf("Update() IntValue = %d, want %d", result.IntValue, 100)
	}
	if result.BoolValue != true {
		t.Errorf("Update() BoolValue = %v, want %v", result.BoolValue, true)
	}
	if result.Category != "updated" {
		t.Errorf("Update() Category = %q, want %q", result.Category, "updated")
	}
	if result.Priority != 10 {
		t.Errorf("Update() Priority = %d, want %d", result.Priority, 10)
	}
}

func TestBenchmarkRepository_Update_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	data := &domain.BenchmarkData{ID: 1, TestKey: "test"}
	err = repo.Update(data)
	if err == nil {
		t.Error("Update() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_Delete(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "delete@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser)

	// Create test data
	testData := &domain.BenchmarkData{
		TestKey:   "delete-key",
		CreatedBy: testUser.ID,
	}
	repo.Create(testData)

	// Verify it exists
	_, err = repo.GetByID(testData.ID)
	if err != nil {
		t.Fatalf("Record should exist before delete: %v", err)
	}

	// Delete
	err = repo.Delete(testData.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify it's gone
	_, err = repo.GetByID(testData.ID)
	if err == nil {
		t.Error("Record should not exist after delete")
	}
}

func TestBenchmarkRepository_Delete_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	err = repo.Delete(1)
	if err == nil {
		t.Error("Delete() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_DeleteByUserID(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test users
	userRepo := NewSQLiteUserRepository(db)
	testUser1 := &domain.User{
		Email:        "deleteuser1@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	testUser2 := &domain.User{
		Email:        "deleteuser2@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser1)
	userRepo.Create(testUser2)

	// Clean up
	repo.DeleteAll()

	// Create test data for both users
	for i := 0; i < 3; i++ {
		repo.Create(&domain.BenchmarkData{
			TestKey:   "user1-key-" + string(rune('a'+i)),
			CreatedBy: testUser1.ID,
		})
	}
	for i := 0; i < 2; i++ {
		repo.Create(&domain.BenchmarkData{
			TestKey:   "user2-key-" + string(rune('a'+i)),
			CreatedBy: testUser2.ID,
		})
	}

	// Verify initial count
	count, _ := repo.Count()
	if count != 5 {
		t.Fatalf("Initial count should be 5, got %d", count)
	}

	// Delete user1's data
	err = repo.DeleteByUserID(testUser1.ID)
	if err != nil {
		t.Fatalf("DeleteByUserID() error = %v", err)
	}

	// Verify user1's data is gone
	filters := domain.BenchmarkFilters{CreatedBy: &testUser1.ID}
	user1Data, _ := repo.ListFiltered(filters, 100, 0)
	if len(user1Data) != 0 {
		t.Errorf("User1's data should be deleted, found %d records", len(user1Data))
	}

	// Verify user2's data still exists
	filters = domain.BenchmarkFilters{CreatedBy: &testUser2.ID}
	user2Data, _ := repo.ListFiltered(filters, 100, 0)
	if len(user2Data) != 2 {
		t.Errorf("User2's data should still exist, found %d records", len(user2Data))
	}
}

func TestBenchmarkRepository_DeleteByUserID_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	err = repo.DeleteByUserID(1)
	if err == nil {
		t.Error("DeleteByUserID() with unsupported driver should return error")
	}
}

func TestBenchmarkRepository_DeleteAll(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "deleteall@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser)

	// Create test data
	for i := 0; i < 5; i++ {
		repo.Create(&domain.BenchmarkData{
			TestKey:   "delete-all-" + string(rune('a'+i)),
			CreatedBy: testUser.ID,
		})
	}

	// Verify data exists
	count, _ := repo.Count()
	if count == 0 {
		t.Fatal("Should have data before DeleteAll")
	}

	// Delete all
	err = repo.DeleteAll()
	if err != nil {
		t.Fatalf("DeleteAll() error = %v", err)
	}

	// Verify all data is gone
	count, _ = repo.Count()
	if count != 0 {
		t.Errorf("Count after DeleteAll should be 0, got %d", count)
	}
}

func TestBenchmarkRepository_Count(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	// Create test user
	userRepo := NewSQLiteUserRepository(db)
	testUser := &domain.User{
		Email:        "count@example.com",
		PasswordHash: "hash",
		Role:         "user",
	}
	userRepo.Create(testUser)

	// Clean up
	repo.DeleteAll()

	// Count should be 0
	count, err := repo.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 0 {
		t.Errorf("Count() = %d, want 0", count)
	}

	// Add some data
	for i := 0; i < 3; i++ {
		repo.Create(&domain.BenchmarkData{
			TestKey:   "count-" + string(rune('a'+i)),
			CreatedBy: testUser.ID,
		})
	}

	// Count should be 3
	count, err = repo.Count()
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 3 {
		t.Errorf("Count() = %d, want 3", count)
	}
}

func TestBenchmarkRepository_Driver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	tests := []struct {
		name   string
		driver string
	}{
		{name: "sqlite3", driver: "sqlite3"},
		{name: "postgres", driver: "postgres"},
		{name: "mysql", driver: "mysql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewBenchmarkRepository(db, tt.driver)
			if repo.Driver() != tt.driver {
				t.Errorf("Driver() = %q, want %q", repo.Driver(), tt.driver)
			}
		})
	}
}

func TestBenchmarkRepository_DatabaseVersion(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "sqlite3")

	version := repo.DatabaseVersion()
	if version == "" || version == "unknown" {
		t.Error("DatabaseVersion() should return a version string for sqlite3")
	}
}

func TestBenchmarkRepository_DatabaseVersion_UnsupportedDriver(t *testing.T) {
	db, cleanup, err := SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	defer cleanup()

	repo := NewBenchmarkRepository(db, "unsupported")

	version := repo.DatabaseVersion()
	if version != "unknown" {
		t.Errorf("DatabaseVersion() = %q, want %q for unsupported driver", version, "unknown")
	}
}

func TestRebindQueryBenchmark(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		driver   string
		expected string
	}{
		{
			name:     "postgres with single placeholder",
			query:    "SELECT * FROM users WHERE id = ?",
			driver:   "postgres",
			expected: "SELECT * FROM users WHERE id = $1",
		},
		{
			name:     "postgres with multiple placeholders",
			query:    "SELECT * FROM users WHERE id = ? AND name = ? AND active = ?",
			driver:   "postgres",
			expected: "SELECT * FROM users WHERE id = $1 AND name = $2 AND active = $3",
		},
		{
			name:     "postgres with insert placeholders",
			query:    "INSERT INTO users (a, b, c, d) VALUES (?, ?, ?, ?)",
			driver:   "postgres",
			expected: "INSERT INTO users (a, b, c, d) VALUES ($1, $2, $3, $4)",
		},
		{
			name:     "postgres with no placeholders",
			query:    "SELECT * FROM users",
			driver:   "postgres",
			expected: "SELECT * FROM users",
		},
		{
			name:     "sqlite3 unchanged",
			query:    "SELECT * FROM users WHERE id = ?",
			driver:   "sqlite3",
			expected: "SELECT * FROM users WHERE id = ?",
		},
		{
			name:     "mysql unchanged",
			query:    "SELECT * FROM users WHERE id = ? AND name = ?",
			driver:   "mysql",
			expected: "SELECT * FROM users WHERE id = ? AND name = ?",
		},
		{
			name:     "empty driver unchanged",
			query:    "SELECT * FROM users WHERE id = ?",
			driver:   "",
			expected: "SELECT * FROM users WHERE id = ?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rebindQueryBenchmark(tt.query, tt.driver)
			if result != tt.expected {
				t.Errorf("rebindQueryBenchmark() = %q, want %q", result, tt.expected)
			}
		})
	}
}
