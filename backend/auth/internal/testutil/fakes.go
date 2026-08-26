package testutil

import (
	"context"
	"sen1or/letslive/auth/domains"
	usergatewaydto "sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"

	"github.com/gofrs/uuid/v5"
)

type FakeAuthRepository struct {
	GetByEmailFunc         func(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any])
	GetByUserIDFunc        func(ctx context.Context, userId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any])
	GetByIDFunc            func(ctx context.Context, authId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any])
	CreateFunc             func(ctx context.Context, auth domains.Auth) (*domains.Auth, *serviceresponse.Response[any])
	UpdatePasswordHashFunc func(ctx context.Context, authId, newPasswordHash string) *serviceresponse.Response[any]
}

func (f *FakeAuthRepository) GetByEmail(ctx context.Context, email string) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.GetByEmailFunc(ctx, email)
}

func (f *FakeAuthRepository) GetByUserID(ctx context.Context, userId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.GetByUserIDFunc(ctx, userId)
}

func (f *FakeAuthRepository) GetByID(ctx context.Context, authId uuid.UUID) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.GetByIDFunc(ctx, authId)
}

func (f *FakeAuthRepository) Create(ctx context.Context, auth domains.Auth) (*domains.Auth, *serviceresponse.Response[any]) {
	return f.CreateFunc(ctx, auth)
}

func (f *FakeAuthRepository) UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) *serviceresponse.Response[any] {
	return f.UpdatePasswordHashFunc(ctx, authId, newPasswordHash)
}

type FakeRefreshTokenRepository struct {
	RevokeAllTokensOfUserFunc func(ctx context.Context, userId uuid.UUID) *serviceresponse.Response[any]
	InsertFunc                func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any]
	FindByValueFunc           func(ctx context.Context, value string) (*domains.RefreshToken, *serviceresponse.Response[any])
	UpdateFunc                func(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any]
}

func (f *FakeRefreshTokenRepository) RevokeAllTokensOfUser(ctx context.Context, userId uuid.UUID) *serviceresponse.Response[any] {
	return f.RevokeAllTokensOfUserFunc(ctx, userId)
}

func (f *FakeRefreshTokenRepository) Insert(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
	return f.InsertFunc(ctx, token)
}

func (f *FakeRefreshTokenRepository) FindByValue(ctx context.Context, value string) (*domains.RefreshToken, *serviceresponse.Response[any]) {
	return f.FindByValueFunc(ctx, value)
}

func (f *FakeRefreshTokenRepository) Update(ctx context.Context, token *domains.RefreshToken) *serviceresponse.Response[any] {
	return f.UpdateFunc(ctx, token)
}

type FakeUserGateway struct {
	CreateNewUserFunc    func(ctx context.Context, requestDTO usergatewaydto.CreateUserRequestDTO) (*usergatewaydto.CreateUserResponseDTO, *serviceresponse.Response[any])
	GetUserStatusFunc    func(ctx context.Context, userId string) (string, *serviceresponse.Response[any])
	UpdateUserStatusFunc func(ctx context.Context, userId string, status string) *serviceresponse.Response[any]
}

func (f *FakeUserGateway) CreateNewUser(ctx context.Context, requestDTO usergatewaydto.CreateUserRequestDTO) (*usergatewaydto.CreateUserResponseDTO, *serviceresponse.Response[any]) {
	return f.CreateNewUserFunc(ctx, requestDTO)
}

func (f *FakeUserGateway) GetUserStatus(ctx context.Context, userId string) (string, *serviceresponse.Response[any]) {
	return f.GetUserStatusFunc(ctx, userId)
}

func (f *FakeUserGateway) UpdateUserStatus(ctx context.Context, userId string, status string) *serviceresponse.Response[any] {
	return f.UpdateUserStatusFunc(ctx, userId, status)
}
