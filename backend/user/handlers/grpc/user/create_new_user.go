package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/dto"
	"sen1or/letslive/user/pkg/tracer"

	userv1 "buf.build/gen/go/letslive/letslive-proto/protocolbuffers/go/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h myUserGRPCHandler) CreateNewUser(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var authProvider string = string(domains.AuthProviderUnknown)
	if req.GetAuthProvider() == userv1.AuthProvider_AUTH_PROVIDER_GOOGLE {
		authProvider = string(domains.AuthProviderGoogle)
	} else if req.GetAuthProvider() == userv1.AuthProvider_AUTH_PROVIDER_LOCAL {
		authProvider = string(domains.AuthProviderLocal)
	}

	// TODO: remove the use of dto in service layer
	dto := dto.CreateUserRequestDTO{
		Username:     req.GetUsername(),
		Email:        req.GetEmail(),
		AuthProvider: authProvider,
	}

	ctx, span := tracer.MyTracer.Start(ctx, "create_user_internal_handler.user_service.create_new_user")
	createdUser, err := h.userService.CreateNewUser(ctx, dto)
	span.End()

	if err != nil {
		return nil, errors.New(err.Key)
	}

	return &userv1.CreateUserResponse{
		Id:           createdUser.Id.String(),
		Username:     createdUser.Username,
		Email:        createdUser.Email,
		CreatedAt:    timestamppb.New(createdUser.CreatedAt),
		StreamApiKey: createdUser.StreamAPIKey.String(),
		DisplayName:  createdUser.DisplayName,
		PhoneNumber:  createdUser.PhoneNumber,
	}, nil
}
