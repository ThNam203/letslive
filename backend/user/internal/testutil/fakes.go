package testutil

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"

	"github.com/gofrs/uuid/v5"
)

type FakeUserRepository struct {
	GetByIdFunc                 func(ctx context.Context, userId uuid.UUID) (*domains.User, error)
	GetAllFunc                  func(ctx context.Context, page int) ([]domains.User, error)
	GetByUsernameFunc           func(ctx context.Context, username string) (*domains.User, error)
	GetByEmailFunc              func(ctx context.Context, email string) (*domains.User, error)
	GetByAPIKeyFunc             func(ctx context.Context, apiKey uuid.UUID) (*domains.User, error)
	GetPublicInfoByIdFunc       func(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, error)
	GetPublicInfosByIdsFunc     func(ctx context.Context, ids []uuid.UUID, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, error)
	GetRecommendedPublicFunc    func(ctx context.Context, excludeUserId *uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, error)
	SearchUsersByUsernameFunc   func(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, error)
	CreateFunc                  func(ctx context.Context, username string, email string, authProvider domains.AuthProvider) (*domains.User, error)
	UpdateFunc                  func(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, error)
	UpdateStreamAPIKeyFunc      func(ctx context.Context, userId uuid.UUID, newKey string) error
	UpdateProfilePictureFunc    func(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) error
	UpdateBackgroundPictureFunc func(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) error
	GetStatusesByIdsFunc        func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, error)
	GetDisabledUserIdsFunc      func(ctx context.Context) ([]uuid.UUID, error)
}

func (f *FakeUserRepository) GetById(ctx context.Context, userId uuid.UUID) (*domains.User, error) {
	return f.GetByIdFunc(ctx, userId)
}
func (f *FakeUserRepository) GetAll(ctx context.Context, page int) ([]domains.User, error) {
	return f.GetAllFunc(ctx, page)
}
func (f *FakeUserRepository) GetByUsername(ctx context.Context, username string) (*domains.User, error) {
	return f.GetByUsernameFunc(ctx, username)
}
func (f *FakeUserRepository) GetByEmail(ctx context.Context, email string) (*domains.User, error) {
	return f.GetByEmailFunc(ctx, email)
}
func (f *FakeUserRepository) GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*domains.User, error) {
	return f.GetByAPIKeyFunc(ctx, apiKey)
}
func (f *FakeUserRepository) GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, error) {
	return f.GetPublicInfoByIdFunc(ctx, userId, authenticatedUserId)
}
func (f *FakeUserRepository) GetPublicInfosByIds(ctx context.Context, ids []uuid.UUID, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, error) {
	return f.GetPublicInfosByIdsFunc(ctx, ids, authenticatedUserId)
}
func (f *FakeUserRepository) GetRecommendedPublic(ctx context.Context, excludeUserId *uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, error) {
	return f.GetRecommendedPublicFunc(ctx, excludeUserId, page, limit)
}
func (f *FakeUserRepository) SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, error) {
	return f.SearchUsersByUsernameFunc(ctx, username, authenticatedUserId)
}
func (f *FakeUserRepository) Create(ctx context.Context, username string, email string, authProvider domains.AuthProvider) (*domains.User, error) {
	return f.CreateFunc(ctx, username, email, authProvider)
}
func (f *FakeUserRepository) Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, error) {
	return f.UpdateFunc(ctx, user)
}
func (f *FakeUserRepository) UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) error {
	return f.UpdateStreamAPIKeyFunc(ctx, userId, newKey)
}
func (f *FakeUserRepository) UpdateProfilePicture(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) error {
	return f.UpdateProfilePictureFunc(ctx, userId, newProfilePictureURL)
}
func (f *FakeUserRepository) UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) error {
	return f.UpdateBackgroundPictureFunc(ctx, userId, newBackgroundPictureURL)
}
func (f *FakeUserRepository) GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, error) {
	return f.GetStatusesByIdsFunc(ctx, userIds)
}
func (f *FakeUserRepository) GetDisabledUserIds(ctx context.Context) ([]uuid.UUID, error) {
	return f.GetDisabledUserIdsFunc(ctx)
}

type FakeFollowRepository struct {
	FollowUserFunc         func(ctx context.Context, followUser, followedUser uuid.UUID) error
	UnfollowUserFunc       func(ctx context.Context, followUser, followedUser uuid.UUID) error
	GetFollowedUserIdsFunc func(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, error)
}

func (f *FakeFollowRepository) FollowUser(ctx context.Context, followUser, followedUser uuid.UUID) error {
	return f.FollowUserFunc(ctx, followUser, followedUser)
}
func (f *FakeFollowRepository) UnfollowUser(ctx context.Context, followUser, followedUser uuid.UUID) error {
	return f.UnfollowUserFunc(ctx, followUser, followedUser)
}
func (f *FakeFollowRepository) GetFollowedUserIds(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, error) {
	return f.GetFollowedUserIdsFunc(ctx, followerId)
}

var _ domains.UserRepository = (*FakeUserRepository)(nil)
var _ domains.FollowRepository = (*FakeFollowRepository)(nil)
