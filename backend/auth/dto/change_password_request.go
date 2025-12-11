package dto

type ChangePasswordRequestDTO struct {
	OldPassword string `json:"oldPassword" validate:"required,gte=8,lte=72" example:"123123123"`
	NewPassword string `json:"newPassword" validate:"required,gte=8,lte=72" example:"123123123"`
}

