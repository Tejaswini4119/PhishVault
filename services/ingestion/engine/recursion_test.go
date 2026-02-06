package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"
)

func TestRecursiveZipIngestion(t *testing.T) {
	// 1. Create a Zip file in memory containing "test.txt"
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.Create("payload.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("malicious_code"))
	if err != nil {
		t.Fatal(err)
	}
	w.Close()

	zipBytes := buf.Bytes()

	// 2. Setup Factory
	factory := NewArtifactFactory()

	// Register Processors
	attProc := NewAttachmentProcessor()
	factory.RegisterProcessor(ArtifactTypeFile, attProc)

	// Note: Generic "Process" for .txt files will also use AttachmentProcessor (or we default to it)
	// In the real app, we might have different processors, but here AttachmentProcessor handles standard files too.

	// 3. Ingest the Zip
	ctx := context.Background()
	artifact, err := factory.Ingest(ctx, ArtifactTypeFile, zipBytes, "TEST-Zip", map[string]interface{}{
		"content_type": "application/zip",
	})
	if err != nil {
		t.Fatalf("Ingestion failed: %v", err)
	}

	// 4. Verify Recursion
	// Expect:
	// - Parent Artifact (Zip)
	// - Child Artifact (payload.txt)

	if len(artifact.Children) < 1 {
		t.Fatal("Expected children extracted from zip, got 0")
	}

	child := artifact.Children[0]
	if child.Metadata["filename"] != "payload.txt" {
		t.Errorf("Expected filename payload.txt, got %v", child.Metadata["filename"])
	}

	// Note: AttachmentProcessor currently sets Content="" for files (binary),
	// but if it recursed, the Child (TypeFile) should have been processed.
	// In AttachmentProcessor.Process, it sets Content="" for the artifact it returns.
	// But during recursion, we call factory.Ingest -> AttachmentProcessor.Process again for the text file.
	// So child.Content will be "" (because it's treated as file).
	// But it might have sub-children if we extracted URLs?
	// "malicious_code" is not a URL.

	// Check metadata source
	if child.IngestionSource != "ATTACHMENT_PROCESSOR" {
		t.Errorf("Expected source ATTACHMENT_PROCESSOR, got %s", child.IngestionSource)
	}
}
