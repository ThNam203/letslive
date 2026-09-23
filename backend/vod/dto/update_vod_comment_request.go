package dto

type UpdateVODCommentRequestDTO struct {
	Content string `json:"content" validate:"required,gte=1,lte=2000"`
}
