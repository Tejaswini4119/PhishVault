package reporting

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

func TestGenerateReport_JSON(t *testing.T) {
	rg := NewReportGenerator()

	sal := domain.SAL{
		ScanID:    "test-scan-1",
		URL:       "http://example.com",
		Timestamp: time.Now(),
		Verdict:   "MALICIOUS",
		RiskScore: 0.95,
	}

	data, err := rg.Generate(sal, "json")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	var result domain.SAL
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON report: %v", err)
	}

	if result.ScanID != sal.ScanID {
		t.Errorf("Expected ScanID %s, got %s", sal.ScanID, result.ScanID)
	}
}

func TestGenerateReport_Text(t *testing.T) {
	rg := NewReportGenerator()

	sal := domain.SAL{
		ScanID:    "test-scan-2",
		URL:       "http://evil.com",
		Timestamp: time.Now(),
		Verdict:   "SAFE",
		RiskScore: 0.1,
	}

	data, err := rg.Generate(sal, "text")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := string(data)
	if !strings.Contains(output, "PHISHVAULT SCAN REPORT") {
		t.Error("Report header missing")
	}
	if !strings.Contains(output, "Target:    http://evil.com") {
		t.Error("Target URL missing in report")
	}
}
