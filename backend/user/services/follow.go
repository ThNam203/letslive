package services

import (
	servererrors "sen1or/lets-live/user/errors"
	"sen1or/lets-live/user/repositories"

	"github.com/gofrs/uuid/v5"
)

type FollowService struct {
	followRepo repositories.FollowRepository
}

func NewFollowService(
	followRepo repositories.FollowRepository,
) *FollowService {
	return &FollowService{
		followRepo: followRepo,
	}
}

func (s FollowService) Follow(followId, followedId string) *servererrors.ServerError {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return servererrors.ErrInvalidInput
	}
	err := s.followRepo.FollowUser(followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}

func (s FollowService) Unfollow(followId, followedId string) *servererrors.ServerError {
	followUUID, err1 := uuid.FromString(followId)
	followedUUID, err2 := uuid.FromString(followedId)
	if err1 != nil || err2 != nil || followId == followedId {
		return servererrors.ErrInvalidInput
	}
	err := s.followRepo.UnfollowUser(followUUID, followedUUID)
	if err != nil {
		return err
	}

	return nil
}
