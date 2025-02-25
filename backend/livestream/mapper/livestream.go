package mapper

import (
	"sen1or/lets-live/livestream/domains"
	"sen1or/lets-live/livestream/dto"
)

func CreateLivestreamRequestDTOToLivestream(dto dto.CreateLivestreamRequestDTO) domains.Livestream {
	return domains.Livestream{
		UserId:       dto.UserId,
		Title:        dto.Title,
		Description:  dto.Description,
		ThumbnailURL: dto.ThumbnailURL,
		Status:       dto.Status,
	}
}
