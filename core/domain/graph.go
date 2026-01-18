package domain

// Graph Node Labels
const (
	LabelDomain      = "Domain"
	LabelIP          = "IP"
	LabelASN         = "ASN"
	LabelCertificate = "Certificate"
	LabelCampaign    = "Campaign"
	LabelURL         = "URL"
)

// Graph Edge Types
const (
	EdgeResolvesTo = "RESOLVES_TO" // Domain -> IP
	EdgeHostedOn   = "HOSTED_ON"   // URL -> Domain
	EdgeIssuedBy   = "ISSUED_BY"   // Domain -> Certificate or Certificate -> Authority
	EdgePartOf     = "PART_OF"     // IP -> Campaign
	EdgeBelongsTo  = "BELONGS_TO"  // IP -> ASN
)

// GraphNode represents a generic graph node for use by the Projector.
type GraphNode struct {
	Label      string                 `json:"label"`
	Key        string                 `json:"key"` // Unique Key (e.g., domain name, ip address)
	Properties map[string]interface{} `json:"properties"`
}

// GraphEdge represents a relationship between two nodes.
type GraphEdge struct {
	SourceType string                 `json:"source_type"` // e.g., LabelDomain
	SourceKey  string                 `json:"source_key"`
	TargetType string                 `json:"target_type"` // e.g., LabelIP
	TargetKey  string                 `json:"target_key"`
	Relation   string                 `json:"relation"` // e.g., EdgeResolvesTo
	Properties map[string]interface{} `json:"properties"`
}

// GraphDatabase defines the contract for persisting graph structures.
type GraphDatabase interface {
	ExecuteBatch(nodes []GraphNode, edges []GraphEdge) error
	Close() error
}
