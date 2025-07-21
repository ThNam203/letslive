package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/services/vod"
	"strconv"

	"github.com/gofrs/uuid/v5"
)

type VODHandler struct {
	ResponseHandler
	vodService *vod.VODService
}

func NewVODHandler(vodService *vod.VODService) *VODHandler {
	return &VODHandler{
		vodService: vodService,
	}
}

func (h VODHandler) GetVODByIdPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	streamId := r.PathValue("vodId")
	if len(streamId) == 0 {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPath)
		return
	}

	vodUUID, err := uuid.FromString(streamId)
	if err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidInput)
		return
	}

	vod, serviceErr := h.vodService.GetVODById(ctx, vodUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vod)
}

func (h VODHandler) GetVODsOfUserPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId := r.URL.Query().Get("userId")
	if len(userId) == 0 {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPath)
		return
	}

	userUUID, err := uuid.FromString(userId)
	if err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidInput)
		return
	}

	page := r.URL.Query().Get("page")
	pageNum, pageErr := strconv.Atoi(page)
	if pageErr != nil || pageNum < 0 {
		h.WriteErrorResponse(w, serviceresponse.ErrMissingPageParameter)
		return
	}

	limit := r.URL.Query().Get("limit")
	limitNum, limitErr := strconv.Atoi(limit)
	if limitErr != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrMissingLimitParameter)
		return
	}

	vods, serviceErr := h.vodService.GetPublicVODsByUser(ctx, userUUID, pageNum, limitNum)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vods)
}

func (h VODHandler) GetVODsOfAuthorPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userUUID, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidInput)
		return
	}

	page := r.URL.Query().Get("page")
	pageNum, pageErr := strconv.Atoi(page)
	if pageErr != nil || pageNum < 0 {
		h.WriteErrorResponse(w, serviceresponse.ErrMissingPageParameter)
		return
	}

	limit := r.URL.Query().Get("limit")
	limitNum, limitErr := strconv.Atoi(limit)
	if limitErr != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrMissingLimitParameter)
		return
	}

	livestreams, serviceErr := h.vodService.GetAllVODsByUser(ctx, *userUUID, pageNum, limitNum)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

// TODO: recommendation system
func (h VODHandler) GetRecommendedVODsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	page := r.URL.Query().Get("page")
	pageNum, pageErr := strconv.Atoi(page)
	if pageErr != nil || pageNum < 0 {
		h.WriteErrorResponse(w, serviceresponse.ErrMissingPageParameter)
		return
	}

	limit := r.URL.Query().Get("limit")
	limitNum, limitErr := strconv.Atoi(limit)
	if limitErr != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrMissingLimitParameter)
		return
	}

	vods, serviceErr := h.vodService.GetRecommendedVODs(ctx, pageNum, limitNum)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vods)
}

func (h VODHandler) UpdateVODMetadataPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	userId, err := getUserIdFromCookie(r)
	if err != nil {
		h.WriteErrorResponse(w, err)
		return
	}

	rawStreamId := r.PathValue("vodId")
	streamId, er := uuid.FromString(rawStreamId)
	if er != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidInput)
		return
	}
	defer r.Body.Close()

	var requestBody dto.UpdateVODRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPayload)
		return
	}

	updatedvod, serviceErr := h.vodService.UpdateVODMetadata(ctx, requestBody, streamId, *userId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedvod)
}

func (h VODHandler) DeleteVODPrivateHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	rawVODId := r.PathValue("vodId")
	vodId, err := uuid.FromString(rawVODId)

	if err != nil {
		h.WriteErrorResponse(w, serviceresponse.ErrInvalidPath)
		return
	}

	userUUID, cErr := getUserIdFromCookie(r)
	if cErr != nil {
		h.WriteErrorResponse(w, cErr)
		return
	}

	serviceErr := h.vodService.Delete(ctx, vodId, *userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
