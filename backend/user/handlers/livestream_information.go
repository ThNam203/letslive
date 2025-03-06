package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sen1or/lets-live/user/domains"
	servererrors "sen1or/lets-live/user/errors"
	"sen1or/lets-live/user/services"
)

type LivestreamInformationHandler struct {
	ErrorHandler
	ctrl         services.LivestreamInformationService
	minioService services.MinIOService
}

func NewLivestreamInformationHandler(ctrl services.LivestreamInformationService, minioService services.MinIOService) *LivestreamInformationHandler {
	return &LivestreamInformationHandler{
		ctrl:         ctrl,
		minioService: minioService,
	}
}

func (h *LivestreamInformationHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		thumbnailUrl = r.FormValue("thumbnailUrl")
	} else {
		savedPath, err := h.minioService.AddFile(file, fileHeader, "thumbnails")
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

	updatedData, updateErr := h.ctrl.Update(updateData)
	if updateErr != nil {
		h.WriteErrorResponse(w, updateErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedData)
}
