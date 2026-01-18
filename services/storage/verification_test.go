package storage

import (
	"context"
	"testing"
)

// TestMinIOConnection simulates the application connecting to MinIO.
// It uses the default docker-compose credentials.
func TestMinIOConnection(t *testing.T) {
	// 1. Config
	endpoint := "localhost:9000"
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	bucketName := "phishvault-verification"

	t.Logf("Connecting to MinIO at %s...", endpoint)

	// 2. Initialize Manager (this connects and creates bucket)
	sm, err := NewStorageManager(endpoint, accessKey, secretKey, bucketName)
	if err != nil {
		t.Fatalf("Failed to connect to MinIO: %v", err)
	}

	t.Logf("Successfully connected to MinIO and accessed bucket '%s'.", bucketName)

	// 3. Test Upload (Small text file)
	ctx := context.Background()

	// We need a file on disk for UploadFile?
	// The current implementation of UploadFile (manager.go isn't fully using Use PutObject directly to avoid temp file for test if possible,
	// but StorageManager.UploadFile takes a path.
	// However, manager.go has SaveDOM which takes content string. Let's use that.

	// Wait, manager.go methods like SaveDOM are on *StorageManager.
	// Let's use SaveDOM just to test upload capability without disk I/O.

	objectName, err := sm.SaveDOM(ctx, "verify-scan-id", "Hello MinIO Verification")
	if err != nil {
		t.Fatalf("Failed to upload verification object: %v", err)
	}

	t.Logf("Successfully uploaded object: %s", objectName)
}
