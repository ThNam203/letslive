package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/config"
	"sen1or/letslive/user/domains"

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
	generalFilesBucket       = "general-files"
	profilePicturesBucket    = "profile-pictures"
	thumbnailsBucket         = "thumbnails"
	backgroundPicturesBucket = "background-pictures"
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

	if err := s.createIfNotExists(generalFilesBucket); err != nil {
		return err
	}

	// TODO: remove all these, use general-files instead
	if err := s.createIfNotExists(profilePicturesBucket); err != nil {
		return err
	}
	if err := s.createIfNotExists(thumbnailsBucket); err != nil {
		return err
	}
	if err := s.createIfNotExists(backgroundPicturesBucket); err != nil {
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

func (s *MinIOService) addImage(ctx context.Context, r io.Reader, bucketName string, rule imageRule) (string, error) {
	upload, err := readImage(r, rule)
	if err != nil {
		return "", err
	}
	stored, err := toStoredImage(upload)
	if err != nil {
		return "", err
	}

	objectName := uuid.NewString() + "." + stored.format.ext
	_, err = s.minioClient.PutObject(ctx, bucketName, objectName, bytes.NewReader(stored.data), int64(len(stored.data)), minio.PutObjectOptions{
		ContentType:  stored.format.contentType,
		CacheControl: "max-age=86400",
	})
	if err != nil {
		return "", fmt.Errorf("store %s/%s: %w: %w", bucketName, objectName, domains.ErrInternal, err)
	}

	return fmt.Sprintf("%s/%s/%s", s.config.ReturnURL, bucketName, objectName), nil
}

func (s *MinIOService) AddLivestreamThumbnail(ctx context.Context, r io.Reader) (string, error) {
	return s.addImage(ctx, r, thumbnailsBucket, livestreamThumbnailRule)
}
