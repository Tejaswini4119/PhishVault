package parser

import (
	"io"

	"github.com/emersion/go-msgauth/dkim"
)

// VerifyDKIM reads the email from r and verifies DKIM signatures.
// Returns a combined result string (e.g. "PASS", "FAIL", "NONE").
func VerifyDKIM(r io.Reader) (string, error) {
	// dkim.Verify reads the message and returns verifications
	verifications, err := dkim.Verify(r)
	if err != nil {
		return "ERROR: " + err.Error(), err
	}

	if len(verifications) == 0 {
		return "NONE", nil
	}

	for _, v := range verifications {
		if v.Err == nil {
			// If at least one valid signature covers the message, we consider it PASS for now.
			// In strict mode, we might want to check *who* signed it (the domain).
			return "PASS", nil
		}
	}

	// If we have signatures but all failed
	return "FAIL", nil
}

// VerifySPF is a placeholder for future IP-based SPF checks.
// Currently we rely on upstream headers analysis in auth.go
func VerifySPF(ip string, domain string) (string, error) {
	return "SKIPPED", nil
}

// VerifyDMARC is a placeholder for DMARC logic.
func VerifyDMARC(domain string, spfResult, dkimResult string) (string, error) {
	// Detailed DMARC requires DNS TXT lookups for _dmarc.domain
	// and reconciling against SPF/DKIM identifiers.
	// For Stage 1, we return SKIPPED.
	return "SKIPPED", nil
}

// Helper to normalize verification results
func normalizeVerification(v *dkim.Verification) string {
	if v.Err == nil {
		return "PASS"
	}
	return "FAIL"
}
