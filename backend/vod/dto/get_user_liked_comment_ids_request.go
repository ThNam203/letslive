package dto

type GetUserLikedCommentIdsRequestDTO struct {
	CommentIds []string `json:"commentIds" validate:"required,min=1,max=100,dive,uuid"`
}
