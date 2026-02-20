package analysis

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/PhishVault/PhishVault-2/core/domain"
	"github.com/PhishVault/PhishVault-2/services/analysis/ai"
	"github.com/PhishVault/PhishVault-2/services/analysis/clustering"
	"github.com/PhishVault/PhishVault-2/services/analysis/config"
	"github.com/PhishVault/PhishVault-2/services/analysis/decision"
	"github.com/PhishVault/PhishVault-2/services/analysis/graph"
	"github.com/PhishVault/PhishVault-2/services/intel"
	"github.com/PhishVault/PhishVault-2/services/intel/provider"
)

// Orchestrator manages the flow of intelligence between engines.
type Orchestrator struct {
	graphProjector *graph.Projector
	logger         *slog.Logger
	cfg            *config.AnalysisConfig

	// Clustering State
	mu            sync.Mutex
	clusterBuffer []*clustering.FeatureVector
}

func NewOrchestrator() *Orchestrator {
	// Initialize singletons
	ai.InitGoldenSet()
	ai.InitBayesian()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.LoadConfig()

	// Connect to Neo4j
	var graphDB domain.GraphDatabase
	neo4jClient, err := intel.NewNeo4jClient(cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword)
	if err != nil {
		logger.Error("failed to connect to neo4j, falling back to console adapter", "error", err)
		// graphDB remains nil, Projector will use ConsoleGraphAdapter
	} else {
		logger.Info("connected to neo4j successfully")
		graphDB = neo4jClient
	}

	return &Orchestrator{
		graphProjector: graph.NewProjector(cfg.GraphBatchSize, graphDB),
		logger:         logger,
		cfg:            cfg,
		clusterBuffer:  make([]*clustering.FeatureVector, 0, cfg.ClusterBatchSize),
	}
}

// ProcessArtifact runs the full Phase 2 & 3 pipeline on a Scan Result.
func (o *Orchestrator) ProcessArtifact(ctx context.Context, input domain.SAL) (domain.SAL, error) {
	o.logger.Info("processing scan artifact", "scan_id", input.ScanID, "url", input.URL)

	// 1. Dispatch by Type
	var similarity float64
	var urgency float64
	var intent string
	var hasLoginForm bool
	var domainAgeDays int
	var visualHash uint64

	switch input.ArtifactType {
	case "URL":
		similarity, urgency, intent, hasLoginForm, domainAgeDays, visualHash = o.analyzeURL(ctx, &input)
	case "EMAIL":
		sig := AnalyzeEmail(input)
		input.Signals = append(input.Signals, sig)
		// Map for OPA
		urgency = sig.Weight
		intent = "EmailPotentialPhish"
	case "FILE":
		sig := AnalyzeFile(input)
		input.Signals = append(input.Signals, sig)
		urgency = sig.Weight
		intent = "FilePotentialPhish"
	}

	// 3. Verdict (OPA)
	opaInput := decision.PolicyInput{
		VisualMatchScore: similarity,
		UrgencyScore:     urgency,
		Intent:           intent,
		HasLoginForm:     hasLoginForm,
		DomainAgeDays:    domainAgeDays,
	}

	verdict, err := decision.EvaluateVerdict(ctx, opaInput)
	if err != nil {
		o.logger.Error("OPA evaluation failed", "error", err)
	} else {
		o.logger.Info("verdict reached", "verdict", verdict.Verdict, "risk_score", verdict.RiskScore)
	}
	input.Verdict = verdict.Verdict
	input.RiskScore = verdict.RiskScore

	// 4. Graph Projection (Phase 3)
	// Project the enriched SAL into the Campaign Graph
	o.graphProjector.ProjectSAL(input)

	// 5. Campaign Clustering Integration
	// Extract features and buffer for clustering
	o.bufferForClustering(input, visualHash)

	return input, nil
}

