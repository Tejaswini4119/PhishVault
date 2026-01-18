package graph

import (
	"testing"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

// Mock Adapter for Testing
type TestGraphAdapter struct {
	Nodes []domain.GraphNode
	Edges []domain.GraphEdge
}

func (t *TestGraphAdapter) ExecuteBatch(nodes []domain.GraphNode, edges []domain.GraphEdge) error {
	t.Nodes = append(t.Nodes, nodes...)
	t.Edges = append(t.Edges, edges...)
	return nil
}

func (t *TestGraphAdapter) Close() error { return nil }

func TestProjector(t *testing.T) {
	// 1. Setup Mock DB
	mockDB := &TestGraphAdapter{}
	p := NewProjector(2, mockDB) // Small batch size

	// 2. Mock SAL Input
	sal1 := domain.SAL{
		ScanID:    "scan1",
		URL:       "http://phishing.com/login",
		FinalURL:  "http://phishing.com/login",
		Timestamp: time.Now(),
	}

	// 3. Execute Projection
	p.ProjectSAL(sal1)

	// Force flush (since flush happens async or on batch limit, explicit Close is safer)
	p.Close()

	// 4. Assertions on the Mock DB
	// Expect: ScanNode, Lure(phishing.com), Edge(Captured)
	if len(mockDB.Nodes) == 0 {
		t.Fatal("Expected nodes in GraphDB, got 0")
	}

	foundScan := false
	for _, n := range mockDB.Nodes {
		// LabelScan string literal "ScanArtifact" used in projector.go
		if n.Label == "ScanArtifact" && n.Key == "scan1" {
			foundScan = true
			break
		}
	}

	if !foundScan {
		t.Error("Did not find Scan Artifact Node in GraphDB")
	}

	t.Logf("Graph Projector successfully persisted %d nodes and %d edges to the DB Adapter.", len(mockDB.Nodes), len(mockDB.Edges))
}
