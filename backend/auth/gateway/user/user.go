package user

import (
	"context"
	"sen1or/letslive/auth/gateway/user/dto"
)

const (
	UserStatusNormal   = "normal"
	UserStatusDisabled = "disabled"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error)
	GetUserStatus(ctx context.Context, userId string) (string, error)
	UpdateUserStatus(ctx context.Context, userId string, status string) error
}
