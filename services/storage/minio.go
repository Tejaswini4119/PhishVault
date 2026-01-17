package storage

import (
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageManager struct {
	Client     *minio.Client
	BucketName string
}

func NewStorageManager(endpoint, accessKey, secretKey, bucketName string) (*StorageManager, error) {
	ctx := context.Background()
	useSSL := false

	// Initialize MinIO client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// Make a bucket if it doesn't exist
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
		log.Printf("Successfully created bucket %s\n", bucketName)
	}

	return &StorageManager{
		Client:     minioClient,
		BucketName: bucketName,
	}, nil
}

// UploadFile uploads a file to MinIO and returns the object name
func (sm *StorageManager) UploadFile(ctx context.Context, objectName string, filePath string, contentType string) error {
	_, err := sm.Client.FPutObject(ctx, sm.BucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	return err
}
