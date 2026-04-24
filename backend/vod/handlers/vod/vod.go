package vod

import (
	"sen1or/letslive/vod/handlers/basehandler"
	"sen1or/letslive/vod/services/vod"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type VODHandler struct {
	basehandler.BaseHandler
	vodService *vod.VODService
}

func NewVODHandler(vodService *vod.VODService) *VODHandler {
	return &VODHandler{
		vodService: vodService,
	}
}
