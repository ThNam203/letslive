package dto

type ChangePasswordRequestDTO struct {
	OldPassword string `json:"oldPassword" validate:"required,password" example:"OldPassword123!"`
	NewPassword string `json:"newPassword" validate:"required,password" example:"NewPassword123!"`
}
