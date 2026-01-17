package analysis

import (
	"context"
	"fmt"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/analysis/ai"
	"github.com/PhishVault/PhishVault-2/services/analysis/decision"
	"github.com/PhishVault/PhishVault-2/services/analysis/graph"
)

// Orchestrator manages the flow of intelligence between engines.
type Orchestrator struct {
	graphProjector *graph.Projector
}

func NewOrchestrator() *Orchestrator {
	// Initialize singletons
	ai.InitGoldenSet()
	ai.InitBayesian()

	return &Orchestrator{
		graphProjector: graph.NewProjector(10, nil), // Batch size 10, Console DB
	}
}

// ProcessArtifact runs the full Phase 2 & 3 pipeline on a Scan Result.
func (o *Orchestrator) ProcessArtifact(ctx context.Context, input domain.SAL) (domain.SAL, error) {
	// 1. Visual Analysis & Golden Set
	// Assume we have the screenshot bytes/path. For MVP, we simulate hash generation or use input.
	// In a real flow, checking if VisualHash is already computed by scanner or needs computing here.
	// Let's assume Scanner provided the Hash (Phase 1.5) or we compute it.

	// Mock: Compute Hash (In prod: ai.ComputePHash(image))
	// For this code, we assume it's passed or we use a dummy.
	visualHash := uint64(0x1234567812345678) // Placeholder for demo

	input.Artifacts.VisualHash = fmt.Sprintf("%x", visualHash)

	// Check Golden Set
	brand, similarity, isMatch := ai.GlobalGoldenSet.FindMatch(visualHash)
	if isMatch {
		sig := domain.Signal{
			EngineName: "VisualAI",
			SignalKey:  "BRAND_IMPERSONATION",
			Confidence: similarity,
			Weight:     1.0,
			Evidence:   map[string]interface{}{"target_brand": brand},
			Tags:       []string{"visual_clone"},
		}
		input.Signals = append(input.Signals, sig)
	}

	// 2. Deep NLP & Structure
	// Use the Text Pipeline
	// Need to fetch content? Assuming SAL has DOMPath or embedded content?
	// The current SAL struct struct doesn't have "Content", but text_pipeline takes string.
	// We'll mock the content loading for now, as Scanner would save it to MinIO.
	// We will use a placeholder "input.URL" content for the logic.

	// AnalyzeContent returns detailed risk struct.
	// We map that risk struct to Signals.
	// Passing the same HTML as Text for now (mocking text extraction)
	risk := ai.AnalyzeContent("<html>...</html>", "Verify your account", input.FinalURL)

	if risk.Intent != "Benign" {
		sig := domain.Signal{
			EngineName: "NLP_Deep",
			SignalKey:  "INTENT_" + risk.Intent,
			Confidence: risk.UrgencyScore, // Using urgency as proxy
			Weight:     0.8,
			Tags:       []string{risk.Intent},
		}
		input.Signals = append(input.Signals, sig)
	}

	if risk.FormRisk.HasPassword {
		input.Signals = append(input.Signals, domain.Signal{
			EngineName: "Structure",
			SignalKey:  "SENSITIVE_FORM",
			Weight:     0.5,
		})
	}

	// 3. Verdict (OPA)
	// Map Signals to OPA Input
	opaInput := decision.PolicyInput{
		VisualMatchScore: similarity,
		UrgencyScore:     risk.UrgencyScore,
		Intent:           risk.Intent,
		HasLoginForm:     risk.FormRisk.HasPassword,
		DomainAgeDays:    0, // need enrichment source
	}

	verdict, err := decision.EvaluateVerdict(ctx, opaInput)
	if err != nil {
		// Log error but proceed with UNKNOWN? Or fail?
		// For robustness, we log and keep going if possible, but Verdict is critical.
		// Since EvaluateVerdict wrapper handles error gracefully (returning default), just log.
		fmt.Printf("OPA Evaluation Warning: %v\n", err)
	}
	input.Verdict = verdict.Verdict
	input.RiskScore = verdict.RiskScore

	// 4. Graph Projection (Phase 3)
	// Project the enriched SAL into the Campaign Graph
	o.graphProjector.ProjectSAL(input)

	return input, nil
}
