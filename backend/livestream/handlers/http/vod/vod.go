package vod

import (
	"sen1or/letslive/livestream/handlers/http/basehandler"
	"sen1or/letslive/livestream/services/vod"
)

type VODHandler struct {
	basehandler.BaseHandler
	vodService *vod.VODService
}

func NewVODHandler(vodService *vod.VODService) *VODHandler {
	return &VODHandler{
		vodService: vodService,
	}
}
