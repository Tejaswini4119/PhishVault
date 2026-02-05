package engine

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"regexp"
	"time"
)

type AttachmentProcessor struct{}

func NewAttachmentProcessor() *AttachmentProcessor {
	return &AttachmentProcessor{}
}

func (p *AttachmentProcessor) Process(ctx context.Context, input []byte, sourceID string, metadata map[string]interface{}) (*IngestedArtifact, error) {
	// 1. Calculate Hashes
	sha256Hash := sha256.Sum256(input)
	md5Hash := md5.Sum(input)

	// 2. Detect Content Type
	contentType := http.DetectContentType(input)
	if metadata != nil && metadata["content_type"] != nil {
		// Prefer metadata type if specific (e.g. from email header) but sanitize
		ct := metadata["content_type"].(string)
		if ct != "" {
			contentType = ct
		}
	}

	// 3. Extract URLs (Basic Regex for text-based files or if we want to scan binary strings)
	// Limiting scan to first 1MB for performance if large
	scanLimit := 1024 * 1024
	scanData := input
	if len(input) > scanLimit {
		scanData = input[:scanLimit]
	}

	urls := extractURLs(string(scanData))

	// 4. Construct Artifact
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["sha256"] = hex.EncodeToString(sha256Hash[:])
	metadata["md5"] = hex.EncodeToString(md5Hash[:])
	metadata["detected_content_type"] = contentType
	metadata["extracted_urls_count"] = len(urls)

	children := make([]*IngestedArtifact, 0)
	for _, u := range urls {
		children = append(children, &IngestedArtifact{
			Type:            ArtifactTypeURL,
			Content:         u,
			SourceID:        sourceID,
			Timestamp:       time.Now(),
			IngestionSource: "ATTACHMENT_EXTRACTION",
		})
	}

	return &IngestedArtifact{
		Type:            ArtifactTypeFile,
		SourceID:        sourceID,
		Metadata:        metadata,
		Timestamp:       time.Now(),
		Content:         "", // Don't store binary here, expected to be in RawPath/Storage
		Children:        children,
		IngestionSource: "ATTACHMENT_PROCESSOR",
	}, nil
}

// Regex for URL extraction (Basic)
var urlRegex = regexp.MustCompile(`https?://[a-zA-Z0-9\-\._~:/?#\[\]@!$&'()*+,;=]+`)

func extractURLs(text string) []string {
	return urlRegex.FindAllString(text, -1)
}
