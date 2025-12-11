package user

import (
	"sen1or/letslive/user/handlers/basehandler"
	"sen1or/letslive/user/services"
)

type UserHandler struct {
	basehandler.BaseHandler
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}
