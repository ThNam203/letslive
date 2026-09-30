package user

import (
	"context"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"
)

func (h *UserHandler) UpdateUserProfilePicturePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userUUID, cookieErr := utils.GetUserIdFromCookie(r)
	if cookieErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	file, fileHeader, uploadErr := utils.ParseUploadedFile(w, r, "profile-picture")
	if uploadErr != nil {
		h.WriteResponse(w, ctx, uploadErr)
		return
	}
	defer file.Close()

	ctx, span := tracer.MyTracer.Start(ctx, "update_user_profile_picture_private_handler.user_service.update_user_profile_picture")
	savedPath, err := h.userService.UpdateUserProfilePicture(ctx, file, fileHeader, *userUUID)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}