func (o *Orchestrator) analyzeURL(ctx context.Context, input *domain.SAL) (similarity, urgency float64, intent string, hasLoginForm bool, domainAgeDays int, visualHash uint64) {
	// 1. Visual Analysis & Golden Set
	if input.Artifacts.VisualHash != "" {
		if val, err := strconv.ParseUint(input.Artifacts.VisualHash, 16, 64); err == nil {
			visualHash = val
		}
	}

	if visualHash != 0 {
		brand, score, isMatch := ai.GlobalGoldenSet.FindMatch(visualHash)
		similarity = score
		if isMatch {
			o.logger.Info("visual brand match detected", "brand", brand, "confidence", score)
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
	}

	// 2. Deep NLP & Structure
	htmlContent := input.Artifacts.RawContent
	if htmlContent == "" {
		htmlContent = "<html></html>"
	}
	risk := ai.AnalyzeContent(htmlContent, "Verify your account", input.FinalURL)
	urgency = risk.UrgencyScore
	intent = risk.Intent
	hasLoginForm = risk.FormRisk.HasPassword

	if risk.Intent != "Benign" {
		input.Signals = append(input.Signals, domain.Signal{
			EngineName: "NLP_Deep",
			SignalKey:  "INTENT_" + risk.Intent,
			Confidence: risk.UrgencyScore,
			Weight:     0.8,
			Tags:       []string{risk.Intent},
		})
	}

	if risk.FormRisk.HasPassword {
		input.Signals = append(input.Signals, domain.Signal{
			EngineName: "Structure",
			SignalKey:  "SENSITIVE_FORM",
			Weight:     0.5,
		})
	}

	// 2.5 Threat Intelligence
	u, err := url.Parse(input.FinalURL)
	var domainName string
	if err == nil {
		domainName = u.Hostname()
	} else {
		domainName = input.FinalURL
	}

	if whoisData, err := provider.FetchWHOIS(domainName); err == nil {
		age := time.Since(whoisData.CreationDate).Hours() / 24
		domainAgeDays = int(age)
		if domainAgeDays < 30 {
			input.Signals = append(input.Signals, domain.Signal{
				EngineName: "ThreatIntel",
				SignalKey:  "YOUNG_DOMAIN",
				Confidence: 1.0,
				Weight:     0.6,
				Evidence:   map[string]interface{}{"age_days": domainAgeDays},
			})
		}
	}

	repResults, _ := provider.CheckReputation(domainName)
	for _, res := range repResults {
		if res.Malicious {
			input.Signals = append(input.Signals, domain.Signal{
				EngineName: "ThreatIntel",
				SignalKey:  "KNOWN_MALICIOUS",
				Confidence: res.Score,
				Weight:     1.0,
				Evidence:   map[string]interface{}{"source": res.Source},
			})
		}
	}

	return
}

func (o *Orchestrator) bufferForClustering(sal domain.SAL, vHash uint64) {
	o.mu.Lock()
	defer o.mu.Unlock()

	fv := &clustering.FeatureVector{
		ID:         sal.ScanID,
		VisualHash: vHash,
		IP:         extractIP(sal),                // Helper to get IP from SAL entities
		DOMTokens:  []string{"login", "password"}, // Mock tokens for now, needs Text Pipeline output
	}
	o.clusterBuffer = append(o.clusterBuffer, fv)

	if len(o.clusterBuffer) >= o.cfg.ClusterBatchSize {
		o.runClustering()
	}
}

func (o *Orchestrator) runClustering() {
	o.logger.Info("running DBSCAN clustering", "batch_size", len(o.clusterBuffer))
	clusters := clustering.RunDBSCAN(o.clusterBuffer)

	for _, c := range clusters {
		o.logger.Info("campaign cluster identified", "cluster_id", c.ID, "size", len(c.Points))
		// Log points in cluster
		ids := make([]string, len(c.Points))
		for i, p := range c.Points {
			ids[i] = p.ID
		}
		o.logger.Info("cluster members", "ids", ids)
	}

	// Reset buffer
	o.clusterBuffer = o.clusterBuffer[:0]
}

func extractIP(sal domain.SAL) string {
	for _, e := range sal.Entities {
		if e.Type == "IP" {
			return e.Value
		}
	}
	return ""
}

// Close gracefully shuts down the orchestrator and its components.
func (o *Orchestrator) Close() {
	if o.graphProjector != nil {
		o.graphProjector.Close()
	}
}
