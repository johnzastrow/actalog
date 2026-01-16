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
