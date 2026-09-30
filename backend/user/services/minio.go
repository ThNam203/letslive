package services

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/config"

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

const (
	ProfilePicturesBucket         = "profile-pictures"
	ProfilePicturesOriginalBucket = "profile-pictures-original"
)

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
		logger.Panicf(ctx, "error setting up minio storage: %s", err)
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

	if err := s.createIfNotExists("general-files", true); err != nil {
		return err
	}

	// TODO: remove all these, use general-files instead
	if err := s.createIfNotExists(ProfilePicturesBucket, true); err != nil {
		return err
	}
	if err := s.createIfNotExists("thumbnails", true); err != nil {
		return err
	}
	if err := s.createIfNotExists("background-pictures", true); err != nil {
		return err
	}

	if err := s.createIfNotExists(ProfilePicturesOriginalBucket, false); err != nil {
		return err
	}

	return nil
}

func (s *MinIOService) createIfNotExists(bucketName string, publicRead bool) error {
	exists, err := s.minioClient.BucketExists(s.ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %v", err)
	}
	if !exists {
		err = s.minioClient.MakeBucket(s.ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %v", err)
		}

		if !publicRead {
			return nil
		}

		err = s.minioClient.SetBucketPolicy(s.ctx, bucketName, getPolicy(bucketName))
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %v", err)
		}
	}

	return nil
}

func (s *MinIOService) PutObject(ctx context.Context, bucketName, objectName string, data []byte, contentType string) error {
	_, err := s.minioClient.PutObject(ctx, bucketName, objectName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "max-age=86400",
	})
	if err != nil {
		return fmt.Errorf("failed to upload %s/%s to minio: %w", bucketName, objectName, err)
	}

	return nil
}

func (s *MinIOService) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	if err := s.minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to remove %s/%s from minio: %w", bucketName, objectName, err)
	}

	return nil
}

func (s *MinIOService) PublicURL(bucketName, objectName string) string {
	return fmt.Sprintf("%s/%s/%s", s.config.ReturnURL, bucketName, objectName)
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

	return s.PublicURL(bucketName, fileName), nil
}
