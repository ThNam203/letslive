package general

import (
	"sen1or/letslive/livestream/handlers/http/basehandler"
)

type GeneralHandler struct {
	basehandler.BaseHandler
}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}
