package dto

type SignUpRequestVerificationRequestDTO struct {
	Email          string `validate:"required,email" example:"hthnam203@gmail.com"`
	TurnstileToken string `validate:"required"`
}

