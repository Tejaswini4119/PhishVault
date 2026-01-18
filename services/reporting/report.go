package reporting

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

// ReportGenerator handles conversion of SAL to export formats.
type ReportGenerator struct{}

// NewReportGenerator creates a new generator.
func NewReportGenerator() *ReportGenerator {
	return &ReportGenerator{}
}

// Generate creates a report in the specified format ("json", "text").
func (rg *ReportGenerator) Generate(sal domain.SAL, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "json":
		return json.MarshalIndent(sal, "", "  ")
	case "text", "pdf": // Mock PDF as text for CLI MVP
		return rg.generateTextReport(sal), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func (rg *ReportGenerator) generateTextReport(sal domain.SAL) []byte {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("PHISHVAULT SCAN REPORT\n"))
	sb.WriteString(fmt.Sprintf("======================\n"))
	sb.WriteString(fmt.Sprintf("Scan ID:   %s\n", sal.ScanID))
	sb.WriteString(fmt.Sprintf("Target:    %s\n", sal.URL))
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", sal.Timestamp.Format(time.RFC1123)))
	sb.WriteString(fmt.Sprintf("Verdict:   %s\n", sal.Verdict))
	sb.WriteString(fmt.Sprintf("RiskScore: %.2f\n", sal.RiskScore))
	sb.WriteString("\n")

	sb.WriteString("SIGNALS DETECTED:\n")
	sb.WriteString("-----------------\n")
	if len(sal.Signals) == 0 {
		sb.WriteString("No significant signals detected.\n")
	} else {
		for _, sig := range sal.Signals {
			sb.WriteString(fmt.Sprintf("- [%s] %s (Conf: %.2f)\n", sig.EngineName, sig.SignalKey, sig.Confidence))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("ENTITIES:\n")
	sb.WriteString("---------\n")
	if len(sal.Entities) == 0 {
		sb.WriteString("No entities extracted.\n")
	} else {
		for _, e := range sal.Entities {
			sb.WriteString(fmt.Sprintf("- %s: %s (%s)\n", e.Type, e.Value, e.Source))
		}
	}

	sb.WriteString("\n")
	sb.WriteString("INTELLIGENCE SUMMARY:\n")
	sb.WriteString("---------------------\n")
	sb.WriteString(fmt.Sprintf("Visual Hash: %s\n", sal.Artifacts.VisualHash))

	return []byte(sb.String())
}
