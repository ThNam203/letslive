package handlers

import (
	"context"
	"errors"
	"net/http"
	"sen1or/letslive/user/domains"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/response"
	"sen1or/letslive/user/services"
)

type LivestreamInformationHandler struct {
	livestreamService services.LivestreamInformationService
	minioService      services.MinIOService
}

func NewLivestreamInformationHandler(livestreamService services.LivestreamInformationService, minioService services.MinIOService) *LivestreamInformationHandler {
	return &LivestreamInformationHandler{
		livestreamService: livestreamService,
		minioService:      minioService,
	}
}

func (h *LivestreamInformationHandler) UpdatePrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	const maxUploadSize = 11 * 1024 * 1024 // for other information outside of image
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
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
			writeResponse(w, ctx, response.NewResponseFromTemplate[any](
				response.RES_ERR_IMAGE_TOO_LARGE,
				nil,
				nil,
				nil,
			))
			return
		}

		writeResponse(w, ctx, response.NewResponseFromTemplate[any](
			response.RES_ERR_INTERNAL_SERVER,
			nil,
			nil,
			nil,
		))
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	var thumbnailUrl string

	file, fileHeader, formErr := r.FormFile("thumbnail")
	if formErr != nil {
		thumbnailUrl = r.FormValue("thumbnailUrl")
	} else {
		savedPath, err := h.minioService.AddFile(ctx, file, fileHeader, "thumbnails")
		if err != nil {
			writeResponse(w, ctx, response.NewResponseFromTemplate[any](
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
		writeResponse(w, ctx, updateErr)
		return
	}

	writeResponse(w, ctx, response.NewResponseFromTemplate(response.RES_SUCC_OK, updatedData, nil, nil))
}
