package general

import (
	"sen1or/letslive/user/handlers/http/basehandler"
)

type GeneralHandler struct {
	basehandler.BaseHandler
}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}
