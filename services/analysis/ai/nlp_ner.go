package ai

import (
	"fmt"
	"strings"

	"github.com/jdkato/prose/v2"
)

// ExtractBrands uses Named Entity Recognition (NER) to find organization names in text.
func ExtractBrands(text string) ([]string, error) {
	// Create a new document with the default configuration:
	// English model, segmentation, tokenization, POS tagging, and NER.
	doc, err := prose.NewDocument(text)
	if err != nil {
		return nil, err
	}

	brands := []string{}
	seen := make(map[string]bool)

	// Iterate over the entities
	for _, ent := range doc.Entities() {
		// "GPE" = Location, "PERSON" = Person, "ORGANIZATION" = Brand/Org
		// Prose v2 uses label "GPE" / "ORGANIZATION" etc.
		if ent.Label == "ORGANIZATION" || ent.Label == "ORG" {
			cleanName := strings.TrimSpace(ent.Text)
			if !seen[cleanName] {
				brands = append(brands, cleanName)
				seen[cleanName] = true
			}
		}
	}

	return brands, nil
}

// CheckBrandMismatch compares detected brands against the hosting domain.
// e.g., Finds "PayPal" in text but domain is "verify-secure.com" -> HIGH RISK.
func CheckBrandMismatch(brands []string, domain string) bool {
	if len(brands) == 0 {
		return false
	}

	// Whitelist / Approximate Match
	// In production, this checks a robust DB of Brand -> Official Domains.
	// For MVP Industrial, simple string containment.

	domainLower := strings.ToLower(domain)

	for _, brand := range brands {
		brandLower := strings.ToLower(brand)
		// If brand is "Google" and domain contains "google", it's likely fine.
		if strings.Contains(domainLower, brandLower) {
			continue // Match found
		}

		// Common False Positives ignored
		if brandLower == "inc" || brandLower == "ltd" || brandLower == "browser" {
			continue
		}

		// Found a brand that is NOT in the domain
		// This is a mismatch signal (Potential Phishing)
		fmt.Printf("Brand Mismatch Detected: %s in %s\n", brand, domain)
		return true
	}

	return false
}
