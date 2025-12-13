package user

import (
	"context"
	"errors"
	"net/http"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
)

func (h *UserHandler) UpdateUserBackgroundPicturePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	const maxUploadSize = 10 * 1024 * 1024
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

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_IMAGE_TOO_LARGE,
				nil,
				nil,
				nil,
			))
			return
		}

		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	file, fileHeader, formErr := r.FormFile("background-picture")
	if formErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}
	defer file.Close()

	ctx, span := tracer.MyTracer.Start(ctx, "update_user_background_picture_private_handler.user_service.update_user_background_picture")
	savedPath, err := h.userService.UpdateUserBackgroundPicture(ctx, file, fileHeader, *userUUID)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, err)
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}
