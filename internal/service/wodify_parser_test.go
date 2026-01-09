package service

import (
	"testing"
	"time"
)

func TestNewWodifyResultParser(t *testing.T) {
	parser := NewWodifyResultParser()
	if parser == nil {
		t.Fatal("NewWodifyResultParser returned nil")
	}
}

func TestWodifyResultParser_ParseResult_Weight(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name         string
		resultString string
		comment      string
		wantSets     *int
		wantReps     *int
		wantWeight   *float64
	}{
		{
			name:         "full format with lbs",
			resultString: "3 x 10 @ 85 lbs",
			comment:      "",
			wantSets:     intPtr(3),
			wantReps:     intPtr(10),
			wantWeight:   floatPtr(85),
		},
		{
			name:         "full format with #",
			resultString: "1 x 5 @ 135#",
			comment:      "",
			wantSets:     intPtr(1),
			wantReps:     intPtr(5),
			wantWeight:   floatPtr(135),
		},
		{
			name:         "weight only",
			resultString: "225 lbs",
			comment:      "",
			wantSets:     nil,
			wantReps:     nil,
			wantWeight:   floatPtr(225),
		},
		{
			name:         "weight from comment",
			resultString: "something else",
			comment:      "@75#",
			wantSets:     nil,
			wantReps:     nil,
			wantWeight:   floatPtr(75),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseResult("Weight", tt.resultString, tt.comment)
			if err != nil {
				t.Fatalf("ParseResult returned error: %v", err)
			}

			if !equalIntPtr(result.Sets, tt.wantSets) {
				t.Errorf("Sets = %v, want %v", ptrIntVal(result.Sets), ptrIntVal(tt.wantSets))
			}
			if !equalIntPtr(result.Reps, tt.wantReps) {
				t.Errorf("Reps = %v, want %v", ptrIntVal(result.Reps), ptrIntVal(tt.wantReps))
			}
			if !equalFloatPtr(result.Weight, tt.wantWeight) {
				t.Errorf("Weight = %v, want %v", ptrFloatVal(result.Weight), ptrFloatVal(tt.wantWeight))
			}
		})
	}
}

func TestWodifyResultParser_ParseResult_Time(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name         string
		resultString string
		wantSeconds  *int
	}{
		{
			name:         "minutes and seconds",
			resultString: "5:30",
			wantSeconds:  intPtr(330),
		},
		{
			name:         "longer time",
			resultString: "19:05",
			wantSeconds:  intPtr(1145),
		},
		{
			name:         "hours minutes seconds",
			resultString: "1:05:30",
			wantSeconds:  intPtr(3930),
		},
		{
			name:         "just seconds",
			resultString: "90",
			wantSeconds:  intPtr(90),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseResult("Time", tt.resultString, "")
			if err != nil {
				t.Fatalf("ParseResult returned error: %v", err)
			}

			if !equalIntPtr(result.TimeSeconds, tt.wantSeconds) {
				t.Errorf("TimeSeconds = %v, want %v", ptrIntVal(result.TimeSeconds), ptrIntVal(tt.wantSeconds))
			}
		})
	}
}

func TestWodifyResultParser_ParseResult_AMRAPRoundsReps(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name         string
		resultString string
		wantRounds   *int
		wantReps     *int
	}{
		{
			name:         "with spaces",
			resultString: "7 + 3",
			wantRounds:   intPtr(7),
			wantReps:     intPtr(3),
		},
		{
			name:         "without spaces",
			resultString: "5+12",
			wantRounds:   intPtr(5),
			wantReps:     intPtr(12),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseResult("AMRAP - Rounds and Reps", tt.resultString, "")
			if err != nil {
				t.Fatalf("ParseResult returned error: %v", err)
			}

			if !equalIntPtr(result.Rounds, tt.wantRounds) {
				t.Errorf("Rounds = %v, want %v", ptrIntVal(result.Rounds), ptrIntVal(tt.wantRounds))
			}
			if !equalIntPtr(result.Reps, tt.wantReps) {
				t.Errorf("Reps = %v, want %v", ptrIntVal(result.Reps), ptrIntVal(tt.wantReps))
			}
		})
	}
}

func TestWodifyResultParser_ParseResult_AMRAPReps(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name         string
		resultString string
		wantReps     *int
	}{
		{
			name:         "with Reps",
			resultString: "50 Reps",
			wantReps:     intPtr(50),
		},
		{
			name:         "lowercase reps",
			resultString: "25 reps",
			wantReps:     intPtr(25),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseResult("AMRAP - Reps", tt.resultString, "")
			if err != nil {
				t.Fatalf("ParseResult returned error: %v", err)
			}

			if !equalIntPtr(result.Reps, tt.wantReps) {
				t.Errorf("Reps = %v, want %v", ptrIntVal(result.Reps), ptrIntVal(tt.wantReps))
			}
		})
	}
}

func TestWodifyResultParser_ParseResult_AMRAPRounds(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name         string
		resultString string
		wantRounds   *int
	}{
		{
			name:         "with Rounds",
			resultString: "5 Rounds",
			wantRounds:   intPtr(5),
		},
		{
			name:         "lowercase rounds",
			resultString: "10 rounds",
			wantRounds:   intPtr(10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseResult("AMRAP - Rounds", tt.resultString, "")
			if err != nil {
				t.Fatalf("ParseResult returned error: %v", err)
			}

			if !equalIntPtr(result.Rounds, tt.wantRounds) {
				t.Errorf("Rounds = %v, want %v", ptrIntVal(result.Rounds), ptrIntVal(tt.wantRounds))
			}
		})
	}
}

