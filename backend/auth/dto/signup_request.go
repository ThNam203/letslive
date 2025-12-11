package dto

type SignUpRequestDTO struct {
	Username string `validate:"required,gte=6,lte=50" example:"sen1or"`
	Email    string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password string `validate:"required,gte=8,lte=72" example:"123123123"`
	OTPCode  string `json:"otpCode" validate:"required,min=6,max=6"` // ??
}

