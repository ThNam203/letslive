package dto

type SignUpRequestVerificationRequestDTO struct {
	Email          string `json:"email" validate:"required,email" example:"hthnam203@gmail.com"`
	TurnstileToken string `json:"turnstileToken" validate:"required"`
}

