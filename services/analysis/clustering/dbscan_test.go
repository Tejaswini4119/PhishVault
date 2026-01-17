package clustering

import (
	"testing"
)

func TestDBSCAN(t *testing.T) {
	// Setup Points
	// Cluster 1: Similar Visual Hash, Same Subnet
	p1 := &FeatureVector{ID: "A", IP: "192.168.1.5", VisualHash: 0x1234567812345678, DOMTokens: []string{"login", "password"}}
	p2 := &FeatureVector{ID: "B", IP: "192.168.1.20", VisualHash: 0x1234567812345679, DOMTokens: []string{"login", "password"}} // Very close hash (1 bit diff)

	// Noise: Totally different
	p3 := &FeatureVector{ID: "C", IP: "10.0.0.1", VisualHash: 0xFFFFFFFFFFFFFFFF, DOMTokens: []string{"blog", "post"}}

	points := []*FeatureVector{p1, p2, p3}

	// Run DBSCAN
	clusters := RunDBSCAN(points)

	// Assertions
	if len(clusters) != 1 {
		t.Fatalf("Expected 1 cluster, got %d", len(clusters))
	}

	c1 := clusters[0]
	if len(c1.Points) != 2 {
		t.Errorf("Expected cluster size 2, got %d", len(c1.Points))
	}

	// Verify membership
	hasA := false
	hasB := false
	for _, p := range c1.Points {
		if p.ID == "A" {
			hasA = true
		}
		if p.ID == "B" {
			hasB = true
		}
	}

	if !hasA || !hasB {
		t.Error("Cluster should contain points A and B")
	}

	t.Logf("DBSCAN Result: Found %d clusters. Cluster 1 size: %d", len(clusters), len(c1.Points))
}

func TestHamming(t *testing.T) {
	h1 := uint64(0xFFFF)
	h2 := uint64(0xFFFE) // 1 bit diff
	dist := hamming(h1, h2)
	if dist != 1 {
		t.Errorf("Expected hamming distance 1, got %d", dist)
	}
}
