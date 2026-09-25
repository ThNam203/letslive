package livestream

import (
	"sen1or/letslive/livestream/domains"
	vodgateway "sen1or/letslive/livestream/gateway/vod"
)

type LivestreamService struct {
	livestreamRepo domains.LivestreamRepository
	vodGateway     vodgateway.VODGateway
}

func NewLivestreamService(livestreamRepo domains.LivestreamRepository, vodGateway vodgateway.VODGateway) *LivestreamService {
	return &LivestreamService{
		livestreamRepo: livestreamRepo,
		vodGateway:     vodGateway,
	}
}
