package engine

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// ArtifactFactory manages processors and orchestrates ingestion.
type ArtifactFactory struct {
	processors map[ArtifactType]Processor
}

func NewArtifactFactory() *ArtifactFactory {
	return &ArtifactFactory{
		processors: make(map[ArtifactType]Processor),
	}
}

func (f *ArtifactFactory) RegisterProcessor(t ArtifactType, p Processor) {
	p.SetFactory(f)
	f.processors[t] = p
}

func (f *ArtifactFactory) GetProcessor(t ArtifactType) (Processor, error) {
	p, ok := f.processors[t]
	if !ok {
		return nil, fmt.Errorf("no processor found for type: %s", t)
	}
	return p, nil
}

// Ingest processes an artifact using the registered processor.
// It handles the high-level flow: Process -> Returns Normalized Structure.
func (f *ArtifactFactory) Ingest(ctx context.Context, t ArtifactType, input []byte, sourceID string, metadata map[string]interface{}) (*IngestedArtifact, error) {
	processor, err := f.GetProcessor(t)
	if err != nil {
		return nil, err
	}

	artifact, err := processor.Process(ctx, input, sourceID, metadata)
	if err != nil {
		return nil, fmt.Errorf("processing failed: %w", err)
	}

	// Calculate ID if not already set by processor (for dedup based on refined content)
	if artifact.ID == "" {
		hash := sha256.Sum256(input)
		artifact.ID = base64.RawURLEncoding.EncodeToString(hash[:])
	}

	return artifact, nil
}
