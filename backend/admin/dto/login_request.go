package dto

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email,lte=320"`
	Password string `json:"password" validate:"required"`
}
