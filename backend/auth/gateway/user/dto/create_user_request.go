package dto

type CreateUserRequestDTO struct {
	Username     string       `json:"username" validate:"required,gte=6,lte=50"`
	Email        string       `json:"email" validate:"required,email,lte=320"`
	AuthProvider AuthProvider `json:"authProvider" validate:"required"`
}

