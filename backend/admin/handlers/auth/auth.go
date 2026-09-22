package auth

import (
	"sen1or/letslive/admin/handlers/basehandler"
	"sen1or/letslive/admin/services"
)

type AuthHandler struct {
	basehandler.BaseHandler
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}
