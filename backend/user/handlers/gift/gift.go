package gifthandler

import (
	"sen1or/letslive/user/handlers/basehandler"
	"sen1or/letslive/user/services"
)

type GiftHandler struct {
	basehandler.BaseHandler
	giftService *services.GiftService
}

func NewGiftHandler(giftService *services.GiftService) *GiftHandler {
	return &GiftHandler{giftService: giftService}
}
