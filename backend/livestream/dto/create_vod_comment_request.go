package dto

type CreateVODCommentRequestDTO struct {
	Content  string  `json:"content" validate:"required,gte=1,lte=2000"`
	ParentId *string `json:"parentId,omitempty" validate:"omitempty,uuid"`
}
