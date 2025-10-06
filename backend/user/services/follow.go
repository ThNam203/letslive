package services

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

type FollowService struct {
	followRepo domains.FollowRepository
}

func NewFollowService(
	followRepo domains.FollowRepository,
) *FollowService {
	return &FollowService{
		followRepo: followRepo,
	}
}

func (s FollowService) Follow(ctx context.Context, followId, followedId string) *response.Response[any] {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}
	err := s.followRepo.FollowUser(ctx, followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}

func (s FollowService) Unfollow(ctx context.Context, followId, followedId string) *response.Response[any] {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		)
	}
	err := s.followRepo.UnfollowUser(ctx, followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}
