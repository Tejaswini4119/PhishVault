package graph

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

// --- Graph Domain Model ---

type NodeLabel string
type EdgeType string

const (
	LabelDomain NodeLabel = "Domain"
	LabelIP     NodeLabel = "IP"
	LabelASN    NodeLabel = "ASN"
	LabelCert   NodeLabel = "Certificate"
	LabelScan   NodeLabel = "ScanArtifact"

	EdgeResolvesTo EdgeType = "RESOLVES_TO"
	EdgeHostedOn   EdgeType = "HOSTED_ON"
	EdgeIssuedBy   EdgeType = "ISSUED_BY"
	EdgeCaptured   EdgeType = "CAPTURED"
)

type GraphNode struct {
	ID         string
	Label      NodeLabel
	Properties map[string]interface{}
}

type GraphEdge struct {
	SourceID string
	TargetID string
	Type     EdgeType
	Props    map[string]interface{}
}

// --- Graph Database Interface ---

// GraphDatabase defines the contract for persisting graph structures.
// This allows swapping the backend (Console, Neo4j, Neptune) without changing logic.
type GraphDatabase interface {
	ExecuteBatch(nodes []GraphNode, edges []GraphEdge) error
	Close() error
}

// --- Console Adapter (Default) ---

type ConsoleGraphAdapter struct{}

func (c *ConsoleGraphAdapter) ExecuteBatch(nodes []GraphNode, edges []GraphEdge) error {
	if len(nodes) == 0 && len(edges) == 0 {
		return nil
	}
	fmt.Printf("--- [GraphDB] Executing Batch: %d Nodes, %d Edges ---\n", len(nodes), len(edges))
	// Log sample query for debug
	if len(nodes) > 0 {
		n := nodes[0]
		fmt.Printf("    Sample Node: MERGE (n:%s {id: '%s'}) ...\n", n.Label, n.ID)
	}
	return nil
}

func (c *ConsoleGraphAdapter) Close() error { return nil }

// --- Projector Implementation ---

// Projector handles the projection of linear SAL data into Graph structures.
type Projector struct {
	mu            sync.Mutex
	nodeBuffer    []GraphNode
	edgeBuffer    []GraphEdge
	batchSize     int
	flushInterval time.Duration
	db            GraphDatabase // Injected dependency
	stopChan      chan struct{}
}

func NewProjector(batchSize int, db GraphDatabase) *Projector {
	if db == nil {
		db = &ConsoleGraphAdapter{}
	}
	p := &Projector{
		batchSize:     batchSize,
		flushInterval: 2 * time.Second,
		nodeBuffer:    make([]GraphNode, 0, batchSize),
		edgeBuffer:    make([]GraphEdge, 0, batchSize),
		db:            db,
		stopChan:      make(chan struct{}),
	}
	// Background flush worker
	go p.flushLoop()
	return p
}

// Close gracefully shuts down the projector and flushes remaining items.
func (p *Projector) Close() {
	close(p.stopChan)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.flush()
}

// ProjectSAL takes a Scan Abstraction Layer object and decomposes it into Graph primitives.
func (p *Projector) ProjectSAL(sal domain.SAL) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// ... (Mapping Logic remains same, omitted for brevity in diff but assumed present if I don't delete it?
	// WOOPS, replace_file_content wipes the block. I need to keep the mapping logic!)

	// Actually, the user asked for robust engineering. I should probably re-paste the mapping logic
	// but purely. Since I am replacing the Whole File content from type Projector downwards,
	// I must include the ProjectSAL logic again.

	// --- 1. Nodes: Artifact & Core Entities ---
	scanNode := GraphNode{
		ID:    sal.ScanID,
		Label: LabelScan,
		Properties: map[string]interface{}{
			"timestamp": sal.Timestamp.Unix(),
			"url":       sal.URL,
			"source":    sal.IngestionSource,
			"verdict":   sal.Verdict,
		},
	}
	p.nodeBuffer = append(p.nodeBuffer, scanNode)

	// --- 2. Origin Path Mapping ---
	previousNodeID := extractHost(sal.URL)
	p.addDomainNode(previousNodeID, "Lure")
	p.addEdge(scanNode.ID, previousNodeID, EdgeCaptured)

	for i, hopURL := range sal.RedirectChain {
		currentNodeID := extractHost(hopURL)
		p.addDomainNode(currentNodeID, "Relay")
		p.addEdge(previousNodeID, currentNodeID, "REDIRECTS_TO")
		previousNodeID = currentNodeID

		p.edgeBuffer[len(p.edgeBuffer)-1].Props = map[string]interface{}{
			"hop_index":  i,
			"confidence": calculateHopConfidence(i),
		}
	}

	finalHost := extractHost(sal.FinalURL)
	if finalHost == "" {
		finalHost = previousNodeID
	}
	if finalHost != previousNodeID {
		p.addDomainNode(finalHost, "Origin")
		p.addEdge(previousNodeID, finalHost, "REDIRECTS_TO")
	}

	// --- 3. Campaign Clustering ---
	if sal.CampaignID != "" {
		campNode := GraphNode{
			ID:    sal.CampaignID,
			Label: "Campaign",
			Properties: map[string]interface{}{
				"first_seen": sal.Timestamp.Unix(),
			},
		}
		p.nodeBuffer = append(p.nodeBuffer, campNode)
		p.addEdge(finalHost, sal.CampaignID, "PART_OF")
	}

	// --- 4. Entities ---
	for _, entity := range sal.Entities {
		if entity.Type == "ASN" {
			p.addDomainNode(entity.Value, "ASN")
			p.addEdge(finalHost, entity.Value, EdgeHostedOn)
		}
	}

	if len(p.nodeBuffer) >= p.batchSize {
		p.flush()
	}
}

func (p *Projector) addDomainNode(host string, role string) {
	n := GraphNode{
		ID:    host,
		Label: LabelDomain,
		Properties: map[string]interface{}{
			"role": role,
		},
	}
	p.nodeBuffer = append(p.nodeBuffer, n)
}

func (p *Projector) addEdge(src, tgt string, edgeType EdgeType) {
	p.edgeBuffer = append(p.edgeBuffer, GraphEdge{
		SourceID: src, TargetID: tgt, Type: edgeType,
	})
}

func calculateHopConfidence(hopIndex int) float64 {
	base := 0.5
	increase := float64(hopIndex) * 0.1
	conf := base + increase
	if conf > 0.95 {
		return 0.95
	}
	return conf
}

func (p *Projector) flushLoop() {
	ticker := time.NewTicker(p.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			if len(p.nodeBuffer) > 0 {
				p.flush()
			}
			p.mu.Unlock()
		case <-p.stopChan:
			return
		}
	}
}

func (p *Projector) flush() {
	if len(p.nodeBuffer) == 0 && len(p.edgeBuffer) == 0 {
		return
	}

	// Delegate to the injected database adapter
	err := p.db.ExecuteBatch(p.nodeBuffer, p.edgeBuffer)
	if err != nil {
		fmt.Printf("Error flushing graph batch: %v\n", err)
	}

	// Clear buffers (re-slice to 0, keep capacity)
	p.nodeBuffer = p.nodeBuffer[:0]
	p.edgeBuffer = p.edgeBuffer[:0]
}

func extractHost(u string) string {
	if strings.Contains(u, "://") {
		parts := strings.Split(u, "://")
		if len(parts) > 1 {
			u = parts[1]
		}
	}
	parts := strings.Split(u, "/")
	return parts[0]
}
