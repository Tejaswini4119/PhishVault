package ai

import (
	"regexp"
	"strings"
)

// ContentRisk represents the deep analysis of the page content.
type ContentRisk struct {
	UrgencyScore float64  // 0.0 to 1.0 (1.0 = highly urgent)
	Intent       string   // "CredentialHarvesting", "Malware", "Scam", "Benign"
	Keywords     []string // Detected suspicious tokens
	FormRisk     FormAnalysis
}

type FormAnalysis struct {
	HasPassword      bool
	HasUpload        bool
	ForeignAction    bool // True if form posts to a different domain
	ObfuscatedFields bool // True if hidden fields carry suspicious values
}

// AnalyzeContent performs structural and semantic analysis on the provided HTML and Text.
// targetDomain is used to check for Cross-Origin form actions.
func AnalyzeContent(domHTML string, visibleText string, targetDomain string) ContentRisk {
	risk := ContentRisk{
		Keywords: []string{},
		Intent:   "Benign",
	}

	lowerText := strings.ToLower(visibleText)
	lowerHTML := strings.ToLower(domHTML)

	// --- 1. Form & Structure Analysis ---
	risk.FormRisk = analyzeForms(lowerHTML, targetDomain)

	// --- 2. AI Semantic Analysis (Bayesian + NER) ---
	// P(Phishing | Text)
	bayesScore := PredictPhishingProb(lowerText)

	// Brand Extraction (NER)
	brands, _ := ExtractBrands(visibleText)
	brandMismatch := CheckBrandMismatch(brands, targetDomain)
	if brandMismatch {
		risk.Keywords = append(risk.Keywords, "brand_mismatch: "+strings.Join(brands, ","))
	}

	// --- 3. Intent Classification ---
	// Hybrid Rules + Bayesian
	if risk.FormRisk.HasPassword {
		risk.Intent = "CredentialHarvesting"
	} else if bayesScore > 0.8 {
		risk.Intent = "PhishingScam" // High prob generic phishing
	} else {
		risk.Intent = "Benign"
	}

	if strings.Contains(lowerText, "download") && strings.Contains(lowerText, ".exe") {
		risk.Intent = "MalwareDistribuition"
	}

	// --- 4. Scoring ---
	// Weighted Score
	// Bayesian Score (0-1) has high weight
	risk.UrgencyScore = bayesScore * 0.6

	// Boosts
	if brandMismatch {
		risk.UrgencyScore += 0.3
	}
	if risk.FormRisk.ForeignAction {
		risk.UrgencyScore += 0.2
	}

	if risk.UrgencyScore > 1.0 {
		risk.UrgencyScore = 1.0
	}

	return risk
}

func analyzeForms(html string, targetDomain string) FormAnalysis {
	fa := FormAnalysis{}

	// Detect Password Fields
	if strings.Contains(html, "type=\"password\"") || strings.Contains(html, "type='password'") {
		fa.HasPassword = true
	}

	// Detect File Uploads
	if strings.Contains(html, "type=\"file\"") {
		fa.HasUpload = true
	}

	// Detect Foreign Actions (Simplistic for text-based HTML analysis)
	// Ideally we use a tokenizer, but regex helps for MVP Industrial Upgrade
	// Look for <form action="http...">
	actionRegex := regexp.MustCompile(`action=["'](http[s]?://[^"']+)["']`)
	matches := actionRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		actionURL := matches[1]
		// If action URL contains http but NOT the target domain -> Foreign Action
		if !strings.Contains(actionURL, targetDomain) {
			fa.ForeignAction = true
		}
	}

	return fa
}

func countMatches(text string, keywords []string) int {
	count := 0
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			count++
		}
	}
	return count
}
