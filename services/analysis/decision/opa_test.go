package decision

import (
	"context"
	"testing"
)

func TestEvaluateVerdict(t *testing.T) {
	tests := []struct {
		name        string
		input       PolicyInput
		wantRisk    float64
		wantVerdict string
	}{
		{
			name: "High Value Phishing",
			input: PolicyInput{
				VisualMatchScore: 0.9,
				UrgencyScore:     0.9,
				Intent:           "CredentialHarvesting",
				HasLoginForm:     true,
				DomainAgeDays:    5,
			},
			// Dynamic Multiplier Calculation:
			// Base: (0.9*0.3) + (0.9*0.2) + (0.8*0.5) = 0.85
			// Boosters: Domain(1.6) * Brand(1.5) * Form(1.2) = 2.88
			// Total: 0.85 * 2.88 = 2.44 -> Capped at 1.0
			wantRisk:    1.0,
			wantVerdict: "MALICIOUS",
		},
		{
			name: "Benign",
			input: PolicyInput{
				VisualMatchScore: 0.0,
				UrgencyScore:     0.1,
				Intent:           "Benign",
				HasLoginForm:     false,
				DomainAgeDays:    100,
			},
			// Base: (0.1*0.2) = 0.02
			// Boosters: 1.0
			wantRisk:    0.02,
			wantVerdict: "SAFE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateVerdict(context.Background(), tt.input)
			if err != nil {
				t.Fatalf("EvaluateVerdict() error = %v", err)
			}
			t.Logf("Got RiskScore: %f, Verdict: %s", got.RiskScore, got.Verdict)
			if got.RiskScore < tt.wantRisk && tt.name == "High Value Phishing" {
				t.Errorf("EvaluateVerdict() RiskScore = %v, want >= %v", got.RiskScore, tt.wantRisk)
			}
			if got.RiskScore == 0 && tt.name == "Benign" {
				// 0.02 should not be 0
				t.Errorf("EvaluateVerdict() RiskScore = %v, want > 0", got.RiskScore)
			}
		})
	}
}
