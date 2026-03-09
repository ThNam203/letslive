package vod

import (
	"sen1or/letslive/livestream/domains"
	miniostorage "sen1or/letslive/livestream/storage/minio"
)

type VODService struct {
	vodRepo          domains.VODRepository
	transcodeJobRepo domains.TranscodeJobRepository
	minioStorage     *miniostorage.MinIOStorage
}

func NewVODService(vodRepo domains.VODRepository, transcodeJobRepo domains.TranscodeJobRepository, minioStorage *miniostorage.MinIOStorage) *VODService {
	return &VODService{
		vodRepo:          vodRepo,
		transcodeJobRepo: transcodeJobRepo,
		minioStorage:     minioStorage,
	}
}
