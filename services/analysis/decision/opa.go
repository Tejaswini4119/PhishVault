package decision

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
)

// PolicyInput represents the data fed into OPA.
type PolicyInput struct {
	VisualMatchScore float64 `json:"visual_match_score"`
	UrgencyScore     float64 `json:"urgency_score"`
	Intent           string  `json:"intent"` // e.g. "CredentialHarvesting"
	HasLoginForm     bool    `json:"has_login_form"`
	DomainAgeDays    int     `json:"domain_age_days"`
}

// VerdictResult holds the output from OPA.
type VerdictResult struct {
	Verdict   string  `json:"verdict"`
	RiskScore float64 `json:"risk_score"`
}

// Embed the policy file
//
//go:embed phishing.rego

var policyData string

// EvaluateVerdict runs the OPA policy against the input.
func EvaluateVerdict(ctx context.Context, input PolicyInput) (VerdictResult, error) {
	// Prepare Rego query
	query, err := rego.New(
		rego.Query("data.phishvault.policy"),
		rego.Module("phishing.rego", policyData),
	).PrepareForEval(ctx)

	if err != nil {
		return VerdictResult{}, fmt.Errorf("failed to prepare rego: %w", err)
	}

	// Evaluate
	results, err := query.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return VerdictResult{}, fmt.Errorf("failed to evaluate policy: %w", err)
	}
	if len(results) == 0 {
		return VerdictResult{Verdict: "UNKNOWN", RiskScore: 0}, nil
	}

	// Extract results (slightly complex with OPA's generic return type)
	bindings := results[0].Bindings

	verdict, ok := bindings["verdict"].(string)
	if !ok {
		verdict = "UNKNOWN"
	}

	// Bindings map might not contain calculated values if we query the whole package.
	// Let's refine. The query "data.phishvault.policy" returns the whole package object.
	expressions := results[0].Expressions
	if len(expressions) > 0 {
		val, ok := expressions[0].Value.(map[string]interface{})
		if ok {
			v, _ := val["verdict"].(string)
			r, _ := val["risk_score"].(float64) // JSON numbers are float64
			// OPA might return encoding/json.Number
			return VerdictResult{Verdict: v, RiskScore: r}, nil
		}
	}

	return VerdictResult{Verdict: verdict, RiskScore: 0}, nil
}
