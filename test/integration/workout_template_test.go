package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/johnzastrow/actalog/pkg/auth"
)

// TestWorkoutTemplateInstructionsField tests the full CRUD lifecycle of the Instructions field
func TestWorkoutTemplateInstructionsField(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	// Generate a valid token for authenticated requests
	token, err := auth.GenerateToken(1, "template-test@example.com", "user", testJWTSecret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	var createdTemplateID float64

	// Test 1: Create template with movement instructions
	t.Run("Create_Template_With_Movement_Instructions", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "Test Template with Instructions",
			"notes": "Template-level notes",
			"movements": []map[string]interface{}{
				{
					"movement_id":  1, // Back Squat (seeded)
					"sets":         4,
					"reps":         8,
					"weight":       185.5,
					"instructions": "**Setup:**\n- Position bar on traps\n- Feet shoulder width apart\n\n**Execution:**\n1. Brace core\n2. Break at hips first\n3. Descend to parallel",
					"notes":        "Progressive overload week 3",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		template, ok := response["template"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected template in response, got: %v", response)
		}

		createdTemplateID = template["id"].(float64)
		t.Logf("Created template ID: %v", createdTemplateID)
	})

	// Test 2: Get template and verify instructions were stored
	t.Run("Get_Template_Verify_Instructions", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID from previous test")
		}

		req := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		template, ok := response["template"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected template in response")
		}

		movements, ok := template["movements"].([]interface{})
		if !ok || len(movements) == 0 {
			t.Fatalf("Expected movements in template")
		}

		movement := movements[0].(map[string]interface{})

		// Verify instructions field was stored correctly
		instructions, ok := movement["instructions"].(string)
		if !ok {
			t.Errorf("Expected instructions field to be a string")
		}

		expectedInstructions := "**Setup:**\n- Position bar on traps\n- Feet shoulder width apart\n\n**Execution:**\n1. Brace core\n2. Break at hips first\n3. Descend to parallel"
		if instructions != expectedInstructions {
			t.Errorf("Instructions mismatch.\nExpected: %s\nGot: %s", expectedInstructions, instructions)
		}

		// Verify other fields were also stored
		if notes, ok := movement["notes"].(string); !ok || notes != "Progressive overload week 3" {
			t.Errorf("Notes mismatch: %v", movement["notes"])
		}
	})

	// Test 3: Update template with new instructions
	t.Run("Update_Template_Instructions", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID from previous test")
		}

		body := map[string]interface{}{
			"name":  "Updated Template with Instructions",
			"notes": "Updated template notes",
			"movements": []map[string]interface{}{
				{
					"movement_id":  1,
					"sets":         5,
					"reps":         5,
					"weight":       200,
					"instructions": "## Updated Instructions\n\n- New step 1\n- New step 2\n\n> Safety note: Always use a spotter",
					"notes":        "Updated notes",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/templates/"+formatID(createdTemplateID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	// Test 4: Verify updated instructions
	t.Run("Verify_Updated_Instructions", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID from previous test")
		}

		req := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		movements := template["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		expectedInstructions := "## Updated Instructions\n\n- New step 1\n- New step 2\n\n> Safety note: Always use a spotter"
		if instructions := movement["instructions"].(string); instructions != expectedInstructions {
			t.Errorf("Updated instructions mismatch.\nExpected: %s\nGot: %s", expectedInstructions, instructions)
		}

		// Verify sets were also updated
		if sets := movement["sets"].(float64); sets != 5 {
			t.Errorf("Sets should be 5, got %v", sets)
		}
	})

	// Test 5: Clear instructions (set to empty string)
	t.Run("Clear_Instructions", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID from previous test")
		}

		body := map[string]interface{}{
			"name":  "Template with Cleared Instructions",
			"notes": "Notes remain",
			"movements": []map[string]interface{}{
				{
					"movement_id":  1,
					"sets":         3,
					"reps":         10,
					"instructions": "", // Clear instructions
					"notes":        "Notes still here",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/templates/"+formatID(createdTemplateID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify instructions were cleared
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var response map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		movements := template["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		// Instructions should be empty or nil (omitempty may exclude it)
		instructions, ok := movement["instructions"]
		if ok && instructions != nil && instructions != "" {
			t.Errorf("Instructions should be empty, got: %v", instructions)
		}

		// Notes should still be present
		notes, ok := movement["notes"].(string)
		if !ok || notes != "Notes still here" {
			t.Errorf("Notes should remain, got: %v", movement["notes"])
		}
	})
}

// TestWorkoutTemplateWODInstructionsField tests WOD instructions in templates
func TestWorkoutTemplateWODInstructionsField(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "wod-test@example.com", "user", testJWTSecret, time.Hour)

	var createdTemplateID float64

	// Test 1: Create template with WOD instructions
	t.Run("Create_Template_With_WOD_Instructions", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "WOD Template with Instructions",
			"wods": []map[string]interface{}{
				{
					"wod_id":       1, // Fran (seeded)
					"instructions": "## Scaling Options\n\n**Rx:**\n- Thrusters: 95/65 lbs\n- Pull-ups: Kipping or Butterfly\n\n**Scaled:**\n- Thrusters: 65/45 lbs\n- Ring Rows",
					"notes":        "Time cap: 12 minutes",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		createdTemplateID = template["id"].(float64)
	})

	// Test 2: Verify WOD instructions were stored
	t.Run("Verify_WOD_Instructions", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID")
		}

		req := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		wods, ok := template["wods"].([]interface{})
		if !ok || len(wods) == 0 {
			t.Fatalf("Expected wods in template")
		}

		wod := wods[0].(map[string]interface{})

		expectedInstructions := "## Scaling Options\n\n**Rx:**\n- Thrusters: 95/65 lbs\n- Pull-ups: Kipping or Butterfly\n\n**Scaled:**\n- Thrusters: 65/45 lbs\n- Ring Rows"
		if instructions := wod["instructions"].(string); instructions != expectedInstructions {
			t.Errorf("WOD instructions mismatch.\nExpected: %s\nGot: %s", expectedInstructions, instructions)
		}

		if notes := wod["notes"].(string); notes != "Time cap: 12 minutes" {
			t.Errorf("WOD notes mismatch: %s", notes)
		}
	})
}

// TestWorkoutTemplateSpecialCharactersInInstructions tests special characters
func TestWorkoutTemplateSpecialCharactersInInstructions(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "special-chars@example.com", "user", testJWTSecret, time.Hour)

	testCases := []struct {
		name         string
		instructions string
	}{
		{
			name:         "Markdown_Formatting",
			instructions: "# Header\n## Subheader\n\n**Bold** and *italic*\n\n```code block```\n\n> Blockquote\n\n| Col1 | Col2 |\n|------|------|\n| A    | B    |",
		},
		{
			name:         "Unicode_Characters",
			instructions: "café naïve 日本語 中文 emoji: 💪🏋️‍♂️🔥",
		},
		{
			name:         "HTML_Like_Characters",
			instructions: "<angle brackets> & ampersand \"double quotes\" 'single quotes'",
		},
		{
			name:         "Newlines_And_Tabs",
			instructions: "Line 1\nLine 2\n\nParagraph 2\n\n\tTabbed content",
		},
		{
			name:         "Special_Symbols",
			instructions: "Weight: 100kg @ 80% 1RM | Rest: 2-3 min | RPE: 8/10",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{
				"name": "Special Chars Template - " + tc.name,
				"movements": []map[string]interface{}{
					{
						"movement_id":  1,
						"sets":         3,
						"reps":         10,
						"instructions": tc.instructions,
					},
				},
			}
			jsonBody, _ := json.Marshal(body)

			// Create
			req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Fatalf("Create failed with %d: %s", rec.Code, rec.Body.String())
			}

			var createResp map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &createResp)
			template := createResp["template"].(map[string]interface{})
			templateID := template["id"].(float64)

			// Read back
			getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
			getRec := httptest.NewRecorder()
			router.ServeHTTP(getRec, getReq)

			if getRec.Code != http.StatusOK {
				t.Fatalf("Get failed with %d", getRec.Code)
			}

			var getResp map[string]interface{}
			json.Unmarshal(getRec.Body.Bytes(), &getResp)

			retrievedTemplate := getResp["template"].(map[string]interface{})
			movements := retrievedTemplate["movements"].([]interface{})
			movement := movements[0].(map[string]interface{})

			retrievedInstructions := movement["instructions"].(string)
			if retrievedInstructions != tc.instructions {
				t.Errorf("Instructions not preserved.\nExpected: %q\nGot: %q", tc.instructions, retrievedInstructions)
			}
		})
	}
}

