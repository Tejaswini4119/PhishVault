package graph

import (
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
)

// --- Graph Database Interface ---
// Defined in core/domain/graph.go

// --- Console Adapter (Default) ---

// --- Console Adapter (Default) ---

type ConsoleGraphAdapter struct{}

func (c *ConsoleGraphAdapter) ExecuteBatch(nodes []domain.GraphNode, edges []domain.GraphEdge) error {
	if len(nodes) == 0 && len(edges) == 0 {
		return nil
	}
	slog.Info("executing graph batch", "node_count", len(nodes), "edge_count", len(edges))
	// Log sample query for debug
	if len(nodes) > 0 {
		n := nodes[0]
		slog.Debug("sample node merge", "label", n.Label, "id", n.Key)
	}
	return nil
}

func (c *ConsoleGraphAdapter) Close() error { return nil }

// --- Projector Implementation ---

type Projector struct {
	mu            sync.Mutex
	nodeBuffer    []domain.GraphNode
	edgeBuffer    []domain.GraphEdge
	batchSize     int
	flushInterval time.Duration
	db            domain.GraphDatabase
	stopChan      chan struct{}
	logger        *slog.Logger
}

func NewProjector(batchSize int, db domain.GraphDatabase) *Projector {
	if db == nil {
		db = &ConsoleGraphAdapter{}
	}
	p := &Projector{
		batchSize:     batchSize,
		flushInterval: 2 * time.Second,
		nodeBuffer:    make([]domain.GraphNode, 0, batchSize),
		edgeBuffer:    make([]domain.GraphEdge, 0, batchSize),
		db:            db,
		stopChan:      make(chan struct{}),
		logger:        slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
	// Background flush worker
	go p.flushLoop()
	return p
}

func (p *Projector) Close() {
	close(p.stopChan)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.flush()
}

func (p *Projector) ProjectSAL(sal domain.SAL) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// --- 1. Nodes: Artifact & Core Entities ---
	scanNodeKey := sal.ScanID
	scanNode := domain.GraphNode{
		Key:   scanNodeKey,
		Label: "ScanArtifact", // Using string literal as requested
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

	// Edge: ScanArtifact -> Domain (Captured)
	// Currently domain package defines EdgeCaptured? No, I need to check.
	// Assume "CAPTURED" is fine string literal if const missing.
	p.addEdge(scanNodeKey, "ScanArtifact", previousNodeID, domain.LabelDomain, "CAPTURED")

	for i, hopURL := range sal.RedirectChain {
		currentNodeID := extractHost(hopURL)
		p.addDomainNode(currentNodeID, "Relay")
		p.addEdge(previousNodeID, domain.LabelDomain, currentNodeID, domain.LabelDomain, domain.EdgeResolvesTo) // Should be "REDIRECTS_TO"?
		// Using "REDIRECTS_TO" literal as originally intended, or strictly defined in domain?
		// domain.EdgeResolvesTo is for IP. Let's use "REDIRECTS_TO" literal.

		// Update the last added edge with properties
		lastIdx := len(p.edgeBuffer) - 1
		p.edgeBuffer[lastIdx].Relation = "REDIRECTS_TO" // Override if needed
		p.edgeBuffer[lastIdx].Properties = map[string]interface{}{
			"hop_index":  i,
			"confidence": calculateHopConfidence(i),
		}

		previousNodeID = currentNodeID
	}

	finalHost := extractHost(sal.FinalURL)
	if finalHost == "" {
		finalHost = previousNodeID
	}
	if finalHost != previousNodeID {
		p.addDomainNode(finalHost, "Origin")
		p.addEdge(previousNodeID, domain.LabelDomain, finalHost, domain.LabelDomain, "REDIRECTS_TO")
	}

	// --- 3. Campaign Clustering ---
	if sal.CampaignID != "" {
		campNode := domain.GraphNode{
			Key:   sal.CampaignID,
			Label: domain.LabelCampaign,
			Properties: map[string]interface{}{
				"first_seen": sal.Timestamp.Unix(),
			},
		}
		p.nodeBuffer = append(p.nodeBuffer, campNode)
		p.addEdge(finalHost, domain.LabelDomain, sal.CampaignID, domain.LabelCampaign, domain.EdgePartOf)
	}

	// --- 4. Entities ---
	for _, entity := range sal.Entities {
		if entity.Type == "ASN" {
			p.addDomainNode(entity.Value, "ASN")
			// domain.LabelASN and domain.EdgeHostedOn
			p.addEdge(finalHost, domain.LabelDomain, entity.Value, domain.LabelASN, domain.EdgeHostedOn)
		}
	}

	if len(p.nodeBuffer) >= p.batchSize {
		p.flush()
	}
}

func (p *Projector) addDomainNode(host string, role string) {
	n := domain.GraphNode{
		Key:   host,
		Label: domain.LabelDomain,
		Properties: map[string]interface{}{
			"role": role,
		},
	}
	p.nodeBuffer = append(p.nodeBuffer, n)
}

func (p *Projector) addEdge(srcKey, srcType, tgtKey, tgtType string, relation string) {
	p.edgeBuffer = append(p.edgeBuffer, domain.GraphEdge{
		SourceKey:  srcKey,
		SourceType: srcType,
		TargetKey:  tgtKey,
		TargetType: tgtType,
		Relation:   relation,
		Properties: make(map[string]interface{}),
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

	err := p.db.ExecuteBatch(p.nodeBuffer, p.edgeBuffer)
	if err != nil {
		p.logger.Error("failed to flush graph batch", "error", err)
	}

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
