package user

import (
	"context"
	"sen1or/letslive/auth/gateway"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO CreateUserRequestDTO) (*CreateUserResponseDTO, *gateway.ErrorResponse)
}
