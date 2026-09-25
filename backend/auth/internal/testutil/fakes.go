package testutil

import (
	"context"
	"sen1or/letslive/auth/domains"
	usergatewaydto "sen1or/letslive/auth/gateway/user/dto"

	"github.com/gofrs/uuid/v5"
)

type FakeAuthRepository struct {
	GetByEmailFunc         func(ctx context.Context, email string) (*domains.Auth, error)
	GetByUserIDFunc        func(ctx context.Context, userId uuid.UUID) (*domains.Auth, error)
	GetByIDFunc            func(ctx context.Context, authId uuid.UUID) (*domains.Auth, error)
	CreateFunc             func(ctx context.Context, auth domains.Auth) (*domains.Auth, error)
	UpdatePasswordHashFunc func(ctx context.Context, authId, newPasswordHash string) error
}

func (f *FakeAuthRepository) GetByEmail(ctx context.Context, email string) (*domains.Auth, error) {
	return f.GetByEmailFunc(ctx, email)
}

func (f *FakeAuthRepository) GetByUserID(ctx context.Context, userId uuid.UUID) (*domains.Auth, error) {
	return f.GetByUserIDFunc(ctx, userId)
}

func (f *FakeAuthRepository) GetByID(ctx context.Context, authId uuid.UUID) (*domains.Auth, error) {
	return f.GetByIDFunc(ctx, authId)
}

func (f *FakeAuthRepository) Create(ctx context.Context, auth domains.Auth) (*domains.Auth, error) {
	return f.CreateFunc(ctx, auth)
}

func (f *FakeAuthRepository) UpdatePasswordHash(ctx context.Context, authId, newPasswordHash string) error {
	return f.UpdatePasswordHashFunc(ctx, authId, newPasswordHash)
}

type FakeRefreshTokenRepository struct {
	RevokeAllTokensOfUserFunc func(ctx context.Context, userId uuid.UUID) error
	InsertFunc                func(ctx context.Context, token *domains.RefreshToken) error
	FindByValueFunc           func(ctx context.Context, value string) (*domains.RefreshToken, error)
	UpdateFunc                func(ctx context.Context, token *domains.RefreshToken) error
}

func (f *FakeRefreshTokenRepository) RevokeAllTokensOfUser(ctx context.Context, userId uuid.UUID) error {
	return f.RevokeAllTokensOfUserFunc(ctx, userId)
}

func (f *FakeRefreshTokenRepository) Insert(ctx context.Context, token *domains.RefreshToken) error {
	return f.InsertFunc(ctx, token)
}

func (f *FakeRefreshTokenRepository) FindByValue(ctx context.Context, value string) (*domains.RefreshToken, error) {
	return f.FindByValueFunc(ctx, value)
}

func (f *FakeRefreshTokenRepository) Update(ctx context.Context, token *domains.RefreshToken) error {
	return f.UpdateFunc(ctx, token)
}

type FakeUserGateway struct {
	CreateNewUserFunc    func(ctx context.Context, requestDTO usergatewaydto.CreateUserRequestDTO) (*usergatewaydto.CreateUserResponseDTO, error)
	GetUserStatusFunc    func(ctx context.Context, userId string) (string, error)
	UpdateUserStatusFunc func(ctx context.Context, userId string, status string) error
}

func (f *FakeUserGateway) CreateNewUser(ctx context.Context, requestDTO usergatewaydto.CreateUserRequestDTO) (*usergatewaydto.CreateUserResponseDTO, error) {
	return f.CreateNewUserFunc(ctx, requestDTO)
}

func (f *FakeUserGateway) GetUserStatus(ctx context.Context, userId string) (string, error) {
	return f.GetUserStatusFunc(ctx, userId)
}

func (f *FakeUserGateway) UpdateUserStatus(ctx context.Context, userId string, status string) error {
	return f.UpdateUserStatusFunc(ctx, userId, status)
}
