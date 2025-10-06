package user

import (
	"context"
	serviceresponse "sen1or/letslive/auth/response"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO CreateUserRequestDTO) (*CreateUserResponseDTO, *serviceresponse.Response[any])
}
