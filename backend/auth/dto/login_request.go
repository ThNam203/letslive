package dto

type LogInRequestDTO struct {
	Email          string `json:"email" validate:"required,email" example:"hthnam203@gmail.com"`
	Password       string `json:"password" validate:"required,password" example:"123123123"`
	TurnstileToken string `json:"turnstileToken" validate:"required"`
}
