package general

import (
	"sen1or/letslive/finance/handlers/basehandler"
)

type GeneralHandler struct {
	basehandler.BaseHandler
}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}
