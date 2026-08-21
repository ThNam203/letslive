package dto

import "github.com/gofrs/uuid/v5"

type GetUsersStatusesRequestDTO struct {
	UserIds []uuid.UUID `json:"userIds" validate:"required,min=1"`
}