// TestWorkoutTemplateMultipleMovementsWithInstructions tests multiple movements
func TestWorkoutTemplateMultipleMovementsWithInstructions(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "multi-movement@example.com", "user", testJWTSecret, time.Hour)

	// Create template with multiple movements, each with instructions
	body := map[string]interface{}{
		"name": "Multi-Movement Template",
		"movements": []map[string]interface{}{
			{
				"movement_id":  1,
				"sets":         5,
				"reps":         5,
				"instructions": "Movement 1 instructions - Back Squat technique",
				"notes":        "Movement 1 notes",
			},
			{
				"movement_id":  2,
				"sets":         5,
				"reps":         5,
				"instructions": "Movement 2 instructions - Bench Press setup",
				"notes":        "Movement 2 notes",
			},
			{
				"movement_id":  3,
				"sets":         5,
				"reps":         5,
				"instructions": "Movement 3 instructions - Deadlift cues",
				"notes":        "Movement 3 notes",
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var createResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &createResp)
	template := createResp["template"].(map[string]interface{})
	templateID := template["id"].(float64)

	// Verify all movements have their instructions
	getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	var getResp map[string]interface{}
	json.Unmarshal(getRec.Body.Bytes(), &getResp)

	retrievedTemplate := getResp["template"].(map[string]interface{})
	movements := retrievedTemplate["movements"].([]interface{})

	if len(movements) != 3 {
		t.Fatalf("Expected 3 movements, got %d", len(movements))
	}

	for i, m := range movements {
		movement := m.(map[string]interface{})
		expectedInstructions := "Movement " + formatID(float64(i+1)) + " instructions"

		instructions := movement["instructions"].(string)
		if instructions == "" {
			t.Errorf("Movement %d instructions should not be empty", i+1)
		}

		notes := movement["notes"].(string)
		if notes == "" {
			t.Errorf("Movement %d notes should not be empty", i+1)
		}

		_ = expectedInstructions // Used for debugging
	}
}

// TestWorkoutTemplateInstructionsOptional verifies instructions field is optional
func TestWorkoutTemplateInstructionsOptional(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "optional-test@example.com", "user", testJWTSecret, time.Hour)

	// Create template without instructions field
	body := map[string]interface{}{
		"name": "Template Without Instructions",
		"movements": []map[string]interface{}{
			{
				"movement_id": 1,
				"sets":        3,
				"reps":        10,
				// No instructions field
				"notes": "Just notes, no instructions",
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should succeed - instructions is optional
	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var createResp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &createResp)
	template := createResp["template"].(map[string]interface{})
	templateID := template["id"].(float64)

	// Verify we can read it back
	getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", getRec.Code)
	}

	var getResp map[string]interface{}
	json.Unmarshal(getRec.Body.Bytes(), &getResp)

	retrievedTemplate := getResp["template"].(map[string]interface{})
	movements := retrievedTemplate["movements"].([]interface{})
	movement := movements[0].(map[string]interface{})

	// Instructions should be empty string or not present
	instructions, exists := movement["instructions"]
	if exists && instructions != "" && instructions != nil {
		t.Logf("Instructions field present with value: %v", instructions)
	}

	// Notes should still be present
	notes := movement["notes"].(string)
	if notes != "Just notes, no instructions" {
		t.Errorf("Notes mismatch: %s", notes)
	}
}

// formatID converts a float64 ID to string for URL path
func formatID(id float64) string {
	return fmt.Sprintf("%.0f", id)
}

// =============================================================================
// MOVEMENT ATTRIBUTE TESTS
// =============================================================================

// TestWorkoutTemplateMovementTimeAndDistance tests time and distance attributes
func TestWorkoutTemplateMovementTimeAndDistance(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "time-distance@example.com", "user", testJWTSecret, time.Hour)

	var createdTemplateID float64

	// Test 1: Create template with time-based movement (e.g., plank hold)
	t.Run("Create_With_Time_Attribute", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Time-Based Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"sets":        3,
					"time":        60, // 60 seconds
					"notes":       "Hold for full duration",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		createdTemplateID = template["id"].(float64)
	})

	// Test 2: Verify time was stored
	t.Run("Verify_Time_Stored", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID")
		}

		req := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		movements := template["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		timeVal, ok := movement["time"].(float64)
		if !ok {
			t.Errorf("Expected time field to be present")
		}
		if timeVal != 60 {
			t.Errorf("Expected time=60, got %v", timeVal)
		}
	})

	// Test 3: Create template with distance-based movement (e.g., run)
	t.Run("Create_With_Distance_Attribute", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Distance-Based Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"sets":        1,
					"distance":    400.5, // 400.5 meters
					"notes":       "Sprint pace",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Verify distance
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		movements := retrievedTemplate["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		distanceVal, ok := movement["distance"].(float64)
		if !ok {
			t.Errorf("Expected distance field to be present")
		}
		if distanceVal != 400.5 {
			t.Errorf("Expected distance=400.5, got %v", distanceVal)
		}
	})

	// Test 4: Update time and distance
	t.Run("Update_Time_And_Distance", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID")
		}

		body := map[string]interface{}{
			"name": "Updated Time-Distance Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"sets":        4,
					"time":        90,    // Updated to 90 seconds
					"distance":    800.0, // Added distance
					"notes":       "Updated timing",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/templates/"+formatID(createdTemplateID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify updates
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var response map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		movements := template["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		if timeVal := movement["time"].(float64); timeVal != 90 {
			t.Errorf("Expected time=90, got %v", timeVal)
		}
		if distanceVal := movement["distance"].(float64); distanceVal != 800.0 {
			t.Errorf("Expected distance=800.0, got %v", distanceVal)
		}
	})
}

