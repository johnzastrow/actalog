package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/johnzastrow/actalog/internal/domain"
)

// WodifyResultParser handles parsing of Wodify performance result strings
type WodifyResultParser struct{}

// NewWodifyResultParser creates a new result parser
func NewWodifyResultParser() *WodifyResultParser {
	return &WodifyResultParser{}
}

// ParseResult parses a Wodify result string based on the result type
func (p *WodifyResultParser) ParseResult(resultType, resultString, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{
		Notes: comment,
	}

	// Clean the result string
	resultString = strings.TrimSpace(resultString)

	switch resultType {
	case "Weight":
		return p.parseWeight(resultString, comment)
	case "Time":
		return p.parseTime(resultString, comment)
	case "AMRAP - Rounds and Reps":
		return p.parseAMRAPRoundsReps(resultString, comment)
	case "AMRAP - Reps":
		return p.parseAMRAPReps(resultString, comment)
	case "AMRAP - Rounds":
		return p.parseAMRAPRounds(resultString, comment)
	case "Max reps":
		return p.parseMaxReps(resultString, comment)
	case "Calories":
		return p.parseCalories(resultString, comment)
	case "Distance":
		return p.parseDistance(resultString, comment)
	case "Each Round":
		return p.parseEachRound(resultString, comment)
	default:
		// Unknown type, just store as notes
		result.Notes = fmt.Sprintf("%s: %s. %s", resultType, resultString, comment)
		return result, nil
	}
}

