package livestream

import (
	"sen1or/letslive/livestream/domains"
	usergateway "sen1or/letslive/livestream/gateway/user"
	vodgateway "sen1or/letslive/livestream/gateway/vod"
)

type LivestreamService struct {
	livestreamRepo domains.LivestreamRepository
	vodGateway     vodgateway.VODGateway
	userGateway    usergateway.UserGateway
}

func NewLivestreamService(livestreamRepo domains.LivestreamRepository, vodGateway vodgateway.VODGateway, userGateway usergateway.UserGateway) *LivestreamService {
	return &LivestreamService{
		livestreamRepo: livestreamRepo,
		vodGateway:     vodGateway,
		userGateway:    userGateway,
	}
}
