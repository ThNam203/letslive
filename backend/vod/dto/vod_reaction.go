package dto

import "sen1or/letslive/vod/domains"

type SetVODReactionRequestDTO struct {
	Reaction string `json:"reaction" validate:"required,oneof=like dislike"`
}

// VODReactionResponseDTO is the caller's own reaction and the public like count.
// The dislike count is intentionally absent.
type VODReactionResponseDTO struct {
	LikeCount int64                    `json:"likeCount"`
	Reaction  *domains.VODReactionType `json:"reaction"`
}