// TestWorkoutTemplateMovementIsRxAndIsPR tests is_rx and is_pr flags
func TestWorkoutTemplateMovementIsRxAndIsPR(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "rx-pr@example.com", "user", testJWTSecret, time.Hour)

	var createdTemplateID float64

	// Test 1: Create template with is_rx=true
	t.Run("Create_With_IsRx_True", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Rx Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"sets":        5,
					"reps":        5,
					"weight":      225,
					"is_rx":       true,
					"notes":       "Prescribed weight",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		createdTemplateID = template["id"].(float64)
	})

	// Test 2: Verify is_rx was stored
	t.Run("Verify_IsRx_Stored", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID")
		}

		req := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		movements := template["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		isRx, ok := movement["is_rx"].(bool)
		if !ok {
			t.Errorf("Expected is_rx field to be a bool")
		}
		if !isRx {
			t.Errorf("Expected is_rx=true, got %v", isRx)
		}
	})

	// Test 3: Create template with is_pr=true
	t.Run("Create_With_IsPR_True", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "PR Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"sets":        1,
					"reps":        1,
					"weight":      315,
					"is_pr":       true,
					"notes":       "Personal record attempt",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Verify is_pr
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		movements := retrievedTemplate["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		isPR, ok := movement["is_pr"].(bool)
		if !ok {
			t.Errorf("Expected is_pr field to be a bool")
		}
		if !isPR {
			t.Errorf("Expected is_pr=true, got %v", isPR)
		}
	})

	// Test 4: Create with both is_rx and is_pr true
	t.Run("Create_With_Both_Flags", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Rx PR Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"sets":        1,
					"reps":        1,
					"weight":      405,
					"is_rx":       true,
					"is_pr":       true,
					"notes":       "Rx PR!",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Verify both flags
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		movements := retrievedTemplate["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		if isRx := movement["is_rx"].(bool); !isRx {
			t.Errorf("Expected is_rx=true")
		}
		if isPR := movement["is_pr"].(bool); !isPR {
			t.Errorf("Expected is_pr=true")
		}
	})

	// Test 5: Update flags from true to false
	t.Run("Update_Flags_To_False", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID")
		}

		body := map[string]interface{}{
			"name": "Updated Flags Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1,
					"sets":        5,
					"reps":        5,
					"weight":      185,
					"is_rx":       false, // Changed from true
					"is_pr":       false,
					"notes":       "Scaled workout",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/templates/"+formatID(createdTemplateID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify updates
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var response map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		movements := template["movements"].([]interface{})
		movement := movements[0].(map[string]interface{})

		// Note: is_rx and is_pr default to false, so they may not be present if false
		isRx, _ := movement["is_rx"].(bool)
		isPR, _ := movement["is_pr"].(bool)

		if isRx {
			t.Errorf("Expected is_rx=false, got true")
		}
		if isPR {
			t.Errorf("Expected is_pr=false, got true")
		}
	})
}

