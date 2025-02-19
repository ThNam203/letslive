package minio

import (
	"context"
	"fmt"
	"path/filepath"
	"sen1or/lets-live/pkg/logger"
	"sen1or/lets-live/transcode/config"

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

type MinIOStrorage struct {
	minioClient *minio.Client
	ctx         context.Context
	config      config.MinIO
}

// If we don't want to connect to bootstrap node, enter a nil value for bootstrapNodeAddr
func NewMinIOStorage(ctx context.Context, config config.MinIO) *MinIOStrorage {
	storage := &MinIOStrorage{
		ctx:    ctx,
		config: config,
	}

	if err := storage.SetUp(); err != nil {
		logger.Panicf("error setting up node: %s", err)
	}

	return storage
}

func (s *MinIOStrorage) SetUp() error {
	minioClient, err := minio.New(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port), &minio.Options{
		Creds:  credentials.NewStaticV4(s.config.AccessKey, s.config.SecretKey, ""),
		Secure: false,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize MinIO client: %s", err)
	}

	s.minioClient = minioClient

	exists, err := minioClient.BucketExists(s.ctx, s.config.BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %v", err)
	}
	if !exists {
		err = minioClient.MakeBucket(s.ctx, s.config.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}

		err = minioClient.SetBucketPolicy(s.ctx, s.config.BucketName, getPolicy(s.config.BucketName))
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
	_, err := s.minioClient.FPutObject(context.Background(), s.config.BucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType:  "video/mp2t",
		CacheControl: "max-age=3600",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to minio: %v", err)
	}

	// Construct the final URL (public access)
	finalURL := fmt.Sprintf("%s:%d/%s/%s", s.config.ClientHost, s.config.Port, s.config.BucketName, objectName)

	return finalURL, nil
}
