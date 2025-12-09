package user

import (
	"context"
	"sen1or/letslive/auth/gateway/user/dto"
	serviceresponse "sen1or/letslive/auth/response"
)

type UserGateway interface {
	CreateNewUser(ctx context.Context, userRequestDTO dto.CreateUserRequestDTO) (*dto.CreateUserResponseDTO, *serviceresponse.Response[any])
}
