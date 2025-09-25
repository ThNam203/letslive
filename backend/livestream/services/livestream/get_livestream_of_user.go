package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

func (s LivestreamService) GetLivestreamOfUser(ctx context.Context, userId uuid.UUID) (*domains.Livestream, *response.Response[any]) {
	return s.livestreamRepo.GetByUser(ctx, userId)
}
