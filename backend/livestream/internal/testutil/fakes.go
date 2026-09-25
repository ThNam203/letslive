package testutil

import (
	"context"
	"sen1or/letslive/livestream/domains"
	usergateway "sen1or/letslive/livestream/gateway/user"

	"github.com/gofrs/uuid/v5"
)

type FakeLivestreamRepository struct {
	GetByIdFunc                   func(ctx context.Context, id uuid.UUID) (*domains.Livestream, error)
	GetByUserFunc                 func(ctx context.Context, userId uuid.UUID) (*domains.Livestream, error)
	GetRecommendedLivestreamsFunc func(ctx context.Context, page, limit int) ([]domains.Livestream, error)
	CreateFunc                    func(ctx context.Context, ls domains.Livestream) (*domains.Livestream, error)
	UpdateFunc                    func(ctx context.Context, ls domains.Livestream) (*domains.Livestream, error)
	DeleteFunc                    func(ctx context.Context, id uuid.UUID) error
	SetAuthorDisabledFunc         func(ctx context.Context, userId uuid.UUID, disabled bool) error
	ReplaceDisabledAuthorsFunc    func(ctx context.Context, userIds []uuid.UUID) error
}

func (f *FakeLivestreamRepository) GetById(ctx context.Context, id uuid.UUID) (*domains.Livestream, error) {
	return f.GetByIdFunc(ctx, id)
}
func (f *FakeLivestreamRepository) GetByUser(ctx context.Context, userId uuid.UUID) (*domains.Livestream, error) {
	return f.GetByUserFunc(ctx, userId)
}
func (f *FakeLivestreamRepository) GetRecommendedLivestreams(ctx context.Context, page, limit int) ([]domains.Livestream, error) {
	return f.GetRecommendedLivestreamsFunc(ctx, page, limit)
}
func (f *FakeLivestreamRepository) Create(ctx context.Context, ls domains.Livestream) (*domains.Livestream, error) {
	return f.CreateFunc(ctx, ls)
}
func (f *FakeLivestreamRepository) Update(ctx context.Context, ls domains.Livestream) (*domains.Livestream, error) {
	return f.UpdateFunc(ctx, ls)
}
func (f *FakeLivestreamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return f.DeleteFunc(ctx, id)
}
func (f *FakeLivestreamRepository) SetAuthorDisabled(ctx context.Context, userId uuid.UUID, disabled bool) error {
	return f.SetAuthorDisabledFunc(ctx, userId, disabled)
}
func (f *FakeLivestreamRepository) ReplaceDisabledAuthors(ctx context.Context, userIds []uuid.UUID) error {
	return f.ReplaceDisabledAuthorsFunc(ctx, userIds)
}

type FakeUserGateway struct {
	GetUserPublicInfoFunc func(ctx context.Context, userId uuid.UUID) (*usergateway.UserPublicInfo, error)
}

func (f *FakeUserGateway) GetUserPublicInfo(ctx context.Context, userId uuid.UUID) (*usergateway.UserPublicInfo, error) {
	return f.GetUserPublicInfoFunc(ctx, userId)
}

var _ domains.LivestreamRepository = (*FakeLivestreamRepository)(nil)
