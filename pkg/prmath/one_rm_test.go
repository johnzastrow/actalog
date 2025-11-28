package prmath

import (
	"math"
	"testing"
)

func TestCalculate1RM(t *testing.T) {
	tests := []struct {
		name            string
		weight          float64
		reps            int
		expectedMin     float64
		expectedMax     float64
		expectedFormula Formula
	}{
		{
			name:            "1 rep - actual weight",
			weight:          225.0,
			reps:            1,
			expectedMin:     225.0,
			expectedMax:     225.0,
			expectedFormula: FormulaActual,
		},
		{
			name:            "5 reps - Epley formula",
			weight:          200.0,
			reps:            5,
			expectedMin:     230.0,
			expectedMax:     235.0,
			expectedFormula: FormulaEpley,
		},
		{
			name:            "10 reps - Epley formula",
			weight:          185.0,
			reps:            10,
			expectedMin:     245.0,
			expectedMax:     250.0,
			expectedFormula: FormulaEpley,
		},
		{
			name:            "15 reps - Wathan formula",
			weight:          135.0,
			reps:            15,
			expectedMin:     200.0,
			expectedMax:     210.0,
			expectedFormula: FormulaWathan,
		},
		{
			name:            "20 reps - Wathan formula",
			weight:          115.0,
			reps:            20,
			expectedMin:     180.0,
			expectedMax:     195.0,
			expectedFormula: FormulaWathan,
		},
		{
			name:            "zero weight - invalid",
			weight:          0,
			reps:            5,
			expectedMin:     0,
			expectedMax:     0,
			expectedFormula: "",
		},
		{
			name:            "zero reps - invalid",
			weight:          200.0,
			reps:            0,
			expectedMin:     0,
			expectedMax:     0,
			expectedFormula: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oneRM, formula := Calculate1RM(tt.weight, tt.reps)

			// Check formula is correct
			if formula != tt.expectedFormula {
				t.Errorf("Calculate1RM(%v, %v) formula = %v, want %v",
					tt.weight, tt.reps, formula, tt.expectedFormula)
			}

			// Check 1RM is within expected range
			if oneRM < tt.expectedMin || oneRM > tt.expectedMax {
				t.Errorf("Calculate1RM(%v, %v) = %v, want between %v and %v",
					tt.weight, tt.reps, oneRM, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestCalculateEpley(t *testing.T) {
	tests := []struct {
		weight   float64
		reps     int
		expected float64
	}{
		{200.0, 5, 233.33},  // 200 * (1 + 5/30) = 233.33
		{100.0, 10, 133.33}, // 100 * (1 + 10/30) = 133.33
	}

	for _, tt := range tests {
		result := calculateEpley(tt.weight, tt.reps)
		if math.Abs(result-tt.expected) > 0.1 {
			t.Errorf("calculateEpley(%v, %v) = %v, want ~%v",
				tt.weight, tt.reps, result, tt.expected)
		}
	}
}

func TestCalculateWathan(t *testing.T) {
	tests := []struct {
		weight      float64
		reps        int
		expectedMin float64
		expectedMax float64
	}{
		{135.0, 15, 200.0, 210.0},
		{115.0, 20, 180.0, 195.0},
	}

	for _, tt := range tests {
		result := calculateWathan(tt.weight, tt.reps)
		if result < tt.expectedMin || result > tt.expectedMax {
			t.Errorf("calculateWathan(%v, %v) = %v, want between %v and %v",
				tt.weight, tt.reps, result, tt.expectedMin, tt.expectedMax)
		}
	}
}

func TestCalculateAllFormulas(t *testing.T) {
	weight := 200.0
	reps := 5

	results := CalculateAllFormulas(weight, reps)

	// Should have multiple formulas
	if len(results) < 4 {
		t.Errorf("CalculateAllFormulas returned %d formulas, expected at least 4", len(results))
	}

	// Check required formulas exist
	if _, ok := results["Epley"]; !ok {
		t.Error("CalculateAllFormulas missing Epley formula")
	}
	if _, ok := results["Brzycki"]; !ok {
		t.Error("CalculateAllFormulas missing Brzycki formula")
	}
	if _, ok := results["Wathan"]; !ok {
		t.Error("CalculateAllFormulas missing Wathan formula")
	}
}

func TestCompareToBaseline(t *testing.T) {
	tests := []struct {
		name        string
		current1RM  float64
		baseline1RM float64
		expectedPct float64
	}{
		{
			name:        "10% improvement",
			current1RM:  220.0,
			baseline1RM: 200.0,
			expectedPct: 10.0,
		},
		{
			name:        "no change",
			current1RM:  200.0,
			baseline1RM: 200.0,
			expectedPct: 0.0,
		},
		{
			name:        "5% decrease",
			current1RM:  190.0,
			baseline1RM: 200.0,
			expectedPct: -5.0,
		},
		{
			name:        "zero baseline - invalid",
			current1RM:  220.0,
			baseline1RM: 0,
			expectedPct: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareToBaseline(tt.current1RM, tt.baseline1RM)
			if math.Abs(result-tt.expectedPct) > 0.01 {
				t.Errorf("CompareToBaseline(%v, %v) = %v, want %v",
					tt.current1RM, tt.baseline1RM, result, tt.expectedPct)
			}
		})
	}
}

func TestCalculateIntensity(t *testing.T) {
	tests := []struct {
		name        string
		weight      float64
		oneRM       float64
		expectedPct float64
	}{
		{
			name:        "80% intensity",
			weight:      160.0,
			oneRM:       200.0,
			expectedPct: 80.0,
		},
		{
			name:        "100% intensity",
			weight:      200.0,
			oneRM:       200.0,
			expectedPct: 100.0,
		},
		{
			name:        "50% intensity",
			weight:      100.0,
			oneRM:       200.0,
			expectedPct: 50.0,
		},
		{
			name:        "zero oneRM - invalid",
			weight:      160.0,
			oneRM:       0,
			expectedPct: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateIntensity(tt.weight, tt.oneRM)
			if math.Abs(result-tt.expectedPct) > 0.01 {
				t.Errorf("CalculateIntensity(%v, %v) = %v, want %v",
					tt.weight, tt.oneRM, result, tt.expectedPct)
			}
		})
	}
}
