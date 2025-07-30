package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sen1or/letslive/livestream/dto"
	"sen1or/letslive/livestream/pkg/tracer"
	serviceresponse "sen1or/letslive/livestream/responses"
	"sen1or/letslive/livestream/services/vod"

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

	ctx, span := tracer.MyTracer.Start(ctx, "get_vod_by_id_public_handler.vod_service.get_vod_by_id")
	vod, serviceErr := h.vodService.GetVODById(ctx, vodUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

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

	page, limit := getPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_vods_of_user_public_handler.vod_service.get_public_vods_by_user")
	vods, serviceErr := h.vodService.GetPublicVODsByUser(ctx, userUUID, page, limit)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

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

	page, limit := getPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_vods_of_author_private_handler.vod_service.get_all_vods_by_user")
	livestreams, serviceErr := h.vodService.GetAllVODsByUser(ctx, *userUUID, page, limit)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(livestreams)
}

// TODO: recommendation system
func (h VODHandler) GetRecommendedVODsPublicHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancelCtx := context.WithCancel(r.Context())
	defer cancelCtx()

	page, limit := getPageAndLimitQuery(r)

	ctx, span := tracer.MyTracer.Start(ctx, "get_recommended_vods_public_handler.vod_service.get_recommended_vods")
	vods, serviceErr := h.vodService.GetRecommendedVODs(ctx, page, limit)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

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

	ctx, span := tracer.MyTracer.Start(ctx, "update_vod_metadata_private_handler.vod_service.update_vod_metadata")
	updatedvod, serviceErr := h.vodService.UpdateVODMetadata(ctx, requestBody, streamId, *userId)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

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

	ctx, span := tracer.MyTracer.Start(ctx, "delete_vod_private_handler.vod_service.delete")
	serviceErr := h.vodService.Delete(ctx, vodId, *userUUID)
	if serviceErr != nil {
		h.WriteErrorResponse(w, serviceErr)
		return
	}
	span.End()

	w.WriteHeader(http.StatusNoContent)
}
