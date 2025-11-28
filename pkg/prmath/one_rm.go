package prmath

import "math"

// Formula represents which 1RM calculation formula was used
type Formula string

const (
	FormulaActual Formula = "Actual 1RM"        // For 1 rep max
	FormulaEpley  Formula = "Epley (2-10 reps)" // For moderate reps
	FormulaWathan Formula = "Wathan (11+ reps)" // For higher reps
)

// Calculate1RM calculates the one-rep max using the hybrid approach
// - For 1 rep: returns actual weight
// - For 2-10 reps: uses Epley formula
// - For 11+ reps: uses Wathan formula
func Calculate1RM(weight float64, reps int) (oneRM float64, formula Formula) {
	if weight <= 0 || reps <= 0 {
		return 0, ""
	}

	switch {
	case reps == 1:
		return weight, FormulaActual
	case reps <= 10:
		return calculateEpley(weight, reps), FormulaEpley
	default:
		return calculateWathan(weight, reps), FormulaWathan
	}
}

// calculateEpley uses the Epley formula: 1RM = weight × (1 + reps/30)
// Best for 2-10 reps, widely used and accepted
func calculateEpley(weight float64, reps int) float64 {
	return weight * (1 + float64(reps)/30.0)
}

// calculateWathan uses the Wathan formula: 1RM = (100 × weight) / (48.8 + 53.8 × e^(-0.075 × reps))
// Best for 11+ reps, more accurate for higher rep ranges
func calculateWathan(weight float64, reps int) float64 {
	exponent := -0.075 * float64(reps)
	denominator := 48.8 + 53.8*math.Exp(exponent)
	return (100 * weight) / denominator
}

// CalculateAllFormulas calculates 1RM using all common formulas for comparison
// Useful for showing users different estimates
func CalculateAllFormulas(weight float64, reps int) map[string]float64 {
	if weight <= 0 || reps <= 0 {
		return map[string]float64{}
	}

	results := make(map[string]float64)

	// Actual (for 1 rep)
	if reps == 1 {
		results["Actual"] = weight
	}

	// Epley: 1RM = weight × (1 + reps/30)
	results["Epley"] = calculateEpley(weight, reps)

	// Brzycki: 1RM = weight × (36 / (37 - reps))
	if reps < 37 {
		results["Brzycki"] = weight * (36.0 / (37.0 - float64(reps)))
	}

	// Lombardi: 1RM = weight × reps^0.10
	results["Lombardi"] = weight * math.Pow(float64(reps), 0.10)

	// Mayhew: 1RM = (100 × weight) / (52.2 + 41.9 × e^(-0.055 × reps))
	mayhewExp := -0.055 * float64(reps)
	mayhewDenom := 52.2 + 41.9*math.Exp(mayhewExp)
	results["Mayhew"] = (100 * weight) / mayhewDenom

	// Wathan: 1RM = (100 × weight) / (48.8 + 53.8 × e^(-0.075 × reps))
	results["Wathan"] = calculateWathan(weight, reps)

	// O'Conner: 1RM = weight × (1 + reps/40)
	results["O'Conner"] = weight * (1 + float64(reps)/40.0)

	return results
}

// CompareToBaseline calculates percentage improvement over a baseline 1RM
func CompareToBaseline(current1RM, baseline1RM float64) float64 {
	if baseline1RM <= 0 {
		return 0
	}
	return ((current1RM - baseline1RM) / baseline1RM) * 100
}

// CalculateIntensity calculates the percentage of 1RM being used
func CalculateIntensity(weight, oneRM float64) float64 {
	if oneRM <= 0 {
		return 0
	}
	return (weight / oneRM) * 100
}
