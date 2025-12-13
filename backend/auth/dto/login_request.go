package dto

type LogInRequestDTO struct {
	Email          string `validate:"required,email" example:"hthnam203@gmail.com"`
	Password       string `validate:"required,password" example:"123123123"`
	TurnstileToken string `validate:"required"`
}
