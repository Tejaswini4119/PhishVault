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
			// Score: (0.9*0.4) + (0.9*0.2) + (1.0*0.3) + 0.2
			// = 0.36 + 0.18 + 0.3 + 0.2 = 1.04 -> 1.0 (capped in logic? No cap in rego, but visual+urgency+intent max is usually 1.0. Wait.
			// 0.4 + 0.2 + 0.3 + 0.2 = 1.1. So it can exceed 1.0.
			wantRisk:    0.7, // Expecting at least > 0.7
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
			// Score: (0*0.4) + (0.1*0.2) + (0*0.3) + 0
			// = 0.02
			wantRisk:    0.01, // Check for non-zero
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
