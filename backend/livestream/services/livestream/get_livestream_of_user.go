package livestream

import (
	"context"
	"sen1or/letslive/livestream/domains"

	"github.com/gofrs/uuid/v5"
)

func (s LivestreamService) GetLivestreamOfUser(ctx context.Context, userId uuid.UUID) (*domains.Livestream, error) {
	return s.livestreamRepo.GetByUser(ctx, userId)
}
