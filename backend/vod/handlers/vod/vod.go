package vod

import (
	"sen1or/letslive/vod/handlers/basehandler"
	"sen1or/letslive/vod/services/vod"
	vodreaction "sen1or/letslive/vod/services/vod_reaction"
)

type VODHandler struct {
	basehandler.BaseHandler
	vodService      *vod.VODService
	reactionService *vodreaction.VODReactionService
}

func NewVODHandler(vodService *vod.VODService, reactionService *vodreaction.VODReactionService) *VODHandler {
	return &VODHandler{
		vodService:      vodService,
		reactionService: reactionService,
	}
}
