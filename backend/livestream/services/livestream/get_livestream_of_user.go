package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"
	"sen1or/letslive/livestream/responses"

	"github.com/gofrs/uuid/v5"
)

func (s LivestreamService) GetLivestreamOfUser(ctx context.Context, userId uuid.UUID) (*domains.Livestream, *serviceresponse.ServiceErrorResponse) {
	return s.livestreamRepo.GetByUser(ctx, userId)
}
