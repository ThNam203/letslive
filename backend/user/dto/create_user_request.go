package dto

type CreateUserRequestDTO struct {
	Username     string `json:"username" validate:"omitempty,gte=6,lte=30"`
	Email        string `json:"email" validate:"required,email"`
	AuthProvider string `json:"authProvider" validate:"oneof=google local"`
}
