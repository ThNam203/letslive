package minio

import (
	"context"
	"fmt"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TODO: make config
const (
	endpoint   = "minio:9000"
	uiEndpoint = "http://localhost:9000"
	bucketName = "livestreams"
	useSSL     = false
	accessKey  = "minioadmin"
	secretKey  = "minioadmin"
)

var policy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::livestreams/*"]
    }
  ]
}`

type MinIOStrorage struct {
	minioClient *minio.Client
	ctx         context.Context
}

// If we don't want to connect to bootstrap node, enter a nil value for bootstrapNodeAddr
func NewMinIOStorage(ctx context.Context) *MinIOStrorage {
	storage := &MinIOStrorage{
		ctx: ctx,
	}
	if err := storage.SetUp(); err != nil {
		logger.Panicf("error setting up node: %s", err)
	}

	return storage
}

func (s *MinIOStrorage) SetUp() error {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize MinIO client: %s", err)
	}

	s.minioClient = minioClient

	exists, err := minioClient.BucketExists(s.ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %v", err)
	}
	if !exists {
		err = minioClient.MakeBucket(s.ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}

		err = minioClient.SetBucketPolicy(s.ctx, bucketName, policy)
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %v", err)
		}
	}

	return nil
}

// uploads a file to MinIO and returns the permanent URL
func (s *MinIOStrorage) AddFile(filePath string, streamId string) (string, error) {
	fileName := filepath.Base(filePath)
	objectName := fmt.Sprintf("%s/%s", streamId, fileName)

	// Upload the file
	// TODO: write config please
	_, err := s.minioClient.FPutObject(context.Background(), bucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType:  "video/mp2t",
		CacheControl: "max-age=3600",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to minio: %v", err)
	}

	// Construct the final URL (public access)
	finalURL := fmt.Sprintf("%s/%s/%s", uiEndpoint, bucketName, objectName)

	return finalURL, nil
}
