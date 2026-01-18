package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

// SaveScreenshot uploads a screenshot to MinIO.
// Returns the object path or error.
func (sm *StorageManager) SaveScreenshot(ctx context.Context, scanID string, data []byte) (string, error) {
	objectName := fmt.Sprintf("scans/%s/screenshot.png", scanID)

	reader := bytes.NewReader(data)
	_, err := sm.Client.PutObject(ctx, sm.BucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "image/png",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload screenshot: %w", err)
	}

	return objectName, nil
}

// SaveDOM uploads the HTML content to MinIO.
func (sm *StorageManager) SaveDOM(ctx context.Context, scanID string, content string) (string, error) {
	objectName := fmt.Sprintf("scans/%s/dom.html", scanID)
	data := []byte(content)

	reader := bytes.NewReader(data)
	_, err := sm.Client.PutObject(ctx, sm.BucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: "text/html",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload DOM: %w", err)
	}

	return objectName, nil
}

// SaveRaw uploads any raw artifact (e.g. .eml, .pdf).
func (sm *StorageManager) SaveRaw(ctx context.Context, scanID string, fileName string, data []byte, contentType string) (string, error) {
	objectName := fmt.Sprintf("scans/%s/raw/%s", scanID, fileName)

	reader := bytes.NewReader(data)
	_, err := sm.Client.PutObject(ctx, sm.BucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload raw artifact: %w", err)
	}

	return objectName, nil
}
