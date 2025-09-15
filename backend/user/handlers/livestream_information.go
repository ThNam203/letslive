package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sen1or/letslive/user/domains"
	servererrors "sen1or/letslive/user/errors"
	"sen1or/letslive/user/pkg/tracer"
	"sen1or/letslive/user/services"
)

type LivestreamInformationHandler struct {
	ErrorHandler
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
		h.WriteErrorResponse(w, servererrors.ErrUnauthorized)
		return
	}
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			h.WriteErrorResponse(w, servererrors.ErrImageTooLarge)
			return
		}

		h.WriteErrorResponse(w, servererrors.ErrInternalServer)
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
			h.WriteErrorResponse(w, servererrors.ErrInternalServer)
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
		h.WriteErrorResponse(w, updateErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedData)
}
