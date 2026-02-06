package engine

import (
	"context"
	"time"
)

// ArtifactType represents the type of artifact being ingested.
type ArtifactType string

const (
	ArtifactTypeURL   ArtifactType = "URL"
	ArtifactTypeEmail ArtifactType = "EMAIL"
	ArtifactTypeFile  ArtifactType = "FILE"
)

// IngestedArtifact represents the normalized internal schema for any artifact.
type IngestedArtifact struct {
	ID              string                 `json:"id"`                  // Unique ID (SHA256 of content)
	Type            ArtifactType           `json:"type"`                // Type of artifact
	RawPath         string                 `json:"raw_path,omitempty"`  // Path to raw blob in storage
	SourceID        string                 `json:"source_id"`           // ID of the scan/request that triggered this
	Metadata        map[string]interface{} `json:"metadata"`            // Extracted metadata
	Content         string                 `json:"content,omitempty"`   // Text content or canonical URL
	ParentID        string                 `json:"parent_id,omitempty"` // For nested artifacts (e.g. attachment in email)
	Children        []*IngestedArtifact    `json:"children,omitempty"`  // Decomposed sub-artifacts
	Timestamp       time.Time              `json:"timestamp"`
	IngestionSource string                 `json:"ingestion_source"` // API, FEED, etc.
}

// Processor interface handles the specific logic for parsing and normalizing an artifact type.
type Processor interface {
	// SetFactory injects the ArtifactFactory to allow recursive ingestion.
	SetFactory(f *ArtifactFactory)
	// Process takes raw input and returns a normalized IngestedArtifact and any sub-artifacts.
	Process(ctx context.Context, input []byte, sourceID string, metadata map[string]interface{}) (*IngestedArtifact, error)
}
