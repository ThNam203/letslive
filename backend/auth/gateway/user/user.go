package user

import (
	"context"
	"sen1or/lets-live/auth/dto"
	"sen1or/lets-live/auth/gateway"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, *gateway.ErrorResponse)
}
