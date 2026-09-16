package user

import (
	"context"
	"sen1or/letslive/auth/gateway/user/dto"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, error)
}
