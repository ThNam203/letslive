package livestream

import (
	"context"
	"sen1or/lets-live/user/gateway"
)

type LivestreamGateway interface {
	GetUserLivestreams(ctx context.Context, userId string) ([]GetLivestreamResponseDTO, *gateway.ErrorResponse)
	CheckIsUserLivestreaming(ctx context.Context, userId string) (bool, *gateway.ErrorResponse)
}
