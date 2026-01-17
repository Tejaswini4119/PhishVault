package clustering

import (
	"strings"
)

// FeatureVector represents the multi-dimensional data of a Scan.
type FeatureVector struct {
	ID         string
	VisualHash uint64 // Perceptual Hash
	IP         string
	DOMTokens  []string // Key tokens for structure
}

// Cluster represents a identified campaign.
type Cluster struct {
	ID     int
	Points []*FeatureVector
}

// DBSCAN parameters
const (
	Epsilon = 0.3 // Custom distance threshold (0.0 - 1.0)
	MinPts  = 2   // Minimum points to form a cluster
)

// RunDBSCAN executes the clustering algorithm.
func RunDBSCAN(points []*FeatureVector) []Cluster {
	clusters := []Cluster{}
	visited := make(map[string]bool)
	noise := make(map[string]bool)
	clusterID := 0

	for _, p := range points {
		if visited[p.ID] {
			continue
		}
		visited[p.ID] = true

		neighbors := regionQuery(points, p, Epsilon)

		if len(neighbors) < MinPts {
			noise[p.ID] = true
		} else {
			clusterID++
			c := Cluster{ID: clusterID, Points: []*FeatureVector{p}}
			clusters = append(clusters, expandCluster(points, p, neighbors, &c, visited, noise))
		}
	}

	// Filter out empty clusters if any logic caused them
	final := []Cluster{}
	for _, c := range clusters {
		if len(c.Points) > 0 {
			final = append(final, c)
		}
	}
	return final
}

func expandCluster(allPoints []*FeatureVector, p *FeatureVector, neighbors []*FeatureVector, c *Cluster, visited map[string]bool, noise map[string]bool) Cluster {
	// We use a queue-like approach, but simplest slice iteration works
	// Note: neighbors is modified during iteration in standard DBSCAN algo descriptions,
	// but in Go slice iteration, appending doesn't affect the range loop if strictly range.
	// We need a standard for loop.

	// Initial neighbors are part of cluster
	for _, n := range neighbors {
		if n.ID != p.ID { // Don't add P again if it's in neighbors
			// Check if point is already in a cluster?
			// Standard DBSCAN: if P' is not visited, mark visited. If P' is not member of any cluster, add to C.
			// Simplified here: All neighbors passed in are clearly Density-Reachable from P.
		}
	}

	q := neighbors
	for i := 0; i < len(q); i++ {
		pointQ := q[i]

		if !visited[pointQ.ID] {
			visited[pointQ.ID] = true
			neighborsQ := regionQuery(allPoints, pointQ, Epsilon)
			if len(neighborsQ) >= MinPts {
				q = append(q, neighborsQ...)
			}
		}

		// Add to cluster if not already member of another cluster
		// (Simplified: we assume if we visit it here, it belongs here or was Noise)
		// For MVP: Check if contained.
		if !contains(c.Points, pointQ) {
			c.Points = append(c.Points, pointQ)
		}
	}

	return *c
}

func regionQuery(points []*FeatureVector, center *FeatureVector, eps float64) []*FeatureVector {
	neighbors := []*FeatureVector{}
	for _, p := range points {
		if distance(center, p) <= eps {
			neighbors = append(neighbors, p)
		}
	}
	return neighbors
}

// distance calculates the "Campaign Similarity" (0.0 = Identical, 1.0 = Different).
// Weighted mix of Visual, IP, and DOM.
func distance(a, b *FeatureVector) float64 {
	// 1. Visual Distance (Hamming on 64-bit Hash)
	// Hamming distance / 64 provides 0.0-1.0
	visualDist := float64(hamming(a.VisualHash, b.VisualHash)) / 64.0

	// 2. IP Distance (0.0 if same /24 subnet, 1.0 otherwise)
	ipDist := 1.0
	if isSameSubnet(a.IP, b.IP) {
		ipDist = 0.0
	} else if a.IP == b.IP {
		ipDist = 0.0
	}

	// 3. DOM Token Overlap (Jaccard Distance)
	domDist := jaccardDistance(a.DOMTokens, b.DOMTokens)

	// Weighted Formula
	// Visual is strongest indicator of Phishing Kits (cloned UI).
	// Infrastructure (IP) is secondary.
	// Weight: Visual 60%, IP 20%, DOM 20%

	total := (visualDist * 0.6) + (ipDist * 0.2) + (domDist * 0.2)
	return total
}

// Helpers

func hamming(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	dist := 0
	for xor > 0 {
		if xor&1 == 1 {
			dist++
		}
		xor >>= 1
	}
	return dist
}

func isSameSubnet(ip1, ip2 string) bool {
	// Simple string check for MVP (x.x.x.*)
	parts1 := strings.Split(ip1, ".")
	parts2 := strings.Split(ip2, ".")
	if len(parts1) == 4 && len(parts2) == 4 {
		return parts1[0] == parts2[0] && parts1[1] == parts2[1] && parts1[2] == parts2[2]
	}
	return false
}

func jaccardDistance(t1, t2 []string) float64 {
	if len(t1) == 0 && len(t2) == 0 {
		return 0.0
	}

	set := make(map[string]bool)
	for _, t := range t1 {
		set[t] = true
	}

	intersection := 0
	for _, t := range t2 {
		if set[t] {
			intersection++
		}
	}

	union := len(t1) + len(t2) - intersection
	if union == 0 {
		return 1.0
	}

	similarity := float64(intersection) / float64(union)
	return 1.0 - similarity // Distance
}

func contains(list []*FeatureVector, p *FeatureVector) bool {
	for _, item := range list {
		if item.ID == p.ID {
			return true
		}
	}
	return false
}
