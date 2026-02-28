package dto

type SignUpRequestVerificationRequestDTO struct {
	Email          string `json:"email" validate:"required,email,lte=320" example:"hthnam203@gmail.com"`
	TurnstileToken string `json:"turnstileToken" validate:"required,lte=2048"`
}

