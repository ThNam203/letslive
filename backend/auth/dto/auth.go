package dto

type LogInRequestDTO struct {
	Email          string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password       string `validate:"required,gte=8,lte=72" example:"123123123"`
	TurnstileToken string `validate:"required"`
}

type SignUpRequestVerificationRequestDTO struct {
	Email          string `validate:"required,email" example:"hthnam203@gmail.com"`
	TurnstileToken string `validate:"required"`
}

type SignUpRequestDTO struct {
	Username string `validate:"required,gte=6,lte=50" example:"sen1or"`
	Email    string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password string `validate:"required,gte=8,lte=72" example:"123123123"`
	OTPCode  string `json:"otpCode" validate:"required,min=6,max=6"` // ??
}

type ChangePasswordRequestDTO struct {
	OldPassword string `json:"oldPassword" validate:"required,gte=8,lte=72" example:"123123123"`
	NewPassword string `json:"newPassword" validate:"required,gte=8,lte=72" example:"123123123"`
}