// parseWeight parses weight results like "3 x 10 @ 85 lbs" or "1 x 5 @ 135 lbs"
func (p *WodifyResultParser) parseWeight(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X x Y @ Z lbs" or "X x Y @ Z#"
	re := regexp.MustCompile(`(\d+)\s*x\s*(\d+)\s*@\s*(\d+(?:\.\d+)?)\s*(?:lbs?|#)`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 4 {
		sets, _ := strconv.Atoi(matches[1])
		reps, _ := strconv.Atoi(matches[2])
		weight, _ := strconv.ParseFloat(matches[3], 64)
		result.Sets = &sets
		result.Reps = &reps
		result.Weight = &weight
		return result, nil
	}

	// Pattern: "X lbs" or "Y#" (weight only)
	re2 := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:lbs?|#)`)
	matches2 := re2.FindStringSubmatch(s)
	if len(matches2) == 2 {
		weight, _ := strconv.ParseFloat(matches2[1], 64)
		result.Weight = &weight
		return result, nil
	}

	// Check comment for weight (e.g., "75#", "@75#")
	if comment != "" {
		re3 := regexp.MustCompile(`[@']?(\d+)\s*#`)
		matches3 := re3.FindStringSubmatch(comment)
		if len(matches3) == 2 {
			weight, _ := strconv.ParseFloat(matches3[1], 64)
			result.Weight = &weight
		}
	}

	return result, nil
}

// parseTime parses time results like "5:30", "19:05", "43:37"
func (p *WodifyResultParser) parseTime(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "MM:SS" or "HH:MM:SS"
	parts := strings.Split(s, ":")

	var totalSeconds int
	switch len(parts) {
	case 2: // MM:SS
		minutes, _ := strconv.Atoi(parts[0])
		seconds, _ := strconv.Atoi(parts[1])
		totalSeconds = minutes*60 + seconds
	case 3: // HH:MM:SS
		hours, _ := strconv.Atoi(parts[0])
		minutes, _ := strconv.Atoi(parts[1])
		seconds, _ := strconv.Atoi(parts[2])
		totalSeconds = hours*3600 + minutes*60 + seconds
	default:
		// Try parsing as just seconds
		if sec, err := strconv.Atoi(s); err == nil {
			totalSeconds = sec
		}
	}

	if totalSeconds > 0 {
		result.TimeSeconds = &totalSeconds
	}

	return result, nil
}

// parseAMRAPRoundsReps parses AMRAP results like "7 + 3" (7 rounds + 3 reps)
func (p *WodifyResultParser) parseAMRAPRoundsReps(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X + Y" or "X+Y"
	re := regexp.MustCompile(`(\d+)\s*\+\s*(\d+)`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 3 {
		rounds, _ := strconv.Atoi(matches[1])
		reps, _ := strconv.Atoi(matches[2])
		result.Rounds = &rounds
		result.Reps = &reps
	}

	return result, nil
}

// parseAMRAPReps parses AMRAP reps like "50 Reps"
func (p *WodifyResultParser) parseAMRAPReps(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X Reps" or "X reps"
	re := regexp.MustCompile(`(\d+)\s*[Rr]eps?`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 2 {
		reps, _ := strconv.Atoi(matches[1])
		result.Reps = &reps
	}

	return result, nil
}

// parseAMRAPRounds parses AMRAP rounds like "5 Rounds"
func (p *WodifyResultParser) parseAMRAPRounds(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X Rounds"
	re := regexp.MustCompile(`(\d+)\s*[Rr]ounds?`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 2 {
		rounds, _ := strconv.Atoi(matches[1])
		result.Rounds = &rounds
	}

	return result, nil
}

// parseMaxReps parses max reps like "3 x 8"
func (p *WodifyResultParser) parseMaxReps(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X x Y" (sets x reps)
	re := regexp.MustCompile(`(\d+)\s*x\s*(\d+)`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 3 {
		sets, _ := strconv.Atoi(matches[1])
		reps, _ := strconv.Atoi(matches[2])
		result.Sets = &sets
		result.Reps = &reps
	}

	return result, nil
}

// parseCalories parses calorie results like "133 Calories"
func (p *WodifyResultParser) parseCalories(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X Calories" or "X Cal"
	re := regexp.MustCompile(`(\d+)\s*(?:[Cc]alories?|[Cc]al)`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 2 {
		calories, _ := strconv.Atoi(matches[1])
		result.Calories = &calories
	}

	return result, nil
}

// parseDistance parses distance results
func (p *WodifyResultParser) parseDistance(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X m" or "X meters"
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:m|meters?)`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 2 {
		distance, _ := strconv.ParseFloat(matches[1], 64)
		result.Distance = &distance
	}

	return result, nil
}

// parseEachRound parses "Each Round" results like "175 Total Reps"
func (p *WodifyResultParser) parseEachRound(s, comment string) (*domain.ParsedPerformanceResult, error) {
	result := &domain.ParsedPerformanceResult{Notes: comment}

	// Pattern: "X Total Reps"
	re := regexp.MustCompile(`(\d+)\s*[Tt]otal\s*[Rr]eps?`)
	matches := re.FindStringSubmatch(s)

	if len(matches) == 2 {
		reps, _ := strconv.Atoi(matches[1])
		result.Reps = &reps
	}

	return result, nil
}

// ParseDate parses a Wodify date string (MM/DD/YYYY) to time.Time
func (p *WodifyResultParser) ParseDate(dateStr string) (time.Time, error) {
	// Try MM/DD/YYYY format
	t, err := time.Parse("01/02/2006", dateStr)
	if err == nil {
		return t, nil
	}

	// Try MM/DD/YY format
	t, err = time.Parse("01/02/06", dateStr)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}

// DetermineMovementType determines the movement type based on component type
func (p *WodifyResultParser) DetermineMovementType(componentType string) string {
	switch strings.ToLower(componentType) {
	case "weightlifting":
		return "weightlifting"
	case "gymnastics":
		return "gymnastics"
	case "cardio":
		return "cardio"
	default:
		return "bodyweight"
	}
}

// DetermineWODScoreType determines the WOD score type based on result type
func (p *WodifyResultParser) DetermineWODScoreType(resultType string) string {
	switch resultType {
	case "Time":
		return "Time (HH:MM:SS)"
	case "AMRAP - Rounds and Reps", "AMRAP - Rounds", "AMRAP - Reps":
		return "Rounds+Reps"
	case "Weight":
		return "Max Weight"
	default:
		return "Time (HH:MM:SS)" // Default
	}
}
