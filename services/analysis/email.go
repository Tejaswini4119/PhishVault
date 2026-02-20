package analysis

import (
	"strings"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

// AnalyzeEmail performs forensic analysis on email metadata and content.
func AnalyzeEmail(task domain.SAL) domain.Signal {
	signal := domain.Signal{
		EngineName: "EmailForensics",
		SignalKey:  "EMAIL_RISK",
		Evidence:   make(map[string]interface{}),
	}

	var riskScore float64
	var reasons []string

	// 1. SPF Check
	if spf, ok := task.Metadata["auth_spf_verdict"].(string); ok {
		signal.Evidence["spf_verdict"] = spf
		if spf == "fail" {
			riskScore += 0.4
			reasons = append(reasons, "SPF Authentication Failed")
		} else if spf == "softfail" {
			riskScore += 0.2
			reasons = append(reasons, "SPF Softfail")
		}
	}

	// 2. DKIM Check
	if dkim, ok := task.Metadata["auth_dkim_independent"].(string); ok {
		signal.Evidence["dkim_verdict"] = dkim
		if dkim == "invalid" || dkim == "fail" {
			riskScore += 0.3
			reasons = append(reasons, "DKIM Signature Invalid")
		}
	}

	// 3. Display Name Spoofing
	from := task.Metadata["From"].(string)
	if from != "" {
		// Check for common phishing pattern: "Display Name <actual@email.com>"
		// If Display Name contains another email address or suspicious keywords
		if strings.Contains(from, "@") && strings.Count(from, "@") > 1 {
			riskScore += 0.5
			reasons = append(reasons, "Possible Display Name Spoofing (Multiple @ symbols)")
		}
	}

	// 4. Urgency Keywords
	subject := task.Metadata["Subject"].(string)
	urgencyKeywords := []string{"urgent", "action required", "suspended", "security alert", "unauthorized"}
	for _, kw := range urgencyKeywords {
		if strings.Contains(strings.ToLower(subject), kw) {
			riskScore += 0.1
			reasons = append(reasons, "Urgent/Threatening Subject Line")
			break
		}
	}

	// Cap risk score at 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	signal.Confidence = 0.9 // High confidence in these deterministic checks
	signal.Weight = riskScore
	signal.Evidence["reasons"] = reasons

	return signal
}
