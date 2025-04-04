package dto

type LogInRequestDTO struct {
	Email          string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password       string `validate:"required,gte=8,lte=72" example:"123123123"`
	TurnstileToken string `validate:"required"`
}

type SignUpRequestDTO struct {
	Username       string `validate:"required,gte=6,lte=50" example:"sen1or"`
	Email          string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password       string `validate:"required,gte=8,lte=72" example:"123123123"`
	TurnstileToken string `validate:"required"`
}

type ChangePasswordRequestDTO struct {
	OldPassword string `json:"oldPassword" validate:"required,gte=8,lte=72" example:"123123123"`
	NewPassword string `json:"newPassword" validate:"required,gte=8,lte=72" example:"123123123"`
}
