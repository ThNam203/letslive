package minio

import (
	"context"
	"fmt"
	"io"
	"os"
	"sen1or/letslive/livestream/config"
	"sen1or/letslive/livestream/pkg/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func getPolicy(bucketName string) string {
	return fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
		    "Effect": "Allow",
		    "Principal": "*",
		    "Action": ["s3:GetObject"],
		    "Resource": ["arn:aws:s3:::%s/*"]
		  }
		]
	}`, bucketName)
}

type MinIOStorage struct {
	client *minio.Client
	config config.MinIO
}

func NewMinIOStorage(ctx context.Context, cfg config.MinIO) *MinIOStorage {
	minioClient, err := minio.New(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD"), ""),
		Secure: false,
	})
	if err != nil {
		logger.Panicf(ctx, "failed to initialize MinIO client: %s", err)
	}

	// Ensure bucket exists
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		logger.Panicf(ctx, "failed to check bucket: %v", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			logger.Panicf(ctx, "failed to create bucket: %v", err)
		}
		err = minioClient.SetBucketPolicy(ctx, cfg.BucketName, getPolicy(cfg.BucketName))
		if err != nil {
			logger.Panicf(ctx, "failed to set bucket policy: %v", err)
		}
	}

	return &MinIOStorage{
		client: minioClient,
		config: cfg,
	}
}

// UploadFile uploads a file to MinIO and returns the object path
func (s *MinIOStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, fileSize int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.config.BucketName, objectName, reader, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to minio: %v", err)
	}

	finalURL := fmt.Sprintf("%s/%s/%s", s.config.ReturnURL, s.config.BucketName, objectName)
	return finalURL, nil
}

// DeleteFile removes a file from MinIO
func (s *MinIOStorage) DeleteFile(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.config.BucketName, objectName, minio.RemoveObjectOptions{})
}

// GetFile downloads a file from MinIO
func (s *MinIOStorage) GetFile(ctx context.Context, objectName string) (*minio.Object, error) {
	return s.client.GetObject(ctx, s.config.BucketName, objectName, minio.GetObjectOptions{})
}
