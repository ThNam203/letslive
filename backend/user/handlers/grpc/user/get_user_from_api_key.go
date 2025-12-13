package user

import (
	"context"
	"errors"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"

	userv1 "buf.build/gen/go/letslive/letslive-proto/protocolbuffers/go/user"
	"github.com/gofrs/uuid/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h myUserGRPCHandler) GetUserInformation(ctx context.Context, req *userv1.GetUserFromAPIKeyRequest) (*userv1.GetUserFromAPIKeyResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	streamAPIKeyString := req.GetStreamApiKey()
	if len(streamAPIKeyString) == 0 {
		return nil, errors.New(response.RES_ERR_INVALID_INPUT_KEY)
	}

	streamAPIKey, err := uuid.FromString(streamAPIKeyString)
	if err != nil {
		return nil, errors.New(response.RES_ERR_INVALID_INPUT_KEY)
	}

	ctx, span := tracer.MyTracer.Start(ctx, "get_user_by_stream_api_key_internal_handler.user_service.get_user_by_stream_api_key")
	user, sErr := h.userService.GetUserByStreamAPIKey(ctx, streamAPIKey)
	span.End()
	if sErr != nil {
		return nil, errors.New(sErr.Key)
	}

	return &userv1.GetUserFromAPIKeyResponse{
		Id:                user.Id.String(),
		Username:          user.Username,
		Email:             user.Email,
		CreatedAt:         timestamppb.New(user.CreatedAt),
		StreamApiKey:      user.StreamAPIKey.String(),
		DisplayName:       &user.Username,
		PhoneNumber:       user.PhoneNumber,
		Bio:               user.Bio,
		ProfilePicture:    user.ProfilePicture,
		BackgroundPicture: user.BackgroundPicture,
		LivestreamInformation: &userv1.LivestreamInformation{
			UserId:       user.Id.String(),
			Title:        user.LivestreamInformation.Title,
			Description:  user.LivestreamInformation.Description,
			ThumbnailUrl: user.LivestreamInformation.ThumbnailURL,
		},
	}, nil
}
