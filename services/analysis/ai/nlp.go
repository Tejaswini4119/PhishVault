package ai

import (
	"regexp"
	"strings"
)

// AnalyzeTextResult holds the output of text analysis.
type AnalyzeTextResult struct {
	UrgencyScore float64  // 0.0 to 1.0 (1.0 = highly urgent)
	Keywords     []string // Detected suspicious keywords
}

// AnalyzeText scans the input text for indicators of phishing.
func AnalyzeText(text string) AnalyzeTextResult {
	lowerText := strings.ToLower(text)
	result := AnalyzeTextResult{
		Keywords: []string{},
	}

	// 1. Keyword Extraction
	// Common phishing keywords related to credentials, banking, or urgency.
	suspiciousKeywords := []string{
		"verify your account",
		"verify your identity",
		"update payment",
		"suspended",
		"unusual activity",
		"confirmation required",
		"expires in 24",
		"immediately",
		"action required",
	}

	foundCount := 0
	for _, kw := range suspiciousKeywords {
		if strings.Contains(lowerText, kw) {
			result.Keywords = append(result.Keywords, kw)
			foundCount++
		}
	}

	// 2. Urgency Detection (Regex)
	// Look for time-bound threats.
	urgencyPatterns := []*regexp.Regexp{
		regexp.MustCompile(`expire[s|d]?\s+in\s+\d+\s+hour`),
		regexp.MustCompile(`action\s+required\s+immediately`),
		regexp.MustCompile(`account\s+will\s+be\s+locked`),
		regexp.MustCompile(`suspended\s+within\s+24`),
	}

	urgencyMatches := 0
	for _, p := range urgencyPatterns {
		if p.MatchString(lowerText) {
			urgencyMatches++
		}
	}

	// Calculate Score
	// Simple heuristic: 0.2 per keyword, 0.4 per urgency match, capped at 1.0
	score := (float64(foundCount) * 0.2) + (float64(urgencyMatches) * 0.4)
	if score > 1.0 {
		score = 1.0
	}
	result.UrgencyScore = score

	return result
}
