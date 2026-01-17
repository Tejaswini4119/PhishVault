package analysis_test

import (
	"context"
	"testing"

	"github.com/PhishVault/PhishVault-2/services/analysis/ai"
	"github.com/PhishVault/PhishVault-2/services/analysis/decision"
)

func TestNLP(t *testing.T) {
	text := "Your account will be suspended within 24 hours. Action Required Immediately!"
	result := ai.AnalyzeText(text)

	if result.UrgencyScore < 0.5 {
		t.Errorf("Expected high urgency, got %f", result.UrgencyScore)
	}
	if len(result.Keywords) == 0 {
		t.Error("Expected keywords to be found")
	}
	t.Logf("NLP Result: Score=%f, Keywords=%v", result.UrgencyScore, result.Keywords)
}

func TestOPA(t *testing.T) {
	// Malicious input
	input := decision.PolicyInput{
		VisualMatchScore: 0.9,
		NLPUrgencyScore:  0.8,
		DomainAgeDays:    5,
	}

	result, err := decision.EvaluateVerdict(context.Background(), input)
	if err != nil {
		t.Fatalf("OPA Eval failed: %v", err)
	}

	if result.Verdict != "MALICIOUS" {
		t.Errorf("Expected MALICIOUS, got %s (Score: %f)", result.Verdict, result.RiskScore)
	}
	t.Logf("OPA Result: Verdict=%s, Risk=%f", result.Verdict, result.RiskScore)

	// Safe input
	safeInput := decision.PolicyInput{
		VisualMatchScore: 0.1,
		NLPUrgencyScore:  0.0,
		DomainAgeDays:    100,
	}
	safeResult, err := decision.EvaluateVerdict(context.Background(), safeInput)
	if err != nil {
		t.Fatalf("OPA Eval failed: %v", err)
	}
	if safeResult.Verdict != "SAFE" {
		t.Errorf("Expected SAFE, got %s", safeResult.Verdict)
	}
}
