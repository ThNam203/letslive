package dto

type SignUpRequestDTO struct {
	Username string `validate:"required,gte=6,lte=50" example:"sen1or"`
	Email    string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password string `validate:"required,password" example:"Password123!"`
	OTPCode  string `json:"otpCode" validate:"required,min=6,max=6"`
}
