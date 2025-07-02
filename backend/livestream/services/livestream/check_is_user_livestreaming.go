package livestream

import (
	"context"
	serviceresponse "sen1or/letslive/livestream/responses"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) CheckIsUserLivestreaming(ctx context.Context, userId uuid.UUID) (bool, *serviceresponse.ServiceErrorResponse) {
	return s.livestreamRepo.CheckIsUserLivestreaming(ctx, userId)
}
