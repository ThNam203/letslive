package general

import (
	"sen1or/letslive/user/handlers/basehandler"
)

type GeneralHandler struct {
	basehandler.BaseHandler
}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}
