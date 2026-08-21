package dto

type ReactivateRequestDTO struct {
	ReactivationToken string `json:"reactivationToken" validate:"required"`
}
