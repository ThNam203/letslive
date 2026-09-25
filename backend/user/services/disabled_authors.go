package services

import (
	"context"
	"fmt"
	"sen1or/letslive/shared/pkg/logger"
	"sen1or/letslive/user/domains"
	"time"

	"github.com/gofrs/uuid/v5"
)

const disabledAuthorsSyncRetryInterval = 15 * time.Second

func (s *UserService) syncAuthorStatus(ctx context.Context, userId uuid.UUID, status string) error {
	disabled := domains.UserStatus(status) == domains.UserStatusDisabled
	if err := s.contentGateway.SetAuthorDisabled(ctx, userId, disabled); err != nil {
		logger.Errorf(ctx, "failed to sync author status of user %s (disabled=%t): %v", userId, disabled, err)
		return fmt.Errorf("%w: %w", domains.ErrInternal, err)
	}

	return nil
}

// SyncDisabledAuthors pushes the full set of disabled users to the content
// services, retrying until it succeeds or ctx is done.
func (s *UserService) SyncDisabledAuthors(ctx context.Context) {
	for {
		err := s.replaceDisabledAuthors(ctx)
		if err == nil {
			logger.Infof(ctx, "synced disabled authors to content services")
			return
		}
		logger.Warnf(ctx, "failed to sync disabled authors, retrying in %s: %v", disabledAuthorsSyncRetryInterval, err)

		select {
		case <-ctx.Done():
			return
		case <-time.After(disabledAuthorsSyncRetryInterval):
		}
	}
}

func (s *UserService) replaceDisabledAuthors(ctx context.Context) error {
	ids, err := s.userRepo.GetDisabledUserIds(ctx)
	if err != nil {
		return err
	}

	return s.contentGateway.ReplaceDisabledAuthors(ctx, ids)
}
