package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"sen1or/letslive/user/config"
	"sen1or/letslive/user/pkg/logger"

	"github.com/google/uuid"
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

type MinIOService struct {
	minioClient *minio.Client
	ctx         context.Context
	config      config.MinIO
}

// If we don't want to connect to bootstrap node, enter a nil value for bootstrapNodeAddr
func NewMinIOService(ctx context.Context, config config.MinIO) *MinIOService {
	storage := &MinIOService{
		ctx:    ctx,
		config: config,
	}

	if err := storage.SetUp(); err != nil {
		logger.Panicf("error setting up minio storage: %s", err)
	}

	return storage
}

func (s *MinIOService) SetUp() error {
	minioClient, err := minio.New(fmt.Sprintf("%s:%d", s.config.Host, s.config.Port), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD"), ""),
		Secure: false,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize MinIO client: %s", err)
	}

	s.minioClient = minioClient

	if err := s.createIfNotExists("general-files"); err != nil {
		return err
	}

	// TODO: remove all these, use general-files instead
	if err := s.createIfNotExists("profile-pictures"); err != nil {
		return err
	}
	if err := s.createIfNotExists("thumbnails"); err != nil {
		return err
	}
	if err := s.createIfNotExists("background-pictures"); err != nil {
		return err
	}

	return nil
}

func (s *MinIOService) createIfNotExists(bucketName string) error {
	exists, err := s.minioClient.BucketExists(s.ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %v", err)
	}
	if !exists {
		err = s.minioClient.MakeBucket(s.ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}

		err = s.minioClient.SetBucketPolicy(s.ctx, bucketName, getPolicy(bucketName))
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %v", err)
		}
	}

	return nil
}

// uploads a file to MinIO and returns the permanent URL
func (s *MinIOService) AddFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, bucketName string) (string, error) {
	fileName := fmt.Sprintf("%s-%s", uuid.New().String(), fileHeader.Filename)

	// Upload the file
	_, err := s.minioClient.PutObject(ctx, bucketName, fileName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType:  fileHeader.Header.Get("Content-Type"),
		CacheControl: "max-age=86400",
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to minio: %v", err)
	}

	// Construct the final URL (public access)
	finalURL := fmt.Sprintf("%s/%s/%s", s.config.ReturnURL, bucketName, fileName)

	return finalURL, nil
}
