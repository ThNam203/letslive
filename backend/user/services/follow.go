package services

import (
	"context"
	"sen1or/letslive/user/domains"

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

func (s FollowService) Follow(ctx context.Context, followId, followedId string) error {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return domains.ErrInvalidInput
	}

	follower, followerErr := s.userRepo.GetById(ctx, followUUID)
	if followerErr != nil {
		return followerErr
	}
	if follower.Status == domains.UserStatusDisabled {
		return domains.ErrAccountDisabled
	}

	followed, followedErr := s.userRepo.GetById(ctx, followedUUID)
	if followedErr != nil {
		return followedErr
	}
	if followed.Status == domains.UserStatusDisabled {
		return domains.ErrAccountDisabled
	}

	err := s.followRepo.FollowUser(ctx, followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}

func (s FollowService) Unfollow(ctx context.Context, followId, followedId string) error {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return domains.ErrInvalidInput
	}
	err := s.followRepo.UnfollowUser(ctx, followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}
