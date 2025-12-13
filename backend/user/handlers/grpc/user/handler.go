package user

import (
	"context"
	"sen1or/letslive/user/services"

	"buf.build/gen/go/letslive/letslive-proto/grpc/go/user/userv1grpc"
	userv1 "buf.build/gen/go/letslive/letslive-proto/protocolbuffers/go/user"
)

type UserGRPCHandler interface {
	CreateNewUser(context.Context, *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error)
	// TODO: should be changed to GetUserFromAPIKey
	GetUserInformation(context.Context, *userv1.GetUserFromAPIKeyRequest) (*userv1.GetUserFromAPIKeyResponse, error)
}

type myUserGRPCHandler struct {
	userService services.UserService
	userv1grpc.UnimplementedUserServiceServer
}

func NewUserGRPCHandler(userService services.UserService) UserGRPCHandler {
	return &myUserGRPCHandler{
		userService: userService,
	}
}
