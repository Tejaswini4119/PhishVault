package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"time"

	"github.com/PhishVault/PhishVault-2/services/ingestion/parser"
)

type EmailProcessor struct {
	factory *ArtifactFactory
}

func NewEmailProcessor() *EmailProcessor {
	return &EmailProcessor{}
}

func (p *EmailProcessor) SetFactory(f *ArtifactFactory) {
	p.factory = f
}

func (p *EmailProcessor) Process(ctx context.Context, input []byte, sourceID string, metadata map[string]interface{}) (*IngestedArtifact, error) {
	reader := bytes.NewReader(input)
	msg, err := mail.ReadMessage(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	// 1. Extract Headers
	headers := make(map[string]interface{})
	for k, v := range msg.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 2. Authentication Analysis (SPF/DKIM) - Reusing existing logic
	authHeaders := make(map[string]string)
	for k, v := range headers {
		authHeaders[k] = v.(string)
	}
	spfResult := parser.AnalyzeSPF(authHeaders)
	headers["auth_spf_verdict"] = spfResult

	// 2b. DKIM Verification (Independent)
	dkimResult, err := parser.VerifyDKIM(bytes.NewReader(input))
	if err != nil {
		headers["auth_dkim_error"] = err.Error()
	}
	headers["auth_dkim_independent"] = dkimResult

	artifact := &IngestedArtifact{
		Type:            ArtifactTypeEmail,
		SourceID:        sourceID,
		Metadata:        headers,
		Timestamp:       time.Now(),
		IngestionSource: "EMAIL_PARSER",
		Children:        make([]*IngestedArtifact, 0),
	}

	// 3. Deconstruct Body and Attachments
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		// Fallback: treat basic text/plain or unknown as simple body
		body, _ := io.ReadAll(msg.Body)
		artifact.Content = string(body)
	} else if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("reading multipart failed: %w", err)
			}

			// Process Part
			childArtifact, err := p.processPart(ctx, part, sourceID)
			if err == nil && childArtifact != nil {
				if childArtifact.Type == ArtifactTypeFile {
					artifact.Children = append(artifact.Children, childArtifact)
				} else {
					// Merge text bodies into main content or handle logic
					artifact.Content += childArtifact.Content + "\n"
				}
			}
		}
	} else {
		// Single part (text/html or text/plain)
		body, _ := io.ReadAll(msg.Body)
		artifact.Content = string(body)
	}

	if metadata != nil {
		for k, v := range metadata {
			artifact.Metadata[k] = v
		}
	}

	return artifact, nil
}

func (p *EmailProcessor) processPart(ctx context.Context, part *multipart.Part, sourceID string) (*IngestedArtifact, error) {
	filename := part.FileName()
	contentType := part.Header.Get("Content-Type")

	data, err := io.ReadAll(part)
	if err != nil {
		return nil, err
	}

	// If it has a filename, treat as attachment
	if filename != "" {
		// Prepare metadata
		meta := map[string]interface{}{"filename": filename, "content_type": contentType, "size": len(data)}

		// Recursion hook: if we have a factory, use it to fully ingest the attachment (which might be a zip, etc)
		if p.factory != nil {
			return p.factory.Ingest(ctx, ArtifactTypeFile, data, sourceID, meta)
		}

		// Fallback for no factory (unit tests?)
		return &IngestedArtifact{
			Type:            ArtifactTypeFile,
			Content:         "",
			Metadata:        meta,
			SourceID:        sourceID,
			Timestamp:       time.Now(),
			IngestionSource: "EMAIL_ATTACHMENT",
		}, nil
	}

	// Handle Nested Emails (Forwarded messages, usually "message/rfc822")
	if strings.Contains(contentType, "message/rfc822") {
		if p.factory != nil {
			return p.factory.Ingest(ctx, ArtifactTypeEmail, data, sourceID, map[string]interface{}{"parent_type": "email_nested"})
		}
	}

	// Otherwise treat as body content
	return &IngestedArtifact{
		Type:     "BODY_PART", // Internal type, processed into parent
		Content:  string(data),
		Metadata: map[string]interface{}{"content_type": contentType},
	}, nil
}
