package livestream

import (
	"sen1or/letslive/livestream/domains"
)

type LivestreamService struct {
	livestreamRepo domains.LivestreamRepository
	vodRepo        domains.VODRepository
}

func NewLivestreamService(livestreamRepo domains.LivestreamRepository, vodRepo domains.VODRepository) *LivestreamService {
	return &LivestreamService{
		livestreamRepo: livestreamRepo,
		vodRepo:        vodRepo,
	}
}
