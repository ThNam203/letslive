package vod

import (
	"sen1or/letslive/vod/domains"
	usergateway "sen1or/letslive/vod/gateway/user"
	miniostorage "sen1or/letslive/vod/storage/minio"
)

type VODService struct {
	vodRepo          domains.VODRepository
	transcodeJobRepo domains.TranscodeJobRepository
	minioStorage     *miniostorage.MinIOStorage
	userGateway      usergateway.UserGateway
}

func NewVODService(vodRepo domains.VODRepository, transcodeJobRepo domains.TranscodeJobRepository, minioStorage *miniostorage.MinIOStorage, userGateway usergateway.UserGateway) *VODService {
	return &VODService{
		vodRepo:          vodRepo,
		transcodeJobRepo: transcodeJobRepo,
		minioStorage:     minioStorage,
		userGateway:      userGateway,
	}
}
