package integration

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/analysis"
)

// Bridge connects the TUI to the Core Services.
// It abstracts away the complexity of setting up the orchestrator and handling context.
type Bridge struct {
	orchestrator *analysis.Orchestrator
}

// NewBridge initializes the bridge and the underlying orchestrator.
func NewBridge() *Bridge {
	return &Bridge{
		orchestrator: analysis.NewOrchestrator(),
	}
}

// ScanURL triggers the full analysis pipeline for a given URL.
// It returns the final SAL or an error.
func (b *Bridge) ScanURL(url string) (domain.SAL, error) {
	// 1. Ingestion / Normalization (Mini-version of Ingestion Service)
	scanID := fmt.Sprintf("%x", sha256.Sum256([]byte(url+time.Now().String())))

	initialSAL := domain.SAL{
		ScanID:          scanID,
		URL:             url,
		FinalURL:        url, // Assume same for start, updated by engines if redirect found
		Timestamp:       time.Now(),
		IngestionSource: "TUI-Console",
		Signals:         []domain.Signal{},
		Artifacts: domain.Artifacts{
			// In a real flow, we'd fetch the screenshot hash here or pass it if available
			VisualHash: "",
		},
	}

	// 2. Orchestratration (Analysis Phase)
	// We use a background context for now, but in TUI we might want cancellable context
	ctx := context.Background()

	finalSAL, err := b.orchestrator.ProcessArtifact(ctx, initialSAL)
	if err != nil {
		return domain.SAL{}, err
	}

	return finalSAL, nil
}

// Close cleans up resources.
func (b *Bridge) Close() {
	if b.orchestrator != nil {
		b.orchestrator.Close()
	}
}
