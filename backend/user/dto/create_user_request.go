package dto

type CreateUserRequestDTO struct {
	Username     string `json:"username" validate:"required,gte=4,lte=50"`
	Email        string `json:"email" validate:"required,email"`
	AuthProvider string `json:"authProvider" validate:"oneof=google local"`
}

