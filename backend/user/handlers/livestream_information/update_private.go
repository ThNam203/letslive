package livestream_information

import (
	"context"
	"net/http"
	"sen1or/letslive/shared/pkg/tracer"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/handlers/utils"
	"sen1or/letslive/user/response"
)

func (h *LivestreamInformationHandler) UpdatePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	userUUID, err := utils.GetUserIdFromCookie(r)
	if err != nil {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_UNAUTHORIZED,
			nil,
			nil,
			nil,
		))
		return
	}
	defer r.Body.Close()

	if errRes := utils.ParseUploadForm(w, r, utils.LivestreamThumbnailMaxFileBytes); errRes != nil {
		h.WriteResponse(w, ctx, errRes)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	if len(title) > 50 || len(description) > 500 {
		h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INVALID_INPUT,
			nil,
			nil,
			nil,
		))
		return
	}

	thumbnailUrl := r.FormValue("thumbnailUrl")

	if _, hasThumbnail := r.MultipartForm.File["thumbnail"]; hasThumbnail {
		file, fileHeader, errRes := utils.UploadedFile(r, "thumbnail", utils.LivestreamThumbnailMaxFileBytes)
		if errRes != nil {
			h.WriteResponse(w, ctx, errRes)
			return
		}
		defer file.Close()

		savedPath, err := h.minioService.AddFile(ctx, file, fileHeader, "thumbnails")
		if err != nil {
			h.WriteResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_INTERNAL_SERVER,
				nil,
				nil,
				nil,
			))
			return
		}

		thumbnailUrl = savedPath
	}

	updateData := domains.LivestreamInformation{
		UserID:       *userUUID,
		Title:        &title,
		Description:  &description,
		ThumbnailURL: &thumbnailUrl,
	}

	ctx, span := tracer.MyTracer.Start(ctx, "update_private_handler.livestream_service.update")
	updatedData, updateErr := h.livestreamService.Update(ctx, updateData)
	span.End()

	if updateErr != nil {
		h.WriteResponse(w, ctx, response.FromError(updateErr))
		return
	}

	h.WriteResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, updatedData, nil, nil))
}
