package analysis

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"net/url"
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

	// 1. Visual Analysis & Golden Set
	// Parse VisualHash from input (computed by Scanner) or default to 0 if missing.
	var visualHash uint64
	if input.Artifacts.VisualHash != "" {
		if val, err := strconv.ParseUint(input.Artifacts.VisualHash, 16, 64); err == nil {
			visualHash = val
		} else {
			o.logger.Warn("invalid visual hash format", "hash", input.Artifacts.VisualHash, "error", err)
		}
	}

	// Check Golden Set if we have a valid hash
	var similarity float64
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
	// AnalyzeContent returns detailed risk struct.
	htmlContent := input.Artifacts.RawContent
	if htmlContent == "" {
		htmlContent = "<html></html>" // Fallback if missing
	}
	risk := ai.AnalyzeContent(htmlContent, "Verify your account", input.FinalURL) // "Verify your account" should ideally come from Text Pipeline extraction

	if risk.Intent != "Benign" {
		o.logger.Info("malicious intent detected", "intent", risk.Intent, "urgency", risk.UrgencyScore)
		sig := domain.Signal{
			EngineName: "NLP_Deep",
			SignalKey:  "INTENT_" + risk.Intent,
			Confidence: risk.UrgencyScore, // Using urgency as proxy
			Weight:     0.8,
			Tags:       []string{risk.Intent},
		}
		input.Signals = append(input.Signals, sig)
	}

	if risk.FormRisk.HasPassword {
		input.Signals = append(input.Signals, domain.Signal{
			EngineName: "Structure",
			SignalKey:  "SENSITIVE_FORM",
			Weight:     0.5,
		})
	}

	// 2.5 Threat Intelligence (Moved before Verdict)
	// Extract Domain
	u, err := url.Parse(input.FinalURL)
	var domainName string
	if err == nil {
		domainName = u.Hostname()
	} else {
		domainName = input.FinalURL // Fallback
	}

	// a. WHOIS Enrichment
	var domainAgeDays int
	whoisData, err := provider.FetchWHOIS(domainName)
	if err == nil {
		age := time.Since(whoisData.CreationDate).Hours() / 24
		domainAgeDays = int(age)

		// Signal for young domains
		if domainAgeDays < 30 {
			o.logger.Info("young domain detected", "age_days", domainAgeDays)
			input.Signals = append(input.Signals, domain.Signal{
				EngineName: "ThreatIntel",
				SignalKey:  "YOUNG_DOMAIN",
				Confidence: 1.0,
				Weight:     0.6,
				Evidence:   map[string]interface{}{"age_days": domainAgeDays},
			})
		}
	}

	// b. Reputation Check
	repResults, _ := provider.CheckReputation(domainName)
	for _, res := range repResults {
		if res.Malicious {
			o.logger.Warn("threat feed hit", "source", res.Source, "score", res.Score)
			input.Signals = append(input.Signals, domain.Signal{
				EngineName: "ThreatIntel",
				SignalKey:  "KNOWN_MALICIOUS",
				Confidence: res.Score,
				Weight:     1.0, // High certainty
				Evidence:   map[string]interface{}{"source": res.Source},
			})
		}
	}

	// 3. Verdict (OPA)
	opaInput := decision.PolicyInput{
		VisualMatchScore: similarity,
		UrgencyScore:     risk.UrgencyScore,
		Intent:           risk.Intent,
		HasLoginForm:     risk.FormRisk.HasPassword,
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
