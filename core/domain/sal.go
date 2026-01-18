package domain

import "time"

// SAL (Signal Abstraction Layer) defines the unified data structure for all scan results.
// This serves as the contract between Ingestion, Scanning, and Analysis engines.
// Conforms to PhishVault-2 Technical Workbook Section 5.14.
type SAL struct {
	// Identity
	ArtifactID string    `json:"artifact_id"` // Unique UUID for the artifact
	ScanID     string    `json:"scan_id"`
	Timestamp  time.Time `json:"timestamp"`

	// Input
	URL             string `json:"url"`
	IngestionSource string `json:"ingestion_source"` // e.g. "API", "Feed", "Email"

	// Temporal & Network State
	RedirectChain []string `json:"redirect_chain,omitempty"` // Full hop trace
	FinalURL      string   `json:"final_url,omitempty"`

	// Captured Data (The "Ground Truth")
	Request   RequestDetails  `json:"request,omitempty"`
	Response  ResponseDetails `json:"response,omitempty"`
	Artifacts Artifacts       `json:"artifacts,omitempty"`

	// Intelligence Layer (Phase 2 & 3)
	Signals    []Signal `json:"signals,omitempty"`     // Section 5.14: Discrete Engine Outputs
	Entities   []Entity `json:"entities,omitempty"`    // Extracted Brands, Emails, etc.
	CampaignID string   `json:"campaign_id,omitempty"` // From Phase 3 Clustering

	// Decision
	Verdict   string                 `json:"verdict,omitempty"`
	RiskScore float64                `json:"risk_score,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"` // Flexible storage
}

// Signal represents a discrete unit of intelligence from an analysis engine.
type Signal struct {
	EngineName string                 `json:"engine_name"`
	SignalKey  string                 `json:"signal_key"` // e.g. "VISUAL_MATCH", "JS_OBFUSCATION"
	Confidence float64                `json:"confidence"` // 0.0 - 1.0 (Engine's self-assessment)
	Weight     float64                `json:"weight"`     // Trust Model weight
	Evidence   map[string]interface{} `json:"evidence"`   // Snippets, line numbers, etc.
	Tags       []string               `json:"tags,omitempty"`
}

// Entity represents a semantic object extracted from the artifact.
type Entity struct {
	Type   string `json:"type"` // "Brand", "Credential", "Email", "ASN"
	Value  string `json:"value"`
	Source string `json:"source"` // Which engine found it
}

// RequestDetails holds information about the HTTP request made
type RequestDetails struct {
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
}

// ResponseDetails holds information about the HTTP response received
type ResponseDetails struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	BodyHash   string            `json:"body_hash"` // SHA256 hash of the content
	IP         string            `json:"ip,omitempty"`
	ASN        string            `json:"asn,omitempty"`
}

// Artifacts holds references to stored files
type Artifacts struct {
	ScreenshotPath string `json:"screenshot_path,omitempty"` // Path or URL to screenshot
	DOMPath        string `json:"dom_path,omitempty"`        // Path or URL to stored DOM
	RawContent     string `json:"raw_content,omitempty"`     // In-memory content for immediate analysis
	VisualHash     string `json:"visual_hash,omitempty"`     // pHash/dHash
}
