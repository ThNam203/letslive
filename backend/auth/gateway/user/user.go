package user

import (
	"context"
	"sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"
)

const (
	UserStatusNormal   = "normal"
	UserStatusDisabled = "disabled"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, *serviceresponse.Response[any])
	GetUserStatus(ctx context.Context, userId string) (string, *serviceresponse.Response[any])
	UpdateUserStatus(ctx context.Context, userId string, status string) *serviceresponse.Response[any]
}
