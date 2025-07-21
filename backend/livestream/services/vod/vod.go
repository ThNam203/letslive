package vod

import (
	"sen1or/letslive/livestream/domains"
)

type VODService struct {
	vodRepo domains.VODRepository
}

func NewVODService(vodRepo domains.VODRepository) *VODService {
	return &VODService{
		vodRepo: vodRepo,
	}
}
