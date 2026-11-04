package dto

type ChangePasswordRequestDTO struct {
	OldPassword string `json:"oldPassword" validate:"required,max=72" example:"OldPassword123!"`
	NewPassword string `json:"newPassword" validate:"required,password" example:"NewPassword123!"`
}
