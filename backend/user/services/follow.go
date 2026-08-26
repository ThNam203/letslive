package services

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

type FollowService struct {
	followRepo domains.FollowRepository
	userRepo   domains.UserRepository
}

func NewFollowService(
	followRepo domains.FollowRepository,
	userRepo domains.UserRepository,
) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
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

	follower, followerErr := s.userRepo.GetById(ctx, followUUID)
	if followerErr != nil {
		return followerErr
	}
	if follower.Status == domains.UserStatusDisabled {
		return response.NewResponseFromTemplate[any](response.RES_ERR_ACCOUNT_DISABLED, nil, nil, nil)
	}

	followed, followedErr := s.userRepo.GetById(ctx, followedUUID)
	if followedErr != nil {
		return followedErr
	}
	if followed.Status == domains.UserStatusDisabled {
		return response.NewResponseFromTemplate[any](response.RES_ERR_ACCOUNT_DISABLED, nil, nil, nil)
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
