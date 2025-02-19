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
		StartedAt:    dto.StartedAt,
	}
}

func UpdateLivestreamRequestDTOToLivestream(dto dto.UpdateLivestreamRequestDTO) domains.Livestream {
	return domains.Livestream{
		Id:           dto.Id,
		Title:        *dto.Title,
		Description:  *dto.Description,
		ThumbnailURL: *dto.ThumbnailURL,
		Status:       *dto.Status,
		PlaybackURL:  *dto.PlaybackURL,
		ViewCount:    *dto.ViewCount,
		EndedAt:      dto.EndedAt,
		StartedAt:    dto.StartedAt,
	}
}
