package mapper

import (
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/dto"
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
