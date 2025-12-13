package livestream_information

import (
	"sen1or/letslive/user/handlers/http/basehandler"
	"sen1or/letslive/user/services"
)

type LivestreamInformationHandler struct {
	basehandler.BaseHandler
	livestreamService services.LivestreamInformationService
	minioService      services.MinIOService
}

func NewLivestreamInformationHandler(livestreamService services.LivestreamInformationService, minioService services.MinIOService) *LivestreamInformationHandler {
	return &LivestreamInformationHandler{
		livestreamService: livestreamService,
		minioService:      minioService,
	}
}
