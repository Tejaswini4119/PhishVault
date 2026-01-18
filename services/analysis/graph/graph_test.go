package graph_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/analysis/graph"
)

// MockGraphDB captures batch operations for assertion.
type MockGraphDB struct {
	Nodes []domain.GraphNode
	Edges []domain.GraphEdge
}

func (m *MockGraphDB) ExecuteBatch(nodes []domain.GraphNode, edges []domain.GraphEdge) error {
	m.Nodes = append(m.Nodes, nodes...)
	m.Edges = append(m.Edges, edges...)
	return nil
}

func (m *MockGraphDB) Close() error { return nil }

func TestProjector_ProjectSAL(t *testing.T) {
	// Setup
	mockDB := &MockGraphDB{}
	projector := graph.NewProjector(10, mockDB)
	// Silence or redirect logger if needed?
	// The new projector uses log/slog.
	// Ideally we could inject a logger that writes to Test output, but default stdout is fine.

	sal := domain.SAL{
		ScanID:          "test-scan-123",
		URL:             "http://login.secure-update.com/signin", // Lure
		FinalURL:        "http://phish-origin.xyz/bad",           // Origin
		Timestamp:       time.Now(),
		IngestionSource: "Test",
		Verdict:         "MALICIOUS",
		RedirectChain: []string{
			"http://bit.ly/short", // Relay
		},
		CampaignID: "camp-alpha",
		Entities: []domain.Entity{
			{Type: "ASN", Value: "AS666"},
		},
	}

	// Execute
	projector.ProjectSAL(sal)
	projector.Close() // Flushes buffer

	// Verify
	if len(mockDB.Nodes) == 0 {
		t.Fatal("Expected nodes to be projected, got 0")
	}

	// Check Scan Node
	foundScan := false
	for _, n := range mockDB.Nodes {
		if n.Key == "test-scan-123" && n.Label == "ScanArtifact" {
			foundScan = true
			break
		}
	}
	if !foundScan {
		t.Errorf("ScanArtifact node not found")
	}

	// Check Edge: Lure -> Relay
	foundRedirect := false
	for _, e := range mockDB.Edges {
		if e.SourceKey == "login.secure-update.com" && e.TargetKey == "bit.ly" && e.Relation == "REDIRECTS_TO" {
			foundRedirect = true
			break
		}
	}
	if !foundRedirect {
		t.Errorf("Expected REDIRECTS_TO edge from Lure to Relay")
	}

	// Check Campaign Link
	foundCampaign := false
	for _, e := range mockDB.Edges {
		if e.TargetKey == "camp-alpha" && e.Relation == "PART_OF" {
			foundCampaign = true
			break
		}
	}
	if !foundCampaign {
		t.Errorf("Expected PART_OF edge to Campaign")
	}

	slog.Info("Graph Verification Test Passed")
}
