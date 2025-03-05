package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sen1or/lets-live/user/domains"
	"sen1or/lets-live/user/repositories"
	"sen1or/lets-live/user/services"
)

type LivestreamInformationHandler struct {
	ErrorHandler
	minioService *services.MinIOService
	ctrl         services.LivestreamInformationController
}

func NewLivestreamInformationHandler(ctrl services.LivestreamInformationController, minioService *services.MinIOService) *LivestreamInformationHandler {
	return &LivestreamInformationHandler{
		ctrl:         ctrl,
		minioService: minioService,
	}
}

func (h *LivestreamInformationHandler) Update(w http.ResponseWriter, r *http.Request) {
	const maxUploadSize = 11 * 1024 * 1024 // for other information outside of image
	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	defer r.Body.Close()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxByteError *http.MaxBytesError
		if errors.As(err, &maxByteError) {
			h.WriteErrorResponse(w, http.StatusRequestEntityTooLarge, err)
			return
		}

		h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("error parsing request body: %s", err.Error()))
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
			h.WriteErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to save the picture: %s", savedPath))
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

	updatedData, err := h.ctrl.Update(updateData)
	if err != nil && errors.Is(err, repositories.ErrRecordNotFound) {
		h.WriteErrorResponse(w, http.StatusNotFound, errors.New("information of user not found"))
		return
	} else if err != nil {
		h.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedData)
}
