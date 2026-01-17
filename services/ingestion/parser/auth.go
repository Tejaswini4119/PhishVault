package parser

import (
	"strings"
)

// AuthResult holds the parsed authentication status
type AuthResult struct {
	SPF   string `json:"spf"`
	DKIM  string `json:"dkim"`
	DMARC string `json:"dmarc"`
}

// AnalyzeAuthHeaders parses the 'Authentication-Results' or separate headers to determine trust.
// It prioritizes the standard 'Authentication-Results' header used by major providers (Gmail, O365).
func AnalyzeAuthHeaders(headers map[string]string) AuthResult {
	res := AuthResult{
		SPF:   "NONE",
		DKIM:  "NONE",
		DMARC: "NONE",
	}

	// 1. Check RFC 8601 Authentication-Results
	if val, ok := headers["Authentication-Results"]; ok {
		parseAuthResults(val, &res)
	}

	// 2. Fallback to individual headers if "Authentication-Results" didn't cover it or wasn't present
	// (e.g. Received-SPF)
	if res.SPF == "NONE" || res.SPF == "UNKNOWN" {
		if val, ok := headers["Received-SPF"]; ok {
			res.SPF = parseSPFSimple(val)
		}
	}

	// Note: Explicit logic for DMARC relies heavily on Auth-Res header usually.

	return res
}

func parseAuthResults(header string, res *AuthResult) {
	// Example: mx.google.com; dkim=pass header.i=@gmail.com; spf=pass (google.com: domain...)
	parts := strings.Fields(strings.ReplaceAll(header, ";", " "))
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.ToLower(kv[0])
		val := strings.ToLower(kv[1])

		switch key {
		case "spf":
			res.SPF = normalizeVerdict(val)
		case "dkim":
			res.DKIM = normalizeVerdict(val)
		case "dmarc":
			res.DMARC = normalizeVerdict(val)
		}
	}
}

func parseSPFSimple(header string) string {
	lower := strings.ToLower(header)
	if strings.HasPrefix(lower, "pass") {
		return "PASS"
	}
	if strings.HasPrefix(lower, "fail") || strings.HasPrefix(lower, "softfail") {
		return "FAIL"
	}
	if strings.Contains(lower, " neutral ") {
		return "NEUTRAL"
	}
	return "UNKNOWN"
}

func normalizeVerdict(v string) string {
	if v == "pass" {
		return "PASS"
	}
	if v == "fail" || v == "hardfail" || v == "softfail" {
		return "FAIL"
	} // Treat softfail as fail for simplicity
	if v == "neutral" || v == "none" {
		return "NEUTRAL"
	}
	return "UNKNOWN"
}
