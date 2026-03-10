package vodcomment

import (
	"sen1or/letslive/vod/handlers/basehandler"
	vodcommentservice "sen1or/letslive/vod/services/vod_comment"
)

type VODCommentHandler struct {
	basehandler.BaseHandler
	commentService *vodcommentservice.VODCommentService
}

func NewVODCommentHandler(commentService *vodcommentservice.VODCommentService) *VODCommentHandler {
	return &VODCommentHandler{
		commentService: commentService,
	}
}
