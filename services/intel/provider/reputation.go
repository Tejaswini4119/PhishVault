package provider

// ReputationResult holds the score from threat feeds.
type ReputationResult struct {
	Source    string
	Malicious bool
	Score     float64 // 0.0 to 1.0 (1.0 = known malware)
}

// CheckReputation simulates checking against threat feeds like VirusTotal or OTX.
func CheckReputation(domain string) ([]ReputationResult, error) {
	// Mock MVP Data
	// In reality, this would query APIs as configured
	results := []ReputationResult{}

	// Simulate hit for specific domains
	if domain == "malware.example.com" || domain == "login-apple-secure.com" {
		results = append(results, ReputationResult{
			Source:    "VirusTotal (Mock)",
			Malicious: true,
			Score:     0.95,
		})
	} else {
		results = append(results, ReputationResult{
			Source:    "VirusTotal (Mock)",
			Malicious: false,
			Score:     0.0,
		})
	}

	return results, nil
}
