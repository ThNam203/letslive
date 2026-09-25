package livestream

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

func (s *LivestreamService) SetAuthorDisabled(ctx context.Context, userId uuid.UUID, disabled bool) error {
	return s.livestreamRepo.SetAuthorDisabled(ctx, userId, disabled)
}

func (s *LivestreamService) ReplaceDisabledAuthors(ctx context.Context, userIds []uuid.UUID) error {
	return s.livestreamRepo.ReplaceDisabledAuthors(ctx, userIds)
}
