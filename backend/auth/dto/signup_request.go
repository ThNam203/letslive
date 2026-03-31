package dto

type SignUpRequestDTO struct {
	Username string `json:"username" validate:"required,gte=6,lte=50" example:"sen1or"`
	Email    string `json:"email" validate:"required,email,lte=320" example:"hthnam203@gmail.com"`
	Password string `json:"password" validate:"required,password" example:"Password123!"`
	OTPCode  string `json:"otpCode" validate:"required,min=6,max=6"`
}