func TestWodifyResultParser_ParseResult_MaxReps(t *testing.T) {
	parser := NewWodifyResultParser()

	result, err := parser.ParseResult("Max reps", "3 x 8", "")
	if err != nil {
		t.Fatalf("ParseResult returned error: %v", err)
	}

	if !equalIntPtr(result.Sets, intPtr(3)) {
		t.Errorf("Sets = %v, want 3", ptrIntVal(result.Sets))
	}
	if !equalIntPtr(result.Reps, intPtr(8)) {
		t.Errorf("Reps = %v, want 8", ptrIntVal(result.Reps))
	}
}

func TestWodifyResultParser_ParseResult_Calories(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name         string
		resultString string
		wantCalories *int
	}{
		{
			name:         "with Calories",
			resultString: "133 Calories",
			wantCalories: intPtr(133),
		},
		{
			name:         "with Cal",
			resultString: "100 Cal",
			wantCalories: intPtr(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseResult("Calories", tt.resultString, "")
			if err != nil {
				t.Fatalf("ParseResult returned error: %v", err)
			}

			if !equalIntPtr(result.Calories, tt.wantCalories) {
				t.Errorf("Calories = %v, want %v", ptrIntVal(result.Calories), ptrIntVal(tt.wantCalories))
			}
		})
	}
}

func TestWodifyResultParser_ParseResult_Distance(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name         string
		resultString string
		wantDistance *float64
	}{
		{
			name:         "with m",
			resultString: "500 m",
			wantDistance: floatPtr(500),
		},
		{
			name:         "with meters",
			resultString: "1000 meters",
			wantDistance: floatPtr(1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseResult("Distance", tt.resultString, "")
			if err != nil {
				t.Fatalf("ParseResult returned error: %v", err)
			}

			if !equalFloatPtr(result.Distance, tt.wantDistance) {
				t.Errorf("Distance = %v, want %v", ptrFloatVal(result.Distance), ptrFloatVal(tt.wantDistance))
			}
		})
	}
}

func TestWodifyResultParser_ParseResult_EachRound(t *testing.T) {
	parser := NewWodifyResultParser()

	result, err := parser.ParseResult("Each Round", "175 Total Reps", "")
	if err != nil {
		t.Fatalf("ParseResult returned error: %v", err)
	}

	if !equalIntPtr(result.Reps, intPtr(175)) {
		t.Errorf("Reps = %v, want 175", ptrIntVal(result.Reps))
	}
}

func TestWodifyResultParser_ParseResult_UnknownType(t *testing.T) {
	parser := NewWodifyResultParser()

	result, err := parser.ParseResult("Unknown Type", "some value", "some comment")
	if err != nil {
		t.Fatalf("ParseResult returned error: %v", err)
	}

	// Should store as notes
	expected := "Unknown Type: some value. some comment"
	if result.Notes != expected {
		t.Errorf("Notes = %q, want %q", result.Notes, expected)
	}
}

func TestWodifyResultParser_ParseDate(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		name    string
		dateStr string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "MM/DD/YYYY format",
			dateStr: "12/25/2024",
			want:    time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "MM/DD/YY format",
			dateStr: "01/15/23",
			want:    time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "invalid format",
			dateStr: "2024-12-25",
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parser.ParseDate(tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !got.Equal(tt.want) {
				t.Errorf("ParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWodifyResultParser_DetermineMovementType(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		componentType string
		want          string
	}{
		{"weightlifting", "weightlifting"},
		{"Weightlifting", "weightlifting"},
		{"gymnastics", "gymnastics"},
		{"Gymnastics", "gymnastics"},
		{"cardio", "cardio"},
		{"Cardio", "cardio"},
		{"other", "bodyweight"},
		{"", "bodyweight"},
	}

	for _, tt := range tests {
		t.Run(tt.componentType, func(t *testing.T) {
			got := parser.DetermineMovementType(tt.componentType)
			if got != tt.want {
				t.Errorf("DetermineMovementType(%q) = %q, want %q", tt.componentType, got, tt.want)
			}
		})
	}
}

func TestWodifyResultParser_DetermineWODScoreType(t *testing.T) {
	parser := NewWodifyResultParser()

	tests := []struct {
		resultType string
		want       string
	}{
		{"Time", "Time (HH:MM:SS)"},
		{"AMRAP - Rounds and Reps", "Rounds+Reps"},
		{"AMRAP - Rounds", "Rounds+Reps"},
		{"AMRAP - Reps", "Rounds+Reps"},
		{"Each Round", "Rounds+Reps"},
		{"Max reps", "Rounds+Reps"},
		{"Weight", "Max Weight"},
		{"Unknown", "Time (HH:MM:SS)"},
	}

	for _, tt := range tests {
		t.Run(tt.resultType, func(t *testing.T) {
			got := parser.DetermineWODScoreType(tt.resultType)
			if got != tt.want {
				t.Errorf("DetermineWODScoreType(%q) = %q, want %q", tt.resultType, got, tt.want)
			}
		})
	}
}

// Helper functions for pointer comparison - use those from test_helpers.go
// intPtr, floatPtr are defined in test_helpers.go

func equalIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalFloatPtr(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrIntVal(p *int) string {
	if p == nil {
		return "nil"
	}
	return string(rune(*p) + '0')
}

func ptrFloatVal(p *float64) string {
	if p == nil {
		return "nil"
	}
	return string(rune(int(*p)) + '0')
}