// TestWorkoutTemplateMovementOrderIndex tests order_index preservation
func TestWorkoutTemplateMovementOrderIndex(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "order-test@example.com", "user", testJWTSecret, time.Hour)

	// Test 1: Create template with multiple movements - verify order is preserved
	t.Run("Create_Multiple_Movements_Order_Preserved", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Ordered Movements Template",
			"movements": []map[string]interface{}{
				{
					"movement_id": 1, // First movement
					"sets":        5,
					"reps":        5,
					"notes":       "First - Warm up",
				},
				{
					"movement_id": 2, // Second movement
					"sets":        3,
					"reps":        8,
					"notes":       "Second - Main lift",
				},
				{
					"movement_id": 3, // Third movement
					"sets":        3,
					"reps":        12,
					"notes":       "Third - Accessory",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Retrieve and verify order
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		movements := retrievedTemplate["movements"].([]interface{})

		if len(movements) != 3 {
			t.Fatalf("Expected 3 movements, got %d", len(movements))
		}

		// Verify order by notes (which indicate expected position)
		expectedOrder := []string{"First - Warm up", "Second - Main lift", "Third - Accessory"}
		for i, m := range movements {
			movement := m.(map[string]interface{})
			notes := movement["notes"].(string)
			if notes != expectedOrder[i] {
				t.Errorf("Movement %d: expected notes %q, got %q", i, expectedOrder[i], notes)
			}

			// Also verify order_index if present (1-based)
			if orderIndex, ok := movement["order_index"].(float64); ok {
				if int(orderIndex) != i+1 {
					t.Errorf("Movement %d: expected order_index=%d, got %v", i, i+1, orderIndex)
				}
			}
		}
	})

	// Test 2: Update and reorder movements
	t.Run("Update_Reorder_Movements", func(t *testing.T) {
		// Create initial template
		body := map[string]interface{}{
			"name": "Reorder Test Template",
			"movements": []map[string]interface{}{
				{"movement_id": 1, "sets": 3, "reps": 10, "notes": "A"},
				{"movement_id": 2, "sets": 3, "reps": 10, "notes": "B"},
				{"movement_id": 3, "sets": 3, "reps": 10, "notes": "C"},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Update with reordered movements (C, A, B)
		updateBody := map[string]interface{}{
			"name": "Reordered Template",
			"movements": []map[string]interface{}{
				{"movement_id": 3, "sets": 3, "reps": 10, "notes": "C"}, // Was third, now first
				{"movement_id": 1, "sets": 3, "reps": 10, "notes": "A"}, // Was first, now second
				{"movement_id": 2, "sets": 3, "reps": 10, "notes": "B"}, // Was second, now third
			},
		}
		updateJsonBody, _ := json.Marshal(updateBody)

		updateReq := httptest.NewRequest("PUT", "/api/templates/"+formatID(templateID), bytes.NewBuffer(updateJsonBody))
		updateReq.Header.Set("Content-Type", "application/json")
		updateReq.Header.Set("Authorization", "Bearer "+token)
		updateRec := httptest.NewRecorder()

		router.ServeHTTP(updateRec, updateReq)

		if updateRec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", updateRec.Code, updateRec.Body.String())
		}

		// Verify new order
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		movements := retrievedTemplate["movements"].([]interface{})

		expectedOrder := []string{"C", "A", "B"}
		for i, m := range movements {
			movement := m.(map[string]interface{})
			notes := movement["notes"].(string)
			if notes != expectedOrder[i] {
				t.Errorf("After reorder, position %d: expected %q, got %q", i, expectedOrder[i], notes)
			}
		}
	})
}

// =============================================================================
// WOD ATTRIBUTE TESTS
// =============================================================================

// TestWorkoutTemplateWODScoreValueAndDivision tests score_value and division attributes
func TestWorkoutTemplateWODScoreValueAndDivision(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "wod-attrs@example.com", "user", testJWTSecret, time.Hour)

	var createdTemplateID float64

	// Test 1: Create template with WOD score_value (time-based)
	t.Run("Create_With_ScoreValue_Time", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Timed WOD Template",
			"wods": []map[string]interface{}{
				{
					"wod_id":      1,
					"score_value": "5:32", // Time format
					"division":    "rx",
					"notes":       "Target time",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		createdTemplateID = template["id"].(float64)
	})

	// Test 2: Verify score_value and division were stored
	t.Run("Verify_ScoreValue_And_Division", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID")
		}

		req := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		wods := template["wods"].([]interface{})
		wod := wods[0].(map[string]interface{})

		scoreValue, ok := wod["score_value"].(string)
		if !ok || scoreValue != "5:32" {
			t.Errorf("Expected score_value='5:32', got %v", wod["score_value"])
		}

		division, ok := wod["division"].(string)
		if !ok || division != "rx" {
			t.Errorf("Expected division='rx', got %v", wod["division"])
		}
	})

	// Test 3: Create with different divisions
	t.Run("Create_With_Different_Divisions", func(t *testing.T) {
		divisions := []string{"rx", "scaled", "beginner"}

		for _, div := range divisions {
			body := map[string]interface{}{
				"name": "Division Test - " + div,
				"wods": []map[string]interface{}{
					{
						"wod_id":      1,
						"score_value": "10:00",
						"division":    div,
						"notes":       div + " division",
					},
				},
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusCreated {
				t.Errorf("Division %s: Expected 201, got %d", div, rec.Code)
				continue
			}

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)
			template := response["template"].(map[string]interface{})
			templateID := template["id"].(float64)

			// Verify
			getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
			getRec := httptest.NewRecorder()
			router.ServeHTTP(getRec, getReq)

			var getResp map[string]interface{}
			json.Unmarshal(getRec.Body.Bytes(), &getResp)

			retrievedTemplate := getResp["template"].(map[string]interface{})
			wods := retrievedTemplate["wods"].([]interface{})
			wod := wods[0].(map[string]interface{})

			if retrievedDiv := wod["division"].(string); retrievedDiv != div {
				t.Errorf("Division %s: Expected %s, got %s", div, div, retrievedDiv)
			}
		}
	})

	// Test 4: Create with rounds+reps score format
	t.Run("Create_With_RoundsReps_Score", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "AMRAP WOD Template",
			"wods": []map[string]interface{}{
				{
					"wod_id":      1,
					"score_value": "8+15", // 8 rounds + 15 reps
					"division":    "scaled",
					"notes":       "12 min AMRAP",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Verify
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		wods := retrievedTemplate["wods"].([]interface{})
		wod := wods[0].(map[string]interface{})

		if scoreValue := wod["score_value"].(string); scoreValue != "8+15" {
			t.Errorf("Expected score_value='8+15', got %s", scoreValue)
		}
	})

	// Test 5: Update score_value and division
	t.Run("Update_ScoreValue_And_Division", func(t *testing.T) {
		if createdTemplateID == 0 {
			t.Skip("No template ID")
		}

		body := map[string]interface{}{
			"name": "Updated WOD Template",
			"wods": []map[string]interface{}{
				{
					"wod_id":      1,
					"score_value": "4:58", // Improved time
					"division":    "scaled",
					"notes":       "Updated",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/templates/"+formatID(createdTemplateID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify updates
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(createdTemplateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var response map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})
		wods := template["wods"].([]interface{})
		wod := wods[0].(map[string]interface{})

		if scoreValue := wod["score_value"].(string); scoreValue != "4:58" {
			t.Errorf("Expected score_value='4:58', got %s", scoreValue)
		}
		if division := wod["division"].(string); division != "scaled" {
			t.Errorf("Expected division='scaled', got %s", division)
		}
	})
}

// TestWorkoutTemplateWODIsPR tests is_pr flag for WODs
func TestWorkoutTemplateWODIsPR(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "wod-pr@example.com", "user", testJWTSecret, time.Hour)

	// Test 1: Create WOD with is_pr=true
	t.Run("Create_WOD_With_IsPR_True", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "PR WOD Template",
			"wods": []map[string]interface{}{
				{
					"wod_id":      1,
					"score_value": "2:58",
					"division":    "rx",
					"is_pr":       true,
					"notes":       "Personal record!",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Verify is_pr
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		wods := retrievedTemplate["wods"].([]interface{})
		wod := wods[0].(map[string]interface{})

		isPR, ok := wod["is_pr"].(bool)
		if !ok || !isPR {
			t.Errorf("Expected is_pr=true, got %v", wod["is_pr"])
		}
	})

	// Test 2: Create WOD with is_pr=false (default)
	t.Run("Create_WOD_Without_PR_Flag", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Non-PR WOD Template",
			"wods": []map[string]interface{}{
				{
					"wod_id":      1,
					"score_value": "6:00",
					"division":    "scaled",
					// is_pr not specified
					"notes": "Regular workout",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Verify is_pr defaults to false
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		wods := retrievedTemplate["wods"].([]interface{})
		wod := wods[0].(map[string]interface{})

		// is_pr should be false or not present
		isPR, _ := wod["is_pr"].(bool)
		if isPR {
			t.Errorf("Expected is_pr=false (default), got true")
		}
	})
}

// TestWorkoutTemplateWODOrderIndex tests order_index preservation for WODs
func TestWorkoutTemplateWODOrderIndex(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "wod-order@example.com", "user", testJWTSecret, time.Hour)

	// Test 1: Create template with multiple WODs - verify order
	t.Run("Create_Multiple_WODs_Order_Preserved", func(t *testing.T) {
		body := map[string]interface{}{
			"name": "Multi-WOD Template",
			"wods": []map[string]interface{}{
				{
					"wod_id": 1,
					"notes":  "First WOD",
				},
				{
					"wod_id": 2,
					"notes":  "Second WOD",
				},
				{
					"wod_id": 1, // Same WOD can appear multiple times
					"notes":  "Third WOD",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID := template["id"].(float64)

		// Verify order
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var getResp map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &getResp)

		retrievedTemplate := getResp["template"].(map[string]interface{})
		wods := retrievedTemplate["wods"].([]interface{})

		if len(wods) != 3 {
			t.Fatalf("Expected 3 WODs, got %d", len(wods))
		}

		expectedOrder := []string{"First WOD", "Second WOD", "Third WOD"}
		for i, w := range wods {
			wod := w.(map[string]interface{})
			notes := wod["notes"].(string)
			if notes != expectedOrder[i] {
				t.Errorf("WOD %d: expected notes %q, got %q", i, expectedOrder[i], notes)
			}
		}
	})
}

// =============================================================================
// COMPREHENSIVE CRUD TEST
// =============================================================================

// TestWorkoutTemplateFullCRUDAllAttributes tests complete CRUD with all attributes
func TestWorkoutTemplateFullCRUDAllAttributes(t *testing.T) {
	router, cleanup := setupFullRouter(t)
	defer cleanup()

	token, _ := auth.GenerateToken(1, "full-crud@example.com", "user", testJWTSecret, time.Hour)

	var templateID float64

	// CREATE with all attributes
	t.Run("Create_Full_Template", func(t *testing.T) {
		body := map[string]interface{}{
			"name":  "Complete Template",
			"notes": "A template with all attributes populated",
			"movements": []map[string]interface{}{
				{
					"movement_id":  1,
					"weight":       225.5,
					"sets":         5,
					"reps":         5,
					"time":         120,
					"distance":     1000.0,
					"is_rx":        true,
					"is_pr":        false,
					"instructions": "Full setup instructions with **markdown**",
					"notes":        "Movement notes",
				},
				{
					"movement_id":  2,
					"weight":       135.0,
					"sets":         3,
					"reps":         10,
					"is_rx":        false,
					"is_pr":        true,
					"instructions": "Second movement instructions",
					"notes":        "PR attempt",
				},
			},
			"wods": []map[string]interface{}{
				{
					"wod_id":       1,
					"score_value":  "4:30",
					"division":     "rx",
					"is_pr":        true,
					"instructions": "WOD instructions with scaling options",
					"notes":        "Time cap 10 min",
				},
				{
					"wod_id":       2,
					"score_value":  "12+8",
					"division":     "scaled",
					"is_pr":        false,
					"instructions": "AMRAP instructions",
					"notes":        "Pace yourself",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/templates", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)
		template := response["template"].(map[string]interface{})
		templateID = template["id"].(float64)
		t.Logf("Created template ID: %.0f", templateID)
	})

	// READ and verify all attributes
	t.Run("Read_Verify_All_Attributes", func(t *testing.T) {
		if templateID == 0 {
			t.Skip("No template ID")
		}

		req := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", rec.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})

		// Verify template-level
		if name := template["name"].(string); name != "Complete Template" {
			t.Errorf("Name mismatch: %s", name)
		}

		// Verify movements
		movements := template["movements"].([]interface{})
		if len(movements) != 2 {
			t.Fatalf("Expected 2 movements, got %d", len(movements))
		}

		m1 := movements[0].(map[string]interface{})
		if weight := m1["weight"].(float64); weight != 225.5 {
			t.Errorf("Movement 1 weight: expected 225.5, got %v", weight)
		}
		if sets := m1["sets"].(float64); sets != 5 {
			t.Errorf("Movement 1 sets: expected 5, got %v", sets)
		}
		if reps := m1["reps"].(float64); reps != 5 {
			t.Errorf("Movement 1 reps: expected 5, got %v", reps)
		}
		if timeVal := m1["time"].(float64); timeVal != 120 {
			t.Errorf("Movement 1 time: expected 120, got %v", timeVal)
		}
		if distance := m1["distance"].(float64); distance != 1000.0 {
			t.Errorf("Movement 1 distance: expected 1000.0, got %v", distance)
		}
		if isRx := m1["is_rx"].(bool); !isRx {
			t.Errorf("Movement 1 is_rx: expected true")
		}

		// Verify WODs
		wods := template["wods"].([]interface{})
		if len(wods) != 2 {
			t.Fatalf("Expected 2 WODs, got %d", len(wods))
		}

		w1 := wods[0].(map[string]interface{})
		if scoreValue := w1["score_value"].(string); scoreValue != "4:30" {
			t.Errorf("WOD 1 score_value: expected '4:30', got %s", scoreValue)
		}
		if division := w1["division"].(string); division != "rx" {
			t.Errorf("WOD 1 division: expected 'rx', got %s", division)
		}
		if isPR := w1["is_pr"].(bool); !isPR {
			t.Errorf("WOD 1 is_pr: expected true")
		}
	})

	// UPDATE all attributes
	t.Run("Update_All_Attributes", func(t *testing.T) {
		if templateID == 0 {
			t.Skip("No template ID")
		}

		body := map[string]interface{}{
			"name":  "Updated Complete Template",
			"notes": "Updated notes",
			"movements": []map[string]interface{}{
				{
					"movement_id":  1,
					"weight":       245.0,  // Updated
					"sets":         3,      // Updated
					"reps":         3,      // Updated
					"time":         90,     // Updated
					"distance":     500.0,  // Updated
					"is_rx":        false,  // Updated
					"is_pr":        true,   // Updated
					"instructions": "Updated instructions",
					"notes":        "Updated movement notes",
				},
			},
			"wods": []map[string]interface{}{
				{
					"wod_id":       1,
					"score_value":  "4:15",   // Updated (faster)
					"division":     "scaled", // Updated
					"is_pr":        false,    // Updated
					"instructions": "Updated WOD instructions",
					"notes":        "Updated WOD notes",
				},
			},
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("PUT", "/api/templates/"+formatID(templateID), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify updates
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		var response map[string]interface{}
		json.Unmarshal(getRec.Body.Bytes(), &response)

		template := response["template"].(map[string]interface{})

		if name := template["name"].(string); name != "Updated Complete Template" {
			t.Errorf("Updated name mismatch: %s", name)
		}

		movements := template["movements"].([]interface{})
		m1 := movements[0].(map[string]interface{})
		if weight := m1["weight"].(float64); weight != 245.0 {
			t.Errorf("Updated weight: expected 245.0, got %v", weight)
		}
		if isPR := m1["is_pr"].(bool); !isPR {
			t.Errorf("Updated is_pr: expected true")
		}

		wods := template["wods"].([]interface{})
		w1 := wods[0].(map[string]interface{})
		if scoreValue := w1["score_value"].(string); scoreValue != "4:15" {
			t.Errorf("Updated score_value: expected '4:15', got %s", scoreValue)
		}
		if division := w1["division"].(string); division != "scaled" {
			t.Errorf("Updated division: expected 'scaled', got %s", division)
		}
	})

	// DELETE
	t.Run("Delete_Template", func(t *testing.T) {
		if templateID == 0 {
			t.Skip("No template ID")
		}

		req := httptest.NewRequest("DELETE", "/api/templates/"+formatID(templateID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		// Should return 200 or 204
		if rec.Code != http.StatusOK && rec.Code != http.StatusNoContent {
			t.Fatalf("Expected 200 or 204, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify deleted
		getReq := httptest.NewRequest("GET", "/api/templates/"+formatID(templateID), nil)
		getRec := httptest.NewRecorder()
		router.ServeHTTP(getRec, getReq)

		if getRec.Code != http.StatusNotFound {
			t.Errorf("Expected 404 after delete, got %d", getRec.Code)
		}
	})
}
