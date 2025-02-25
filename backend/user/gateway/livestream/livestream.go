package transcode

import (
	"context"
	"sen1or/lets-live/user/dto"
	"sen1or/lets-live/user/gateway"
)

type LivestreamGateway interface {
	GetUserLivestreams(ctx context.Context, userId string) ([]dto.GetLivestreamResponseDTO, *gateway.ErrorResponse)
}
