package dto

type CreateUserRequestDTO struct {
	Username     string       `json:"username"`
	Email        string       `json:"email"`
	AuthProvider AuthProvider `json:"authProvider"`
}

