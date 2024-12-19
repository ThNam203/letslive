package dto

import "github.com/gofrs/uuid/v5"

type LogInRequestDTO struct {
	Email    string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password string `validate:"required,gte=8,lte=72" example:"123123123"`
}

type SignUpRequestDTO struct {
	Username string `validate:"required,gte=6,lte=50" example:"sen1or"`
	Email    string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password string `validate:"required,gte=8,lte=72" example:"123123123"`
}

type SignUpResponseDTO struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"userID"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"isVerified"`
}

type ChangePasswordRequestDTO struct {
	OldPassword        string `json:"oldPassword" validate:"required,gte=8,lte=72" example:"123123123"`
	NewPassword        string `json:"newPassword" validate:"required,gte=8,lte=72" example:"123123123"`
	ConfirmNewPassword string `json:"confirmNewPassword" validate:"required,gte=8,lte=72" example:"123123123"`
}
