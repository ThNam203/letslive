package testutil

import (
	"context"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/response"

	"github.com/gofrs/uuid/v5"
)

type FakeUserRepository struct {
	GetByIdFunc                 func(ctx context.Context, userId uuid.UUID) (*domains.User, *response.Response[any])
	GetAllFunc                  func(ctx context.Context, page int) ([]domains.User, *response.Response[any])
	GetByUsernameFunc           func(ctx context.Context, username string) (*domains.User, *response.Response[any])
	GetByEmailFunc              func(ctx context.Context, email string) (*domains.User, *response.Response[any])
	GetByAPIKeyFunc             func(ctx context.Context, apiKey uuid.UUID) (*domains.User, *response.Response[any])
	GetPublicInfoByIdFunc       func(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *response.Response[any])
	GetPublicInfosByIdsFunc     func(ctx context.Context, ids []uuid.UUID, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any])
	GetRecommendedPublicFunc    func(ctx context.Context, excludeUserId *uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, *response.Response[any])
	SearchUsersByUsernameFunc   func(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any])
	CreateFunc                  func(ctx context.Context, username string, email string, authProvider domains.AuthProvider) (*domains.User, *response.Response[any])
	UpdateFunc                  func(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any])
	UpdateStreamAPIKeyFunc      func(ctx context.Context, userId uuid.UUID, newKey string) *response.Response[any]
	UpdateProfilePictureFunc    func(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) *response.Response[any]
	UpdateBackgroundPictureFunc func(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) *response.Response[any]
	GetStatusesByIdsFunc        func(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any])
}

func (f *FakeUserRepository) GetById(ctx context.Context, userId uuid.UUID) (*domains.User, *response.Response[any]) {
	return f.GetByIdFunc(ctx, userId)
}
func (f *FakeUserRepository) GetAll(ctx context.Context, page int) ([]domains.User, *response.Response[any]) {
	return f.GetAllFunc(ctx, page)
}
func (f *FakeUserRepository) GetByUsername(ctx context.Context, username string) (*domains.User, *response.Response[any]) {
	return f.GetByUsernameFunc(ctx, username)
}
func (f *FakeUserRepository) GetByEmail(ctx context.Context, email string) (*domains.User, *response.Response[any]) {
	return f.GetByEmailFunc(ctx, email)
}
func (f *FakeUserRepository) GetByAPIKey(ctx context.Context, apiKey uuid.UUID) (*domains.User, *response.Response[any]) {
	return f.GetByAPIKeyFunc(ctx, apiKey)
}
func (f *FakeUserRepository) GetPublicInfoById(ctx context.Context, userId uuid.UUID, authenticatedUserId *uuid.UUID) (*dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.GetPublicInfoByIdFunc(ctx, userId, authenticatedUserId)
}
func (f *FakeUserRepository) GetPublicInfosByIds(ctx context.Context, ids []uuid.UUID, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.GetPublicInfosByIdsFunc(ctx, ids, authenticatedUserId)
}
func (f *FakeUserRepository) GetRecommendedPublic(ctx context.Context, excludeUserId *uuid.UUID, page, limit int) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.GetRecommendedPublicFunc(ctx, excludeUserId, page, limit)
}
func (f *FakeUserRepository) SearchUsersByUsername(ctx context.Context, username string, authenticatedUserId *uuid.UUID) ([]dto.GetUserPublicResponseDTO, *response.Response[any]) {
	return f.SearchUsersByUsernameFunc(ctx, username, authenticatedUserId)
}
func (f *FakeUserRepository) Create(ctx context.Context, username string, email string, authProvider domains.AuthProvider) (*domains.User, *response.Response[any]) {
	return f.CreateFunc(ctx, username, email, authProvider)
}
func (f *FakeUserRepository) Update(ctx context.Context, user dto.UpdateUserRequestDTO) (*domains.User, *response.Response[any]) {
	return f.UpdateFunc(ctx, user)
}
func (f *FakeUserRepository) UpdateStreamAPIKey(ctx context.Context, userId uuid.UUID, newKey string) *response.Response[any] {
	return f.UpdateStreamAPIKeyFunc(ctx, userId, newKey)
}
func (f *FakeUserRepository) UpdateProfilePicture(ctx context.Context, userId uuid.UUID, newProfilePictureURL string) *response.Response[any] {
	return f.UpdateProfilePictureFunc(ctx, userId, newProfilePictureURL)
}
func (f *FakeUserRepository) UpdateBackgroundPicture(ctx context.Context, userId uuid.UUID, newBackgroundPictureURL string) *response.Response[any] {
	return f.UpdateBackgroundPictureFunc(ctx, userId, newBackgroundPictureURL)
}
func (f *FakeUserRepository) GetStatusesByIds(ctx context.Context, userIds []uuid.UUID) (map[uuid.UUID]domains.UserStatus, *response.Response[any]) {
	return f.GetStatusesByIdsFunc(ctx, userIds)
}

type FakeFollowRepository struct {
	FollowUserFunc         func(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any]
	UnfollowUserFunc       func(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any]
	GetFollowedUserIdsFunc func(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, *response.Response[any])
}

func (f *FakeFollowRepository) FollowUser(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any] {
	return f.FollowUserFunc(ctx, followUser, followedUser)
}
func (f *FakeFollowRepository) UnfollowUser(ctx context.Context, followUser, followedUser uuid.UUID) *response.Response[any] {
	return f.UnfollowUserFunc(ctx, followUser, followedUser)
}
func (f *FakeFollowRepository) GetFollowedUserIds(ctx context.Context, followerId uuid.UUID) ([]uuid.UUID, *response.Response[any]) {
	return f.GetFollowedUserIdsFunc(ctx, followerId)
}
