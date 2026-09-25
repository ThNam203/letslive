package vod

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

func (s *VODService) SetAuthorDisabled(ctx context.Context, userId uuid.UUID, disabled bool) error {
	return s.vodRepo.SetAuthorDisabled(ctx, userId, disabled)
}

func (s *VODService) ReplaceDisabledAuthors(ctx context.Context, userIds []uuid.UUID) error {
	return s.vodRepo.ReplaceDisabledAuthors(ctx, userIds)
}
