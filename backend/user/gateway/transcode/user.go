package transcode

import (
	"context"
	"sen1or/lets-live/user/dto"
	"sen1or/lets-live/user/gateway"
)

type TranscodeGateway interface {
	GetUserVODs(ctx context.Context, userId string) (*dto.TranscodeService_GetUserResponse, *gateway.ErrorResponse)
}
