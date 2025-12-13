package livestream

import (
	"sen1or/letslive/livestream/handlers/http/basehandler"
	"sen1or/letslive/livestream/services/livestream"
)

type LivestreamHandler struct {
	basehandler.BaseHandler
	livestreamService *livestream.LivestreamService
}

func NewLivestreamHandler(livestreamService *livestream.LivestreamService) *LivestreamHandler {
	return &LivestreamHandler{
		livestreamService: livestreamService,
	}
}
