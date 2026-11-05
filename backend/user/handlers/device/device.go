package devicehandler

import (
	"sen1or/letslive/user/handlers/basehandler"
	"sen1or/letslive/user/services"
)

// DeviceHandler serves the signed-in user's push devices
// (/user/me/devices).
type DeviceHandler struct {
	basehandler.BaseHandler
	deviceService *services.DeviceService
}

// NewDeviceHandler builds a DeviceHandler over deviceService.
func NewDeviceHandler(deviceService *services.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}
