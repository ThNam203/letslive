package dmmessage

import (
	"sen1or/letslive/chat/handlers/basehandler"
	"sen1or/letslive/chat/services"
)

type DmMessageHandler struct {
	basehandler.BaseHandler
	dmMessageService *services.DmMessageService
}

func NewDmMessageHandler(dmMessageService *services.DmMessageService) *DmMessageHandler {
	return &DmMessageHandler{dmMessageService: dmMessageService}
}
