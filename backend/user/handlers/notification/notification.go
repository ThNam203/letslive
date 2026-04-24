package notification

import (
	"sen1or/letslive/user/handlers/basehandler"
	"sen1or/letslive/user/services"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type NotificationHandler struct {
	basehandler.BaseHandler
	notificationService services.NotificationService
}

func NewNotificationHandler(notificationService services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}
