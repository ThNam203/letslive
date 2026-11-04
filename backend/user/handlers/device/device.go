package devicehandler

import (
	"sen1or/letslive/user/handlers/basehandler"
	"sen1or/letslive/user/services"
)

type DeviceHandler struct {
	basehandler.BaseHandler
	deviceService *services.DeviceService
}

func NewDeviceHandler(deviceService *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}
