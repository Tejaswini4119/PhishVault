package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type AttachmentProcessor struct {
	factory *ArtifactFactory
}

func NewAttachmentProcessor() *AttachmentProcessor {
	return &AttachmentProcessor{}
}

func (p *AttachmentProcessor) SetFactory(f *ArtifactFactory) {
	p.factory = f
}

func (p *AttachmentProcessor) Process(ctx context.Context, input []byte, sourceID string, metadata map[string]interface{}) (*IngestedArtifact, error) {
	// 1. Calculate Hashes
	sha256Hash := sha256.Sum256(input)
	md5Hash := md5.Sum(input)

	var children []*IngestedArtifact

	// 2. Detect Content Type
	contentType := http.DetectContentType(input)
	if metadata != nil && metadata["content_type"] != nil {
		// Prefer metadata type if specific (e.g. from email header) but sanitize
		ct := metadata["content_type"].(string)
		if ct != "" {
			contentType = ct
		}
	}

	// 3. Recursive Processing (Zip/Archive)
	contentType = normalizeContentType(contentType, metadata)

	// Check for Archives
	if isArchive(contentType) {
		// Use injected factory to process children
		childArtifacts, err := p.processArchive(ctx, input, sourceID, contentType)
		if err == nil {
			children = append(children, childArtifacts...)
		} else {
			// Log error but continue as file
			metadata["processing_error"] = fmt.Sprintf("Archive extraction failed: %v", err)
		}
	}

	// 4. Extract URLs (Basic Regex)
	// Limiting scan to first 1MB for performance if large
	scanLimit := 1024 * 1024
	scanData := input
	if len(input) > scanLimit {
		scanData = input[:scanLimit]
	}

	urls := extractURLs(string(scanData))
	for _, u := range urls {
		children = append(children, &IngestedArtifact{
			Type:            ArtifactTypeURL,
			Content:         u,
			SourceID:        sourceID,
			Timestamp:       time.Now(),
			IngestionSource: "ATTACHMENT_EXTRACTION",
		})
	}

	// 5. Construct Artifact
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["sha256"] = hex.EncodeToString(sha256Hash[:])
	metadata["md5"] = hex.EncodeToString(md5Hash[:])
	metadata["detected_content_type"] = contentType
	metadata["extracted_urls_count"] = len(urls)

	return &IngestedArtifact{
		Type:            ArtifactTypeFile,
		SourceID:        sourceID,
		Metadata:        metadata,
		Timestamp:       time.Now(),
		Content:         "", // Don't store binary here
		Children:        children,
		IngestionSource: "ATTACHMENT_PROCESSOR",
	}, nil
}

func normalizeContentType(detected string, metadata map[string]interface{}) string {
	if metadata != nil && metadata["content_type"] != nil {
		ct := metadata["content_type"].(string)
		if ct != "" {
			return ct
		}
	}
	return detected
}

func isArchive(ct string) bool {
	return strings.Contains(ct, "zip") || strings.Contains(ct, "compressed")
}

func (p *AttachmentProcessor) processArchive(ctx context.Context, input []byte, sourceID string, contentType string) ([]*IngestedArtifact, error) {
	if p.factory == nil {
		return nil, nil // No factory, no recursion
	}

	var children []*IngestedArtifact

	// Only supporting Zip for now
	if strings.Contains(contentType, "zip") {
		r, err := zip.NewReader(bytes.NewReader(input), int64(len(input)))
		if err != nil {
			return nil, err
		}

		// Security Caps
		maxFiles := 10
		maxSize := 10 * 1024 * 1024 // 10MB

		for i, f := range r.File {
			if i >= maxFiles {
				break
			}
			if f.FileInfo().IsDir() {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				continue
			}

			// Read with limit
			content, err := io.ReadAll(io.LimitReader(rc, int64(maxSize)))
			rc.Close()
			if err != nil {
				continue
			}

			// RECURSIVE CALL
			childMeta := map[string]interface{}{
				"filename":            f.Name,
				"parent_archive_type": "zip",
			}

			child, err := p.factory.Ingest(ctx, ArtifactTypeFile, content, sourceID, childMeta)
			if err == nil {
				children = append(children, child)
			}
		}
	}
	return children, nil
}

// Regex for URL extraction (Basic)
var urlRegex = regexp.MustCompile(`https?://[a-zA-Z0-9\-\._~:/?#\[\]@!$&'()*+,;=]+`)

func extractURLs(text string) []string {
	return urlRegex.FindAllString(text, -1)
}
