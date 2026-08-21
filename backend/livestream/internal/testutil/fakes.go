package testutil

import (
	"context"
	"sen1or/letslive/livestream/domains"
	usergateway "sen1or/letslive/livestream/gateway/user"
	"sen1or/letslive/livestream/response"

	"github.com/gofrs/uuid/v5"
)

type FakeLivestreamRepository struct {
	GetByIdFunc                   func(ctx context.Context, id uuid.UUID) (*domains.Livestream, *response.Response[any])
	GetByUserFunc                 func(ctx context.Context, userId uuid.UUID) (*domains.Livestream, *response.Response[any])
	GetRecommendedLivestreamsFunc func(ctx context.Context, page, limit int) ([]domains.Livestream, *response.Response[any])
	CreateFunc                    func(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any])
	UpdateFunc                    func(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any])
	DeleteFunc                    func(ctx context.Context, id uuid.UUID) *response.Response[any]
}

func (f *FakeLivestreamRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.Livestream, *response.Response[any]) {
	return f.GetByIdFunc(ctx, id)
}
func (f *FakeLivestreamRepository) GetByUser(ctx context.Context, userId uuid.UUID) (*domains.Livestream, *response.Response[any]) {
	return f.GetByUserFunc(ctx, userId)
}
func (f *FakeLivestreamRepository) GetRecommendedLivestreams(ctx context.Context, page, limit int) ([]domains.Livestream, *response.Response[any]) {
	return f.GetRecommendedLivestreamsFunc(ctx, page, limit)
}
func (f *FakeLivestreamRepository) Create(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any]) {
	return f.CreateFunc(ctx, ls)
}
func (f *FakeLivestreamRepository) Update(ctx context.Context, ls domains.Livestream) (*domains.Livestream, *response.Response[any]) {
	return f.UpdateFunc(ctx, ls)
}
func (f *FakeLivestreamRepository) Delete(ctx context.Context, id uuid.UUID) *response.Response[any] {
	return f.DeleteFunc(ctx, id)
}

type FakeUserGateway struct {
	GetUserPublicInfoFunc func(ctx context.Context, userId uuid.UUID) (*usergateway.UserPublicInfo, error)
	GetUsersStatusesFunc  func(ctx context.Context, userIds []uuid.UUID) (map[string]string, error)
}

func (f *FakeUserGateway) GetUserPublicInfo(ctx context.Context, userId uuid.UUID) (*usergateway.UserPublicInfo, error) {
	return f.GetUserPublicInfoFunc(ctx, userId)
}
func (f *FakeUserGateway) GetUsersStatuses(ctx context.Context, userIds []uuid.UUID) (map[string]string, error) {
	return f.GetUsersStatusesFunc(ctx, userIds)
}
