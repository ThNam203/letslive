package dto

type CreateUserRequestDTO struct {
	Username     *string      `json:"username,omitempty" validate:"omitempty,gte=6,lte=30"`
	Email        string       `json:"email" validate:"required,email,lte=320"`
	AuthProvider AuthProvider `json:"authProvider" validate:"required"`
}

