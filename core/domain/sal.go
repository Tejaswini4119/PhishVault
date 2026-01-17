package domain

import "time"

// SAL (Signal Abstraction Layer) defines the unified data structure for all scan results.
// This serves as the contract between Ingestion, Scanning, and Analysis engines.
type SAL struct {
	ScanID          string    `json:"scan_id"`
	URL             string    `json:"url"`
	Timestamp       time.Time `json:"timestamp"`
	IngestionSource string    `json:"ingestion_source"`

	// Request details captured during the scan
	Request RequestDetails `json:"request,omitempty"`

	// Response details captured from the target
	Response ResponseDetails `json:"response,omitempty"`

	// Artifacts generated during the scan
	Artifacts Artifacts `json:"artifacts,omitempty"`

	// Analysis Phase placeholders (to be populated in Phase 2)
	Verdict   string  `json:"verdict,omitempty"`
	RiskScore float64 `json:"risk_score,omitempty"`
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
	FinalURL   string            `json:"final_url"` // URL after redirect chains
	IP         string            `json:"ip,omitempty"`
	ASN        string            `json:"asn,omitempty"`
}

// Artifacts holds references to stored files
type Artifacts struct {
	ScreenshotPath string `json:"screenshot_path,omitempty"` // Path or URL to screenshot
	DOMPath        string `json:"dom_path,omitempty"`        // Path or URL to stored DOM
}
