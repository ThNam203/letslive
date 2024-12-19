package user

import (
	"context"
	"sen1or/lets-live/auth/gateway"
	"sen1or/lets-live/user/dto"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, *gateway.ErrorResponse)
}
