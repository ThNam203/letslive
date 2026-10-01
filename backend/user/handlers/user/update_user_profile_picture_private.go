package user

import (
	"context"
	"fmt"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/services"
	"strconv"
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

	file, _, uploadErr := utils.ParseUploadedFile(w, r, "profile-picture")
	if uploadErr != nil {
		h.WriteResponse(w, ctx, uploadErr)
		return
	}
	defer file.Close()

	crop, cropErr := parseAvatarCrop(r)
	if cropErr != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_PAYLOAD,
			nil,
			nil,
			nil,
		))
		return
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_user_profile_picture_private_handler.user_service.update_user_profile_picture")
	savedPath, err := h.userService.UpdateUserProfilePicture(ctx, file, crop, *userUUID)
	span.End()

	if err != nil {
		h.WriteResponse(w, ctx, response.FromError(err))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, &savedPath, nil, nil))
}

// parseAvatarCrop reads the optional crop-x, crop-y and crop-size fields,
// which must be sent together; nil means none were sent. Bounds are checked
// once the image is decoded.
func parseAvatarCrop(r *http.Request) (*services.AvatarCrop, error) {
	names := [3]string{"crop-x", "crop-y", "crop-size"}
	if r.FormValue(names[0]) == "" && r.FormValue(names[1]) == "" && r.FormValue(names[2]) == "" {
		return nil, nil
	}

	var values [3]int
	for i, name := range names {
		v, err := strconv.Atoi(r.FormValue(name))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		values[i] = v
	}

	return &services.AvatarCrop{X: values[0], Y: values[1], Size: values[2]}, nil
}
